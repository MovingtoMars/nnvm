package ssa

// Unreachable is an instruction used as a terminating instruction for a block that control should never reach the end of.
// It is used to satisfy the validator.
// Implementation is undefined.
type Unreachable struct {
	BlockHandler
}

func newUnreachable() *Unreachable {
	return &Unreachable{}
}

func (_ Unreachable) String() string {
	return "unreachable"
}

func (_ Unreachable) IsTerminating() bool {
	return true
}

func (_ Unreachable) operands() []*Value {
	return nil
}
