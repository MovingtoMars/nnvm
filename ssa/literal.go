package ssa

import (
	"fmt"
	"math"

	"github.com/MovingtoMars/nnvm/types"
)

type Literal interface {
	Value
	LiteralValue() interface{}
}

type IntLiteral struct {
	ReferenceHandler

	typ   *types.Int
	value uint64
}

func NewIntLiteral(value uint64, typ *types.Int) *IntLiteral {
	if typ.Width() < 64 {
		value &= ((1 << uint(typ.Width())) - 1)
	}

	return &IntLiteral{
		typ:   typ,
		value: value,
	}
}

func (v IntLiteral) Type() types.Type {
	return v.typ
}

func (v IntLiteral) Name() string {
	return fmt.Sprintf("%d", v.value)
}

func (_ IntLiteral) SetName(_ string) {}

func (v IntLiteral) LiteralValue() interface{} {
	return v.value
}

type FloatLiteral struct {
	ReferenceHandler

	typ   *types.Float
	value uint64
}

func NewFloat64Literal(value float64) *FloatLiteral {
	return &FloatLiteral{
		typ:   types.NewFloat(types.Float64),
		value: math.Float64bits(value),
	}
}

func NewFloat32Literal(value float32) *FloatLiteral {
	return &FloatLiteral{
		typ:   types.NewFloat(types.Float32),
		value: uint64(math.Float32bits(value)),
	}
}

func (v FloatLiteral) Type() types.Type {
	return v.typ
}

func (v FloatLiteral) Name() string {
	var length int

	switch v.typ.Type() {
	case types.Float64:
		length = 16
	case types.Float32:
		length = 8
	default:
		panic("FloatLiteral.Name: invalid FloatWidth")
	}

	return fmt.Sprintf("0x%0*X", length, v.value)
}

// Use for display purposes only, can be inaccurate!
func (v FloatLiteral) Float64() float64 {
	switch v.typ.Type() {
	case types.Float64:
		return math.Float64frombits(v.value)
	case types.Float32:
		return float64(math.Float32frombits(uint32(v.value)))
	default:
		panic("FloatLiteral.Name: invalid FloatWidth")
	}
}

func (_ FloatLiteral) SetName(_ string) {}

func (v FloatLiteral) LiteralValue() interface{} {
	return v.value
}

type StringLiteral struct {
	ReferenceHandler

	value string
}

func NewStringLiteral(value string, appendNullByte bool) *StringLiteral {
	if appendNullByte {
		value += string(0)
	}

	return &StringLiteral{
		value: value,
	}
}

func (v StringLiteral) Type() types.Type {
	return types.NewArray(types.NewInt(8), len(v.value))
}

func (v StringLiteral) LiteralValue() interface{} {
	return v.value
}

func (v StringLiteral) Name() string {
	return "\"" + EscapeString(v.value) + "\""
}

func (_ StringLiteral) SetName(string) {}
