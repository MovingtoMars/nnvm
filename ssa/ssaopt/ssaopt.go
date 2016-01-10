package ssaopt

import (
	"github.com/MovingtoMars/nnvm/ssa"
	"github.com/MovingtoMars/nnvm/ssa/validate"
)

type Pass interface {
	Run(*ssa.Module)
	String() string
}

type PassList struct {
	passes []Pass
}

func NewPassList(passes ...Pass) *PassList {
	return &PassList{
		passes: passes,
	}
}

// Will panic if the module is invalid!
// Can be called on multiple modules in parallel.
func (v PassList) Run(mod *ssa.Module) {
	if err := validate.Validate(mod); err != nil {
		panic(err)
	}

	for _, pass := range v.passes {
		pass.Run(mod)

		if err := validate.Validate(mod); err != nil {
			panic("Pass `" + pass.String() + "` caused validation error: " + err.Error())
		}
	}
}

func (v *PassList) Add(passes ...Pass) {
	v.passes = append(v.passes, passes...)
}

func (v PassList) String() string {
	str := "PassList: {"

	for i, pass := range v.passes {
		str += pass.String()

		if i < len(v.passes)-1 {
			str += ", "
		}
	}

	return str + "}"
}
