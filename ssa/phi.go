package ssa

import "github.com/MovingtoMars/nnvm/types"

type Phi struct {
	ReferenceHandler
	BlockHandler
	NameHandler

	typ            types.Type
	incomingValues []Value
	incomingBlocks []Value
}

func newPhi(t types.Type) *Phi {
	return &Phi{
		typ: t,
	}
}

func (v *Phi) AddIncoming(val Value, block *Block) {
	v.incomingValues = append(v.incomingValues, val)
	v.incomingBlocks = append(v.incomingBlocks, block)

	val.addReference(v)
	block.addReference(v)
}

func (v Phi) GetIncoming(index int) (Value, *Block) {
	if index >= len(v.incomingValues) {
		panic("Phi.GetIncoming: index out of range")
	}

	return v.incomingValues[index], v.incomingBlocks[index].(*Block)
}

func (v Phi) NumIncoming() int {
	return len(v.incomingValues)
}

func (v *Phi) RemoveIncoming(index int) {
	if index >= len(v.incomingValues) {
		panic("Phi.RemoveIncoming: index out of range")
	}

	slices := []*[]Value{
		&v.incomingValues,
		&v.incomingBlocks,
	}

	for _, slice := range slices {
		(*slice)[index].removeReference(v)
		copy((*slice)[index:], (*slice)[index+1:])
		(*slice) = (*slice)[:len(*slice)-1]
	}
}

func (v Phi) Type() types.Type {
	return v.typ
}

func (v *Phi) operands() []*Value {
	var ops []*Value
	for i, val := range v.incomingValues {
		ops = append(ops, &val, &v.incomingBlocks[i])
	}

	return ops
}

func (v Phi) String() string {
	str := "phi " + v.typ.String() + " "

	for i, val := range v.incomingValues {
		str += "[ " + ValueIdentifier(val) + ", " + ValueIdentifier(v.incomingBlocks[i]) + " ]"

		if i < len(v.incomingValues)-1 {
			str += ", "
		}
	}

	return str
}

func (_ Phi) IsTerminating() bool {
	return false
}
