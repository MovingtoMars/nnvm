package types

type Struct struct {
	fields TypeList
	packed bool
}

func NewStruct(fields []Type, packed bool) *Struct {
	return &Struct{
		fields: fields,
		packed: packed,
	}
}

func (v Struct) Fields() TypeList {
	return v.fields
}

func (v Struct) Packed() bool {
	return v.packed
}

func (v Struct) String() string {
	str := "{ "

	if v.packed {
		str += "packed "
	}

	for i, field := range v.fields {
		str += field.String()
		if i < len(v.fields)-1 {
			str += ", "
		}
	}

	str += " }"
	return str
}

func (v Struct) Equals(t Type) bool {
	struc, ok := t.(*Struct)
	if !ok {
		return false
	}

	if !v.fields.Equals(struc.fields) {
		return false
	}

	if v.packed != struc.packed {
		return false
	}

	return true
}
