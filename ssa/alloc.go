package ssa

import "github.com/MovingtoMars/nnvm/types"

type Alloc struct {
	NameHandler
	ReferenceHandler
	BlockHandler

	typ types.Type
}

func newAlloc(typ types.Type) *Alloc {
	return &Alloc{
		typ: typ,
	}
}

func (v Alloc) String() string {
	return "alloc " + v.typ.String()
}

func (v Alloc) Type() types.Type {
	return types.NewPointer(v.typ)
}

func (_ Alloc) operands() []*Value {
	return nil
}

func (_ Alloc) IsTerminating() bool {
	return false
}
