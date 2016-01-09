package validate

import "github.com/MovingtoMars/nnvm/ssa"

func checkFunctionNames(mod *ssa.Module) error {
	seenNames := make(map[string]bool, len(mod.Functions()))

	for _, fn := range mod.Functions() {
		if fn.Name() == "" {
			return &FunctionError{
				Function: fn,
				Message:  "Empty function name",
			}
		}

		if seenNames[fn.Name()] {
			return &FunctionError{
				Function: fn,
				Message:  "Duplicate function name `" + fn.Name() + "`",
			}
		}

		seenNames[fn.Name()] = true
	}

	return nil
}
