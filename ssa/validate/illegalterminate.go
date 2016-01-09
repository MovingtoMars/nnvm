package validate

import "github.com/MovingtoMars/nnvm/ssa"

func checkIllegalTerminate(mod *ssa.Module) error {
	for _, fn := range mod.Functions() {
		for _, block := range fn.Blocks() {
			instrs := block.Instrs()
			for _, instr := range instrs[:len(instrs)-1] {
				if instr.IsTerminating() {
					return &InstrError{
						Instr:   instr,
						Message: "Terminating instruction in middle of block",
					}
				}
			}
		}
	}

	return nil
}
