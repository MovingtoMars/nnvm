package ssa

type CondBr struct {
	BlockHandler

	condition               Value // must be i1
	trueTarget, falseTarget Value // must be blocks
}

func newCondBr(condition Value, trueTarget, falseTarget *Block) *CondBr {
	return &CondBr{
		condition:   condition,
		trueTarget:  trueTarget,
		falseTarget: falseTarget,
	}
}

func (v CondBr) String() string {
	return "condbr " + ValueString(v.condition) + ", " + ValueString(v.trueTarget) + ", " + ValueString(v.falseTarget)
}

func (v *CondBr) operands() []*Value {
	return []*Value{&v.condition, &v.trueTarget, &v.falseTarget}
}

func (_ CondBr) IsTerminating() bool {
	return true
}
