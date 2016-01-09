package ssa

type Store struct {
	BlockHandler

	location Value
	value    Value
}

func newStore(location, value Value) *Store {
	return &Store{
		location: location,
		value:    value,
	}
}

func (v *Store) operands() []*Value {
	return []*Value{&v.location, &v.value}
}

func (v Store) String() string {
	return "store " + ValueString(v.location) + ", " + ValueString(v.value)
}

func (_ Store) IsTerminating() bool {
	return false
}
