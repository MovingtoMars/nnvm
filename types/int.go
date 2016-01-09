package types

import "fmt"

const (
	MaxIntWidth    = (1 << 31) - 1
	MaxArrayLength = MaxIntWidth
)

// Int represents an arbitrary-width integer. Int holds no sign information. Note that not all backends support all widths.
type Int struct {
	width int
}

// Int returns a new Int with the specified width. Panics if the width is greater than MaxIntSize.
func NewInt(width int) *Int {
	if width > MaxIntWidth {
		panic("types.NewInt: width > MaxIntWidth")
	} else if width < 0 {
		panic("types.NewInt: width < 0")
	}

	return &Int{
		width: width,
	}
}

func (v Int) String() string {
	return fmt.Sprintf("i%d", v.width)
}

func (v Int) Equals(t Type) bool {
	i, ok := t.(*Int)
	if !ok {
		return false
	}

	return v.width == i.width
}

func (v Int) Width() int {
	return v.width
}
