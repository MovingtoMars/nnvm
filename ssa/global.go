package ssa

import "github.com/MovingtoMars/nnvm/types"

type Initialiser interface {
	Initialiser()
	String() string
}

// Used for literals as initialisers
type LiteralInitialiser struct {
	lit Literal
}

// TODO reference handling?
func NewLiteralInitialiser(lit Literal) *LiteralInitialiser {
	return &LiteralInitialiser{
		lit: lit,
	}
}

func (_ LiteralInitialiser) Initialiser() {}

func (v LiteralInitialiser) String() string {
	return "literal " + ValueString(v.lit)
}

func (v LiteralInitialiser) Literal() Literal {
	return v.lit
}

// Used to zero out a global as an initialiser
type ZeroInitialiser struct{}

func NewZeroInitialiser() *ZeroInitialiser {
	return &ZeroInitialiser{}
}

func (_ ZeroInitialiser) Initialiser() {}

func (_ ZeroInitialiser) String() string {
	return "zero"
}

type Global struct {
	ReferenceHandler
	NameHandler

	typ         types.Type
	initialiser Initialiser
}

func newGlobal(typ types.Type, init Initialiser, name string) *Global {
	return &Global{
		typ:         typ,
		initialiser: init,
		NameHandler: NameHandler{name: name},
	}
}

func (v *Global) SetInitialiser(init Initialiser) {
	v.initialiser = init
}

func (v Global) Initialiser() Initialiser {
	return v.initialiser
}

func (v Global) Type() types.Type {
	return types.NewPointer(v.typ)
}

func (v *Global) String() string {
	return "glob " + ValueString(v) + " = " + v.initialiser.String()
}
