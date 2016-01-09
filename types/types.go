package types

type Type interface {
	String() string
	Equals(Type) bool
}

type TypeList []Type

func (v TypeList) Equals(t TypeList) bool {
	if len(v) != len(t) {
		return false
	}

	for i, typ := range v {
		if !typ.Equals(t[i]) {
			return false
		}
	}

	return true
}

func IsFirstClass(t Type) bool {
	switch t.(type) {
	case Void, *Signature:
		return false
	}
	return true
}
