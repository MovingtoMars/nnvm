package validate

import (
	"github.com/MovingtoMars/nnvm/ssa"
	"github.com/MovingtoMars/nnvm/types"
)

func checkMismatchedTypesGlobal(t1, t2 types.Type, g *ssa.Global) error {
	if !t1.Equals(t2) {
		return &GlobalError{
			Global:  g,
			Message: "Mismatched types: `" + t1.String() + "` and `" + t2.String() + "`",
		}
	}

	return nil
}

func checkGlobals(mod *ssa.Module) error {
	for _, global := range mod.Globals() {
		init := global.Initialiser()

		switch init := init.(type) {
		case *ssa.LiteralInitialiser:
			litPtr := types.NewPointer(init.Literal().Type())
			if err := checkMismatchedTypesGlobal(global.Type(), litPtr, global); err != nil {
				return err
			}

		case *ssa.GlobalPointerInitialiser:
			globPtr := types.NewPointer(init.Global().Type())
			if err := checkMismatchedTypesGlobal(global.Type(), globPtr, global); err != nil {
				return err
			}

		case *ssa.ZeroInitialiser:
			// do nothing

		default:
			panic("unim")
		}
	}

	return nil
}
