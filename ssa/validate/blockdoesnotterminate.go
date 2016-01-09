package validate

import "github.com/MovingtoMars/nnvm/ssa"

func checkBlockDoesNotTerminate(mod *ssa.Module) error {
	for _, fn := range mod.Functions() {
		for _, block := range fn.Blocks() {
			lastInstr := block.LastInstr()
			if lastInstr == nil || !lastInstr.IsTerminating() {
				return &BlockError{
					Block:   block,
					Message: "Non-terminating block",
				}
			}
		}
	}

	return nil
}
