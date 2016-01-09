package types

import "fmt"

// Array represents the type of an aggregate of a type, arranges sequentially in memory.
type Array struct {
	element Type
	length  int
}

func NewArray(element Type, length int) *Array {
	if element == nil {
		panic("NewArrayType: element cannot be nil")
	} else if _, ok := element.(Void); ok {
		panic("NewArrayType: element cannot be void")
	}

	if length > MaxArrayLength {
		panic("NewArrayType: length > MaxArrayLength")
	} else if length < 0 {
		panic("NewArrayType: length < 0")
	}

	return &Array{
		element: element,
		length:  length,
	}
}

func (v Array) String() string {
	return fmt.Sprintf("[%d]%s", v.length, v.element)
}

func (v Array) Equals(t Type) bool {
	arr, ok := t.(*Array)
	if !ok {
		return false
	}

	return v.length == arr.length && v.element.Equals(arr.element)
}

func (v Array) Element() Type {
	return v.element
}

func (v Array) Length() int {
	return v.length
}
