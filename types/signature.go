package types

// Signature represents a function signature.
type Signature struct {
	parameters TypeList
	returnType Type
	variadic   bool
}

func (v Signature) Parameters() TypeList {
	return v.parameters
}

func (v Signature) ReturnType() Type {
	return v.returnType
}

func (v Signature) Variadic() bool {
	return v.variadic
}

// NewSignature returns a signature with the specified parameter and return types. Panics if and supplied types are nil.
func NewSignature(parameters []Type, returnType Type, variadic bool) *Signature {
	if returnType == nil {
		panic("NewSignature: cannot have nil return type")
	}

	for _, par := range parameters {
		if par == nil {
			panic("NewSignature: cannot have nil parameter type")
		} else if _, ok := par.(Void); ok {
			panic("NewSignature: cannot have void parameter type")
		}
	}

	return &Signature{
		parameters: parameters,
		returnType: returnType,
		variadic:   variadic,
	}
}

func (v Signature) String() string {
	str := "func " + v.returnType.String() + "("

	for i, par := range v.parameters {
		str += par.String()
		if v.variadic || i < len(v.parameters)-1 {
			str += ", "
		}
	}

	if v.variadic {
		str += "..."
	}

	str += ")"

	return str
}

func (v Signature) Equals(t Type) bool {
	fn, ok := t.(*Signature)
	if !ok {
		return false
	}

	if !v.parameters.Equals(fn.parameters) {
		return false
	}

	if !v.returnType.Equals(fn.returnType) {
		return false
	}

	if v.variadic != fn.variadic {
		return false
	}

	return true
}
