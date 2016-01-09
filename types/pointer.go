package types

// Pointer represents the type of a pointer to another type.
type Pointer struct {
	element Type
}

// element cannot be an instance of Void
func NewPointer(element Type) *Pointer {
	if element == nil {
		panic("types.NewPointer: element cannot be nil")
	} else if _, ok := element.(Void); ok {
		panic("types.NewPointer: element cannot be void")
	}

	return &Pointer{
		element: element,
	}
}

func (v Pointer) String() string {
	return "*" + v.element.String()
}

func (v Pointer) Element() Type {
	return v.element
}

func (v Pointer) Equals(t Type) bool {
	ptr, ok := t.(*Pointer)
	if !ok {
		return false
	}

	return v.element.Equals(ptr.element)
}
