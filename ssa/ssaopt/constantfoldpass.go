package ssaopt

import "github.com/MovingtoMars/nnvm/ssa"

type ConstantFoldPass struct {
}

func NewConstantFoldPass() *ConstantFoldPass {
	return &ConstantFoldPass{}
}

func (_ ConstantFoldPass) String() string {
	return "constant fold"
}

func (v ConstantFoldPass) Run(mod *ssa.Module) {

}
