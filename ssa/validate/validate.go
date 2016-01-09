package validate

import (
	"fmt"

	"github.com/MovingtoMars/nnvm/types"

	"github.com/MovingtoMars/nnvm/ssa"
)

// TODO the check functions should be renamed to make it more clear what they do

var moduleCheckFunctions = []func(*ssa.Module) error{
	checkEmptyBlock,
	checkIllegalTerminate,
	checkBlockDoesNotTerminate,
	checkBranchToEntry,
	checkEntryPhi,
	checkInstrs,
	checkFunctionNames,
	checkGlobals,
}

// Validate attempts to validate the passed module, returning an error if validation fails.
// Note that if the module is modified between calling Validate and viewing the error, the error may be incorrect or nonsensical.
func Validate(mod *ssa.Module) error {
	for _, fn := range moduleCheckFunctions {
		if err := fn(mod); err != nil {
			return err
		}
	}

	return nil
}

func errIfMismatchedTypes(t1, t2 types.Type, i ssa.Instruction) error {
	if !t1.Equals(t2) {
		return &InstrError{
			Message: fmt.Sprintf("Mismatched types `%s` and `%s`", t1, t2),
			Instr:   i,
		}
	}

	return nil
}

// TODO check subtypes
func errIfNonFirstClassType(t types.Type, i ssa.Instruction) error {
	if !types.IsFirstClass(t) {
		return &InstrError{
			Message: "Illegal non-first class type",
			Instr:   i,
		}
	}

	return nil
}

func errIfNonFirstClassTypes(i ssa.Instruction, ty ...types.Type) error {
	for _, t := range ty {
		if err := errIfNonFirstClassType(t, i); err != nil {
			return err
		}
	}

	return nil
}

// TODO check subtypes
func errIfVoidType(t types.Type, i ssa.Instruction) error {
	_, ok := t.(types.Void)
	if ok {
		return &InstrError{
			Message: "Illegal void type",
			Instr:   i,
		}
	}

	return nil
}

func VisitInstrs(mod *ssa.Module, visitFn func(ssa.Instruction)) {
	for _, fn := range mod.Functions() {
		for _, block := range fn.Blocks() {
			for _, instr := range block.Instrs() {
				visitFn(instr)
			}
		}
	}
}

func VisitBlocks(mod *ssa.Module, visitFn func(*ssa.Block)) {
	for _, fn := range mod.Functions() {
		for _, block := range fn.Blocks() {
			visitFn(block)
		}
	}
}

type FunctionError struct {
	Message  string
	Function *ssa.Function
}

func (v FunctionError) Error() string {
	return fmt.Sprintf("FunctionError: %s\n -> %s", v.Message, ssa.FunctionTrace(v.Function))
}

type InstrError struct {
	Message string
	Instr   ssa.Instruction
}

func (v InstrError) Error() string {
	return fmt.Sprintf("InstrError: %s\n -> %s", v.Message, ssa.InstrTrace(v.Instr))
}

type GlobalError struct {
	Message string
	Global  *ssa.Global
}

func (v GlobalError) Error() string {
	return fmt.Sprintf("GlobalError: %s\n -> %s", v.Message, ssa.GlobalTrace(v.Global))
}

type BlockError struct {
	Message string
	Block   *ssa.Block
}

func (v BlockError) Error() string {
	return fmt.Sprintf("BlockError: %s\n -> %s", v.Message, ssa.BlockTrace(v.Block))
}
