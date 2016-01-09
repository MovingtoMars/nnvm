package ssa

type Br struct {
	BlockHandler

	target Value // must be a block
}

func newBr(target *Block) *Br {
	return &Br{
		target: target,
	}
}

func (v Br) String() string {
	return "br " + ValueString(v.target)
}

func (v *Br) operands() []*Value {
	return []*Value{&v.target}
}

func (_ Br) IsTerminating() bool {
	return true
}
