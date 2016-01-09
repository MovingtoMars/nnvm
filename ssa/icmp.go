package ssa

import (
	"strings"

	"github.com/MovingtoMars/nnvm/types"
)

//go:generate stringer -type=IntPredicate
type IntPredicate int

const (
	IntEQ  IntPredicate = iota // equal
	IntNEQ                     // not equal
	IntUGT                     // unsigned greater than
	IntUGE                     // unsigned greater or equal
	IntULT                     // unsigned less than
	IntULE                     // unsigned less or equal
	IntSGT                     // signed greater than
	IntSGE                     // signed greater or equal
	IntSLT                     // signed less than
	IntSLE                     // signed less or equal

)

type ICmp struct {
	BlockHandler
	NameHandler
	ReferenceHandler

	predicate IntPredicate
	x, y      Value
}

func newICmp(x, y Value, predicate IntPredicate) *ICmp {
	return &ICmp{
		predicate: predicate,
		x:         x,
		y:         y,
	}
}

func (v ICmp) Predicate() IntPredicate {
	return v.predicate
}

func (v *ICmp) operands() []*Value {
	return []*Value{&v.x, &v.y}
}

func (v ICmp) String() string {
	return "icmp " + strings.ToLower(v.predicate.String()[3:]) + " " + ValueString(v.x) + ", " + ValueString(v.y)
}

func (_ ICmp) Type() types.Type {
	return types.NewInt(1)
}

func (_ ICmp) IsTerminating() bool {
	return false
}
