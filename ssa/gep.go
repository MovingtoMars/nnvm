package ssa

import "github.com/MovingtoMars/nnvm/types"

// Note: for int literals accessing structs, the int type will be ignored, ie. no overflow/underflow will occur due to the int type.
type GEP struct {
	NameHandler
	ReferenceHandler
	BlockHandler

	value   Value
	indexes []Value
}

func newGEP(value Value, indexes []Value) *GEP {
	return &GEP{
		value:   value,
		indexes: indexes,
	}
}

func (v GEP) String() string {
	return "gep " + ValueString(v.value) + ", " + valueListString(v.indexes)
}

func (v *GEP) operands() []*Value {
	ops := []*Value{&v.value}

	for i := 0; i < len(v.indexes); i++ {
		ops = append(ops, &v.indexes[i])
	}

	return ops
}

func (v *GEP) Type() types.Type {
	typ := v.value.Type()

	for i, index := range v.indexes {
		switch styp := typ.(type) {
		case *types.Pointer:
			if i != 0 {
				return types.NewVoid() // only the first index can dereference a pointer
			}

			typ = styp.Element()

		case *types.Array:
			typ = styp.Element()

		case *types.Struct:
			lit, ok := index.(*IntLiteral)
			if !ok {
				return types.NewVoid()
			}

			if int(lit.value) >= len(styp.Fields()) {
				return types.NewVoid()
			}

			typ = styp.Fields()[lit.value]

		default:
			return types.NewVoid()
		}
	}

	return types.NewPointer(typ)
}

func (_ GEP) IsTerminating() bool {
	return false
}
