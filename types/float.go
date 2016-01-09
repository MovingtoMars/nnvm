package types

type FloatType int

const (
	// IEEE 754
	Float32 FloatType = iota
	Float64
)

func (v FloatType) CanExtendTo(w FloatType) bool {
	switch v {
	case Float32:
		return true
	case Float64:
		return false
	default:
		panic("unim")
	}
}

func (v FloatType) CanTruncateTo(w FloatType) bool {
	switch v {
	case Float32:
		return false
	case Float64:
		return true
	default:
		panic("unim")
	}
}

// Width returns the width of the float type in bits.
func (v FloatType) Width() int {
	switch v {
	case Float32:
		return 32
	case Float64:
		return 64
	default:
		panic("unim")
	}
}

type Float struct {
	width FloatType
}

func NewFloat(width FloatType) *Float {
	return &Float{width}
}

func (v Float) String() string {
	switch v.width {
	case Float32:
		return "f32"
	case Float64:
		return "f64"
	default:
		panic("FloatType.String: invalid FloatWidth")
	}
}

func (v Float) Type() FloatType {
	return v.width
}

func (v Float) Equals(t Type) bool {
	f, ok := t.(*Float)
	return ok && f.width == v.width
}
