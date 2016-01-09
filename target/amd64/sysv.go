package amd64

import (
	"fmt"
	"reflect"

	"github.com/MovingtoMars/nnvm/ssa"
	"github.com/MovingtoMars/nnvm/types"
)

// see the System V ABI
type sysVParameterClass int

const (
	sysVClassINTEGER sysVParameterClass = iota
	sysVClassSSE
	sysVClassSSEUP
	sysVClassX87
	sysVClassX87UP
	sysVClassCOMPLEX_X87
	sysVClassNO_CLASS
	sysVClassMEMORY
)

func sysVClassifyType(typ types.Type) []sysVParameterClass {
	switch typ := typ.(type) {
	case *types.Struct:
		return sysVClassifyStructType(typ)

	case *types.Array:
		return sysVClassifyArrayType(typ)

	case *types.Int:
		switch typ.Width() {
		case 1, 8, 16, 32, 64, 128:
			return []sysVParameterClass{sysVClassINTEGER}
		default:
			panic("unim")
		}

	case *types.Pointer:
		return []sysVParameterClass{sysVClassINTEGER}

	case *types.Float:
		switch typ.Type() {
		case types.Float32, types.Float64:
			return []sysVParameterClass{sysVClassSSE}
		default:
			panic("unim")
		}

	case types.Void:
		return []sysVParameterClass{sysVClassNO_CLASS}

	default:
		panic("amd64.classifyType: unim: " + reflect.TypeOf(typ).String())
	}
}

func sysVRecursiveClassifyFields(c1 sysVParameterClass, c2 []sysVParameterClass) sysVParameterClass {
	if len(c2) == 0 {
		return c1
	} else if len(c2) == 1 {
		c2 := c2[0]

		switch {
		case c1 == c2:
			return c1
		case c1 == sysVClassNO_CLASS:
			return c2
		case c2 == sysVClassNO_CLASS:
			return c1
		case c1 == sysVClassMEMORY || c2 == sysVClassMEMORY:
			return sysVClassMEMORY
		case c1 == sysVClassINTEGER || c2 == sysVClassINTEGER:
			return sysVClassINTEGER
		case c1 == sysVClassX87 || c1 == sysVClassX87UP || c1 == sysVClassCOMPLEX_X87 ||
			c2 == sysVClassX87 || c2 == sysVClassX87UP || c2 == sysVClassCOMPLEX_X87:
			return sysVClassMEMORY
		default:
			return sysVClassSSE
		}
	} else {
		return sysVRecursiveClassifyFields(sysVRecursiveClassifyFields(c1, c2[0:1]), c2[1:])
	}
}

func sysVClassifyStructType(typ *types.Struct) []sysVParameterClass {
	layout := newStructLayout(typ)
	if layout.hasUnalignedFields() || layout.size > 4*64 {
		return []sysVParameterClass{sysVClassMEMORY}
	}

	eightbytes := sysVGetStructTypeEightbytes(typ, layout)
	_ = eightbytes

	panic("amd64.classifyStructType: unim")

	/*eightbyteClasses := make([]parameterClass, len(eightbytes))

	for i, eightbyte := range eightbytes {
		var fieldClasses []parameterClass
		for _, field := range eightbyte {
			fieldClasses = append(fieldClasses, classifyType(field))
		}

		if len(fieldClasses) == 1 {
			eightbyteClasses[i] = fieldClasses[0]
		}
	}

	return nil*/
}

func sysVGetStructTypeEightbytes(typ *types.Struct, layout structLayout) [][]types.Type {
	// an array of eightbyte field arrays
	var eightbyteFields [][]types.Type
	eightbyteIndex := 0

	if len(layout.fields) > 0 {
		eightbyteFields = append(eightbyteFields, nil)
	}

	bits := 0
	for i, field := range layout.fields {
		bits += TypeSizeInBits(field.field)
		eightbyteFields[eightbyteIndex] = append(eightbyteFields[eightbyteIndex], field.field)

		for bits > 64 {
			eightbyteIndex++
			eightbyteFields = append(eightbyteFields, nil)
			eightbyteFields[eightbyteIndex] = append(eightbyteFields[eightbyteIndex], field.field)
			bits -= 64
		}

		bits += field.paddingBits

		if i < len(layout.fields)-1 {
			for bits >= 64 {
				eightbyteIndex++
				eightbyteFields = append(eightbyteFields, nil)
				bits -= 64
			}
		}
	}

	for _, eightbyte := range eightbyteFields {
		fmt.Println(eightbyte)
	}

	return eightbyteFields
}

func sysVClassifyArrayType(typ *types.Array) []sysVParameterClass {
	sz := TypeSizeInBits(typ) // TODO should be TypeStoreSizeInBits?
	if sz > 4*64 {
		return []sysVParameterClass{sysVClassMEMORY}
	}

	panic("unim")
}

var sysVParameterRegSeq = []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}

// warning: this function sucks
//
// If reverse is false, move parameters to values
// If reverse is true, move values to call arguments
func (v Target) sysVCopyFunctionVals(a *allocator, vals []ssa.Value, sig *types.Signature, reverse bool) {
	regSeqIndex := 0
	nextReg := func() string {
		if regSeqIndex >= len(sysVParameterRegSeq) {
			return ""
		}
		regSeqIndex++
		return sysVParameterRegSeq[regSeqIndex-1]
	}

	if sysVClassifyType(sig.ReturnType())[0] == sysVClassMEMORY {
		regSeqIndex++
	}

	curStackEightbyteIndex := 0

	if reverse {
		v.wop("andq $-16, #rsp")
	}

	for _, val := range vals {
		classList := sysVClassifyType(val.Type())

		switch val.Type().(type) {
		case *types.Struct, *types.Array:
			panic("unim")
		}

		if len(classList) != 1 {
			panic("internal error: non-aggregate type has ABI classList of length > 1")
		}

		class := classList[0]

		switch class {
		case sysVClassINTEGER:
			storesz := TypeStoreSizeInBits(val.Type())
			if storesz > 64 {
				panic("unim")
			}
			if !isPow2(storesz) {
				panic("unim")
			}

			if reg := nextReg(); reg != "" {
				if !reverse {
					v.wop("mov%s #%s, %s", sizeSuffixBits(storesz), regToSize(reg, storesz), a.valStr(val))
				} else {
					if storesz != 64 {
						v.wop("xorq #%s, #%s", reg, reg)
					}
					v.wop("mov%s %s, #%s", sizeSuffixBits(storesz), a.valStr(val), regToSize(reg, storesz))
				}
			} else {
				if !reverse {
					v.wop("movq $%d, #r11", curStackEightbyteIndex)
					v.wop("movq 16(#rbp, #r11, 8), #r11")
					v.wop("mov%s #%s, %s", sizeSuffixBits(storesz), regToSize("r11", storesz), a.valStr(val))
				} else {
					panic("fuck")
				}

				curStackEightbyteIndex++
			}
		default:
			panic("unim")
		}
	}
}

func (v Target) sysVCopyReturnValue(a *allocator, val ssa.Value, saving bool) {
	if val == nil {
		return
	}
	if _, ok := val.Type().(types.Void); ok {
		return
	}

	classList := sysVClassifyType(val.Type())

	switch val.Type().(type) {
	case *types.Struct, *types.Array:
		panic("unim")
	}

	if len(classList) != 1 {
		panic("internal error: non-aggregate type has ABI classList of length > 1")
	}

	class := classList[0]

	switch class {
	case sysVClassINTEGER:
		storesz := TypeStoreSizeInBits(val.Type())
		if saving {
			v.wop("mov%s #%s, %s", sizeSuffixBits(storesz), regToSize("rax", storesz), a.valStr(val))
		} else {
			v.moveIntToReg(a, val, "rax")
		}

	default:
		panic("unim")
	}
}
