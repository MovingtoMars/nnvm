package validate

import "github.com/MovingtoMars/nnvm/ssa"

func checkEmptyBlock(mod *ssa.Module) error {
	for _, fn := range mod.Functions() {
		for _, block := range fn.Blocks() {
			if len(block.Instrs()) == 0 {
				return &BlockError{
					Block:   block,
					Message: "Empty block",
				}
			}
		}
	}

	return nil
}
