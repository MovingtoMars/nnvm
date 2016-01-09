package ssa

type Ret struct {
	BlockHandler

	returnValue Value // nil for void return
}

func newRet(val Value) *Ret {
	return &Ret{
		returnValue: val,
	}
}

func (v *Ret) operands() []*Value {
	return []*Value{&v.returnValue}
}

func (v Ret) String() string {
	if v.returnValue == nil {
		return "ret"
	}
	return "ret " + ValueString(v.returnValue)
}

func (_ Ret) IsTerminating() bool { return true }
