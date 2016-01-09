package types

// Label represents the type of a block.
type Label struct{}

func NewLabel() *Label {
	return &Label{}
}

func (_ Label) String() string {
	return "label"
}

func (v Label) Equals(t Type) bool {
	_, ok := t.(*Label)
	return ok
}
