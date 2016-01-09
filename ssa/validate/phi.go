package validate

import (
	"fmt"

	"github.com/MovingtoMars/nnvm/ssa"
	"github.com/MovingtoMars/nnvm/ssa/analysis"
)

func checkPhi(instr *ssa.Phi, dom *analysis.DominatorTree) error {
	if err := errIfNonFirstClassType(instr.Type(), instr); err != nil {
		return err
	}

	incomingBlockMap := make(map[*ssa.Block]bool, instr.NumIncoming())
	for _, pred := range dom.BlockCFG().NodeForBlock(instr.Block()).Prev() {
		incomingBlockMap[pred.Block()] = false
	}

	for i := 0; i < instr.NumIncoming(); i++ {
		val, block := instr.GetIncoming(i)

		if seen, exists := incomingBlockMap[block]; !exists {
			return &InstrError{
				Instr:   instr,
				Message: "Impossible incoming block `" + block.Name() + "`",
			}
		} else if seen {
			return &InstrError{
				Instr:   instr,
				Message: "Duplicate incoming block `" + block.Name() + "`",
			}
		}

		if err := errIfMismatchedTypes(instr.Type(), val.Type(), instr); err != nil {
			return err
		}

		incomingBlockMap[block] = true

		if valInstr, ok := val.(ssa.Instruction); ok {
			// for an incoming value to be valid, the incoming block must be dominated by the block the value comes from
			if !dom.NodeForBlock(block).DominatedBy(dom.NodeForBlock(valInstr.Block()), false) {
				return &InstrError{
					Instr:   instr,
					Message: fmt.Sprintf("Value `%s` must dominate block `%s`", val.Name(), block.Name()),
				}
			}
		}

	}

	for block, seen := range incomingBlockMap {
		if !seen {
			return &InstrError{
				Instr:   instr,
				Message: "Missing incoming block `" + block.Name() + "`",
			}
		}
	}

	return nil
}
