package validate

import "github.com/MovingtoMars/nnvm/ssa"

func checkEntryPhi(mod *ssa.Module) error {
	for _, fn := range mod.Functions() {
		if !fn.IsPrototype() {
			for _, instr := range fn.EntryBlock().Instrs() {
				_, ok := instr.(*ssa.Phi)
				if ok {
					return InstrError{
						Instr:   instr,
						Message: "Phi node in entry block",
					}
				}
			}
		}
	}

	return nil
}
