package validate

import (
	"fmt"

	"github.com/MovingtoMars/nnvm/ssa"
	"github.com/MovingtoMars/nnvm/ssa/analysis"
	"github.com/MovingtoMars/nnvm/types"
)

func checkInstrs(mod *ssa.Module) error {
	for _, fn := range mod.Functions() {
		var blockDomTree *analysis.DominatorTree
		if !fn.IsPrototype() {
			blockDomTree = analysis.NewBlockDominatorTree(analysis.NewBlockCFG(fn))
		}

		for _, block := range fn.Blocks() {
			for _, instr := range block.Instrs() {
				if err := checkInstr(instr, blockDomTree); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func checkInstr(instr ssa.Instruction, blockDomTree *analysis.DominatorTree) error {
	thisBlockNode := blockDomTree.NodeForBlock(instr.Block())

	if _, ok := instr.(*ssa.Phi); !ok {
		for _, op := range ssa.GetOperands(instr) {
			if opInstr, ok := op.(ssa.Instruction); ok {
				opInstrBlock := opInstr.Block()
				if thisBlockNode.DominatedBy(blockDomTree.NodeForBlock(opInstrBlock), true) {
					continue
				} else if opInstrBlock == instr.Block() {
					if opInstrBlock.InstrIndex(opInstr) < opInstrBlock.InstrIndex(instr) {
						continue
					}
				}

				return &InstrError{
					Instr:   instr,
					Message: "Instruction is not dominated by operand `" + ssa.ValueString(op) + "`",
				}
			}
		}
	}

	switch i := instr.(type) {
	case *ssa.BinOp:
		return checkBinOp(i)
	case *ssa.Load:
		return checkLoad(i)
	case *ssa.Call:
		return checkCall(i)
	case *ssa.Alloc:
		return checkAlloc(i)
	case *ssa.Store:
		return checkStore(i)
	case *ssa.Convert:
		return checkConvert(i)
	case *ssa.ICmp:
		return checkICmp(i)
	case *ssa.CondBr:
		return checkCondBr(i)
	case *ssa.Br:
		return checkBr(i)
	case *ssa.Phi:
		return checkPhi(i, blockDomTree)
	case *ssa.Ret:
		return checkRet(i)
	case *ssa.GEP:
		return checkGEP(i)
	case *ssa.Unreachable:
		// do nothing
	default:
		panic("unim")
	}

	return nil
}

func checkGEP(instr *ssa.GEP) error {
	ops := ssa.GetOperands(instr)

	instrErr := func(message string) error {
		return &InstrError{
			Instr:   instr,
			Message: message,
		}
	}

	value := ops[0]
	indexes := ops[1:]

	typ := value.Type()

	for i, index := range indexes {
		switch styp := typ.(type) {
		case *types.Pointer:
			if i != 0 {
				return instrErr("Index " + fmt.Sprintf("%d", i) + " dereferences a pointer (only the index 0 may dereference a pointer)")
			}

			typ = styp.Element()

		case *types.Array:
			typ = styp.Element()

		case *types.Struct:
			lit, ok := index.(*ssa.IntLiteral)
			if !ok {
				return instrErr("Expected int literal at index " + fmt.Sprintf("%d", i))
			}

			if int(lit.LiteralValue().(uint64)) >= len(styp.Fields()) {
				return instrErr("Index " + fmt.Sprintf("%d", i) + " has value greater than number of struct fields")
			}

			typ = styp.Fields()[lit.LiteralValue().(uint64)]

		default:
			return instrErr("Index " + fmt.Sprintf("%d", i) + " is invalid")
		}
	}

	return nil
}

func checkRet(instr *ssa.Ret) error {
	fnReturnType := instr.Block().Function().Type().(*types.Signature).ReturnType()
	ops := ssa.GetOperands(instr)

	if ops[0] == nil {
		if _, ok := fnReturnType.(types.Void); !ok {
			return &InstrError{
				Instr:   instr,
				Message: "Expected return value of type `" + fnReturnType.String() + "`",
			}
		}
	} else if err := errIfMismatchedTypes(ops[0].Type(), fnReturnType, instr); err != nil {
		return err
	}

	return nil
}

func checkBr(instr *ssa.Br) error {
	ops := ssa.GetOperands(instr)
	if err := errIfNotLabelType(instr, ops[0].Type()); err != nil {
		return err
	}
	return nil
}

func checkCondBr(instr *ssa.CondBr) error {
	ops := ssa.GetOperands(instr)

	if err := errIfNotIntType(instr, ops[0].Type()); err != nil {
		return err
	}

	if !ops[0].Type().Equals(types.NewInt(1)) {
		return &InstrError{
			Instr:   instr,
			Message: "Expected type i1, found `" + ops[0].Type().String() + "`",
		}
	}

	for _, target := range ops[1:] {
		if err := errIfNotLabelType(instr, target.Type()); err != nil {
			return err
		}
	}

	return nil
}

func checkICmp(instr *ssa.ICmp) error {
	ops := ssa.GetOperands(instr)

	if err := errIfMismatchedTypes(ops[0].Type(), ops[1].Type(), instr); err != nil {
		return err
	} else if err := errIfNotIntType(instr, ops[0].Type()); err != nil {
		return err
	}

	return nil
}

func errIfNotLabelType(i ssa.Instruction, t types.Type) error {
	_, ok := t.(*types.Label)
	if !ok {
		return &InstrError{
			Instr:   i,
			Message: "Expected label type, found `" + t.String() + "`",
		}
	}
	return nil
}

func errIfNotIntType(i ssa.Instruction, t types.Type) error {
	_, ok := t.(*types.Int)
	if !ok {
		return &InstrError{
			Instr:   i,
			Message: "Expected int type, found `" + t.String() + "`",
		}
	}
	return nil
}

func errIfNotFloatType(i ssa.Instruction, t types.Type) error {
	_, ok := t.(*types.Float)
	if !ok {
		return &InstrError{
			Instr:   i,
			Message: "Expected float type, found `" + t.String() + "`",
		}
	}
	return nil
}

func errIfNotPointerType(i ssa.Instruction, t types.Type) error {
	_, ok := t.(*types.Pointer)
	if !ok {
		return &InstrError{
			Instr:   i,
			Message: "Expected pointer type, found `" + t.String() + "`",
		}
	}
	return nil
}

// this function is bad
func checkConvert(instr *ssa.Convert) error {
	ops := ssa.GetOperands(instr)
	srcType := ops[0].Type()
	destType := instr.Type()

	mustBeInt := make([]types.Type, 0, 2)
	mustBeFloat := make([]types.Type, 0, 2)
	mustBePointer := make([]types.Type, 0, 2)

	switch instr.ConvertType() {
	case ssa.ConvertSExt, ssa.ConvertZExt, ssa.ConvertTrunc:
		mustBeInt = append(mustBeInt, srcType, destType)

	case ssa.ConvertBitcast:
		mustBePointer = append(mustBePointer, srcType, destType)

	case ssa.ConvertFExt, ssa.ConvertFTrunc:
		mustBeFloat = append(mustBeFloat, srcType, destType)

	case ssa.ConvertFToUI, ssa.ConvertFToSI:
		mustBeFloat = append(mustBeFloat, srcType)
		mustBeInt = append(mustBeInt, destType)

	case ssa.ConvertUIToF, ssa.ConvertSIToF:
		mustBeFloat = append(mustBeFloat, destType)
		mustBeInt = append(mustBeInt, srcType)

	case ssa.ConvertPtrToInt:
		mustBePointer = append(mustBePointer, srcType)
		mustBeInt = append(mustBeInt, destType)

	case ssa.ConvertIntToPtr:
		mustBeInt = append(mustBeInt, srcType)
		mustBePointer = append(mustBePointer, destType)

	default:
		panic("unim")
	}

	for _, t := range mustBeInt {
		if err := errIfNotIntType(instr, t); err != nil {
			return err
		}
	}
	for _, t := range mustBeFloat {
		if err := errIfNotFloatType(instr, t); err != nil {
			return err
		}
	}
	for _, t := range mustBePointer {
		if err := errIfNotPointerType(instr, t); err != nil {
			return err
		}
	}

	instrError := func(message string) error {
		return &InstrError{
			Instr:   instr,
			Message: message,
		}
	}

	switch instr.ConvertType() {
	case ssa.ConvertSExt, ssa.ConvertZExt:
		if srcType.(*types.Int).Width() >= destType.(*types.Int).Width() {
			return instrError("sext/zext requires src width < dest width")
		}

	case ssa.ConvertTrunc:
		if srcType.(*types.Int).Width() <= destType.(*types.Int).Width() {
			return instrError("trunc requires src width > dest width")
		}

	case ssa.ConvertBitcast:
		// do nothing

	case ssa.ConvertFExt:
		if srcType.(*types.Float).Type().CanExtendTo(destType.(*types.Float).Type()) {
			return instrError("fext cannot convert from " + srcType.String() + " to " + destType.String())
		}

	case ssa.ConvertFTrunc:
		if srcType.(*types.Float).Type().CanTruncateTo(destType.(*types.Float).Type()) {
			return instrError("ftrunc cannot convert from " + srcType.String() + " to " + destType.String())
		}

	case ssa.ConvertFToUI, ssa.ConvertFToSI, ssa.ConvertUIToF, ssa.ConvertSIToF, ssa.ConvertPtrToInt, ssa.ConvertIntToPtr:
		// nothing to do

	default:
		panic("unim")
	}

	return nil
}

func checkStore(instr *ssa.Store) error {
	ops := ssa.GetOperands(instr)

	ptr, ok := ops[0].Type().(*types.Pointer)
	if !ok {
		return &InstrError{
			Instr:   instr,
			Message: "Expected pointer type, found `" + ops[0].Type().String() + "`",
		}
	}

	if err := errIfMismatchedTypes(ptr.Element(), ops[1].Type(), instr); err != nil {
		return err
	} else if err := errIfNonFirstClassType(ops[1].Type(), instr); err != nil {
		return err
	}

	return nil
}

func checkAlloc(instr *ssa.Alloc) error {
	return errIfNonFirstClassType(instr.Type(), instr)
}

func checkCall(instr *ssa.Call) error {
	ops := ssa.GetOperands(instr)

	sig, ok := ops[0].Type().(*types.Signature)
	if !ok {
		return &InstrError{
			Instr:   instr,
			Message: "Expected function type, found `" + ops[0].Type().String() + "`",
		}
	}

	fn := ops[0].(*ssa.Function)

	if len(ops)-1 > len(sig.Parameters()) {
		if !sig.Variadic() {
			return &InstrError{
				Instr:   instr,
				Message: "Too many arguments to function `" + ssa.ValueIdentifier(fn) + "`",
			}
		}
	} else if len(ops)-1 < len(sig.Parameters()) {
		return &InstrError{
			Instr:   instr,
			Message: "Too few arguments to function `" + ssa.ValueIdentifier(fn) + "`",
		}
	}

	for _, arg := range ops[1:] {
		if err := errIfNonFirstClassType(arg.Type(), instr); err != nil {
			return err
		}
	}

	for i, par := range sig.Parameters() {
		if err := errIfMismatchedTypes(ops[i+1].Type(), par, instr); err != nil {
			return err
		}
	}

	return nil
}

func checkLoad(instr *ssa.Load) error {
	op := ssa.GetOperands(instr)[0]

	ptr, ok := op.Type().(*types.Pointer)
	if !ok {
		return &InstrError{
			Instr:   instr,
			Message: "Expected pointer type, found `" + instr.Type().String() + "`",
		}
	}

	if !types.IsFirstClass(ptr.Element()) {
		return &InstrError{
			Instr:   instr,
			Message: "Pointer element type is not first class",
		}
	}

	return nil
}

func checkBinOp(instr *ssa.BinOp) error {
	ops := ssa.GetOperands(instr)

	if err := errIfMismatchedTypes(ops[0].Type(), ops[1].Type(), instr); err != nil {
		return err
	}

	switch instr.BinOpType() {
	case ssa.BinOpAdd,
		ssa.BinOpSub,
		ssa.BinOpMul,
		ssa.BinOpSDiv,
		ssa.BinOpUDiv,
		ssa.BinOpSRem,
		ssa.BinOpURem,
		ssa.BinOpShl,
		ssa.BinOpLShr,
		ssa.BinOpAShr,
		ssa.BinOpAnd,
		ssa.BinOpOr,
		ssa.BinOpXor:
		_, ok := ops[0].Type().(*types.Int)
		if !ok {
			return &InstrError{
				Instr:   instr,
				Message: "`" + instr.BinOpType().String() + "` requires int",
			}
		}

	case ssa.BinOpFAdd,
		ssa.BinOpFSub,
		ssa.BinOpFMul,
		ssa.BinOpFDiv,
		ssa.BinOpFRem:
		_, ok := ops[0].Type().(*types.Float)
		if !ok {
			return &InstrError{
				Instr:   instr,
				Message: "`" + instr.BinOpType().String() + "` requires float",
			}
		}

	default:
		panic("unim")
	}

	return nil
}
