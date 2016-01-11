package amd64

import (
	"fmt"

	"github.com/MovingtoMars/nnvm/types"
)

func TypeStoreSizeInBits(typ types.Type) int {
	return ((TypeSizeInBits(typ) + 7) / 8) * 8
}

func TypeSizeInBits(typ types.Type) int {
	switch typ := typ.(type) {
	case *types.Int:
		w := typ.Width()
		for w%8 != 0 {
			w++
		}
		return w

	case *types.Float:
		return typ.Type().Width()

	case *types.Pointer:
		return 64

	case *types.Array:
		return TypeSizeInBits(typ.Element()) * typ.Length()

	case *types.Struct:
		return newStructLayout(typ).size

	default:
		panic("unim")
	}
}

type structField struct {
	field       types.Type
	paddingBits int // padding bits after field
}

func (v structField) String() string {
	return fmt.Sprintf("%s (%d bits) + %d bits", v.field, TypeSizeInBits(v.field), v.paddingBits)
}

type structLayout struct {
	fields    []structField
	size      int
	alignment int
}

func newStructLayout(typ *types.Struct) structLayout {
	if typ.Packed() {
		panic("packes structs unimplemented")
	}

	layout := structLayout{}
	sz := 0
	maxAlign := 0

	for i, field := range typ.Fields() {
		fieldsz := TypeStoreSizeInBits(field)
		fieldAlign := TypeAlignmentInBits(field)

		if fieldAlign > maxAlign {
			maxAlign = fieldAlign
		}

		for sz%fieldAlign != 0 {
			sz++
			layout.fields[i-1].paddingBits++
		}

		sz += fieldsz

		layout.fields = append(layout.fields, structField{field: field})
	}

	for sz%maxAlign != 0 {
		sz++
		layout.fields[len(layout.fields)-1].paddingBits++
	}

	layout.size = sz
	layout.alignment = maxAlign

	return layout
}

func (v structLayout) hasUnalignedFields() bool {
	// TODO: packed
	return false
}

func (v structLayout) fieldOffsetBits(index int) int {
	bits := 0

	for _, field := range v.fields[:index] {
		bits += field.paddingBits + TypeStoreSizeInBits(field.field)
	}

	return bits
}

func (v structLayout) String() string {
	str := "struct {\n"

	for _, field := range v.fields {
		str += "    " + field.String() + "\n"
	}

	return str + "}"
}

func PrintStructLayout(struc *types.Struct) {
	fmt.Println(newStructLayout(struc))
}

// will always be multiple of 8
func TypeAlignmentInBits(typ types.Type) int {
	switch typ := typ.(type) {
	case *types.Struct:
		return newStructLayout(typ).alignment
	case *types.Array:
		return TypeAlignmentInBits(typ.Element())
	}

	sz := TypeSizeInBits(typ)

	i := 8
	for i < sz {
		i *= 2
	}

	return i
}
