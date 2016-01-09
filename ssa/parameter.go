package ssa

import "github.com/MovingtoMars/nnvm/types"

type Parameter struct {
	ReferenceHandler
	NameHandler

	typ types.Type
}

func newParameter(typ types.Type, name string) *Parameter {
	return &Parameter{
		typ:         typ,
		NameHandler: NameHandler{name: name},
	}
}

func (v Parameter) Type() types.Type {
	return v.typ
}
