package ssa

import (
	"strings"

	"github.com/MovingtoMars/nnvm/types"
)

//go:generate stringer -type=BinOpType
type BinOpType int

const (
	// Integer arithmetic
	BinOpAdd BinOpType = iota
	BinOpSub
	BinOpMul
	BinOpSDiv
	BinOpUDiv
	BinOpSRem
	BinOpURem

	// Float arithmetic
	BinOpFAdd
	BinOpFSub
	BinOpFMul
	BinOpFDiv
	BinOpFRem

	// Bitwise
	BinOpShl
	BinOpLShr
	BinOpAShr
	BinOpAnd
	BinOpOr
	BinOpXor
)

type BinOp struct {
	BlockHandler
	NameHandler
	ReferenceHandler

	binOpType BinOpType
	x, y      Value
}

func newBinOp(x, y Value, binOpType BinOpType) *BinOp {
	return &BinOp{
		binOpType: binOpType,
		x:         x,
		y:         y,
	}
}

func (v BinOp) BinOpType() BinOpType {
	return v.binOpType
}

func (v *BinOp) operands() []*Value {
	return []*Value{&v.x, &v.y}
}

func (v BinOp) String() string {
	return strings.ToLower(v.binOpType.String()[5:]) + " " + ValueString(v.x) + ", " + ValueString(v.y)
}

func (v BinOp) Type() types.Type {
	return v.x.Type()
}

func (_ BinOp) IsTerminating() bool { return false }
