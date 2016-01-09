package types

// Void represents the type associated with no value.
type Void struct{}

func NewVoid() Void {
	return Void{}
}

func (_ Void) String() string {
	return "void"
}

func (v Void) Equals(t Type) bool {
	_, ok := t.(Void)
	return ok
}
