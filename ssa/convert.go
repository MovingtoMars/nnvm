package ssa

import (
	"strings"

	"github.com/MovingtoMars/nnvm/types"
)

//go:generate stringer -type=ConvertType
type ConvertType int

const (
	// Integer
	ConvertSExt ConvertType = iota
	ConvertZExt
	ConvertTrunc

	// Pointer
	ConvertBitcast

	// Float
	ConvertFExt
	ConvertFTrunc

	// Float-Integer
	ConvertFToUI
	ConvertFToSI
	ConvertUIToF
	ConvertSIToF

	// Pointer-Integer
	ConvertPtrToInt
	ConvertIntToPtr
)

type Convert struct {
	NameHandler
	BlockHandler
	ReferenceHandler

	value       Value
	target      types.Type
	convertType ConvertType
}

func newConvert(value Value, target types.Type, convertType ConvertType) *Convert {
	return &Convert{
		value:       value,
		target:      target,
		convertType: convertType,
	}
}

func (v Convert) ConvertType() ConvertType {
	return v.convertType
}

func (v Convert) String() string {
	return strings.ToLower(v.convertType.String()[7:]) + " " + ValueString(v.value) + " to " + v.target.String()
}

func (v Convert) Type() types.Type {
	return v.target
}

func (v *Convert) operands() []*Value {
	return []*Value{&v.value}
}

func (_ Convert) IsTerminating() bool {
	return false
}
