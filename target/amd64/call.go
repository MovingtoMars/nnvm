package amd64

import (
	"github.com/MovingtoMars/nnvm/ssa"
	"github.com/MovingtoMars/nnvm/target/platform"
	"github.com/MovingtoMars/nnvm/types"
)

func (v Target) genSaveFunctionParameters(a *allocator, fn *ssa.Function) {
	var vals []ssa.Value
	for _, par := range fn.Parameters() {
		vals = append(vals, par)
	}

	if v.Platform.IsUnixLike() {
		v.sysVCopyFunctionVals(a, vals, fn.Type().(*types.Signature), false)
	} else if v.Platform == platform.Windows {
		v.winSaveFunctionParameters(a, vals, fn.Type().(*types.Signature))
	} else {
		panic("unim")
	}
}

func (v Target) genLoadCallArguments(a *allocator, call *ssa.Call) {
	ops := ssa.GetOperands(call)
	if v.Platform.IsUnixLike() {
		v.sysVCopyFunctionVals(a, ops[1:], ops[0].Type().(*types.Signature), true)
	} else if v.Platform == platform.Windows {
		v.winLoadCallArguments(a, ops[1:], ops[0].Type().(*types.Signature))
	} else {
		panic("unim")
	}
}

func (v Target) genSaveReturnValue(a *allocator, val ssa.Value) {
	if v.Platform.IsUnixLike() {
		v.sysVCopyReturnValue(a, val, true)
	} else if v.Platform == platform.Windows {
		v.winCopyReturnValue(a, val, true)
	} else {
		panic("unim")
	}
}

func (v Target) genLoadReturnValue(a *allocator, val ssa.Value) {
	if v.Platform.IsUnixLike() {
		v.sysVCopyReturnValue(a, val, false)
	} else if v.Platform == platform.Windows {
		v.winCopyReturnValue(a, val, false)
	} else {
		panic("unim")
	}
}
