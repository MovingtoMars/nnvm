package validate

import (
	"fmt"

	"github.com/MovingtoMars/nnvm/ssa"
)

func checkBranchToEntry(mod *ssa.Module) error {
	makeError := func(i ssa.Instruction) error {
		return &InstrError{
			Instr:   i,
			Message: fmt.Sprintf("Branch to entry block `%%%s`", i.Block().Name()),
		}
	}

	for _, fn := range mod.Functions() {
		for _, block := range fn.Blocks() {
			lastInstr := block.LastInstr()

			switch i := lastInstr.(type) {
			case *ssa.Br:
				if block := ssa.GetOperands(i)[0].(*ssa.Block); block.IsEntry() {
					return makeError(i)
				}

			case *ssa.CondBr:
				for _, op := range ssa.GetOperands(i)[1:3] {
					if op.(*ssa.Block).IsEntry() {
						return makeError(i)
					}
				}

			}
		}
	}

	return nil
}
