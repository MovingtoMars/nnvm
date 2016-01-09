package ssa

import "github.com/MovingtoMars/nnvm/types"

type Call struct {
	NameHandler
	ReferenceHandler
	BlockHandler

	function  Value
	arguments []Value
}

func newCall(target *Function, args []Value) *Call {
	return &Call{
		function:  target,
		arguments: args,
	}
}

func (v Call) Type() types.Type {
	return v.function.(*Function).typ.ReturnType()
}

func (v *Call) operands() []*Value {
	ops := []*Value{&v.function}

	for i := 0; i < len(v.arguments); i++ {
		ops = append(ops, &v.arguments[i])
	}

	return ops
}

func (v Call) String() string {
	return "call " + v.Type().String() + " @" + v.function.Name() + "(" + valueListString(v.arguments) + ")"
}

func (_ Call) IsTerminating() bool {
	return false
}
