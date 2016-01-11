package amd64

import (
	"github.com/MovingtoMars/nnvm/ssa"
	"github.com/MovingtoMars/nnvm/target/platform"
	"github.com/MovingtoMars/nnvm/types"
)

func (v Target) genInstr(a *allocator, instr ssa.Instruction, blockLabelMap map[*ssa.Block]string) {
	v.wstring("#" + instr.String() + "\n")

	switch instr := instr.(type) {
	case *ssa.Ret:
		v.genRet(a, instr)
	case *ssa.BinOp:
		v.genBinOp(a, instr)
	case *ssa.ICmp:
		v.genICmp(a, instr)
	case *ssa.Br:
		v.genBr(a, instr, blockLabelMap)
	case *ssa.CondBr:
		v.genCondBr(a, instr, blockLabelMap)
	case *ssa.Call:
		v.genCall(a, instr)
	case *ssa.Convert:
		v.genConvert(a, instr)
	case *ssa.GEP:
		v.genGEP(a, instr)
	case *ssa.Load:
		v.genLoad(a, instr)
	case *ssa.Alloc:
		v.genAlloc(a, instr)
	case *ssa.Store:
		v.genStore(a, instr)
	case *ssa.Phi:
		// do nothing
	case *ssa.Unreachable:
		v.wop("hlt")
	default:
		panic("unim")
	}
}

func (v Target) genLoad(a *allocator, instr *ssa.Load) {
	v.moveIntToReg(a, ssa.GetOperands(instr)[0], "r11")
	v.moveMemToMem("r11", "rbp", 0, -a.valOffset(instr), TypeStoreSizeInBits(instr.Type())/8)
}

func (v Target) genStore(a *allocator, instr *ssa.Store) {
	ops := ssa.GetOperands(instr)
	v.moveIntToReg(a, ops[0], "r11")
	v.moveMemToMem("rbp", "r11", -a.valOffset(ops[1]), 0, TypeStoreSizeInBits(ops[1].Type())/8)
}

func (v Target) genAlloc(a *allocator, instr *ssa.Alloc) {
	v.wop("subq $%d, #rsp", TypeStoreSizeInBits(instr.Type())/8)
	v.wop("movq #rsp, %s", a.valStr(instr))
}

func (v Target) genGEP(a *allocator, instr *ssa.GEP) {
	ops := ssa.GetOperands(instr)
	val := ops[0]
	indexes := ops[1:]

	typ := val.Type()

	v.moveIntToReg(a, val, "rax")

	for _, index := range indexes {
		switch styp := typ.(type) {
		case *types.Pointer:
			v.handleGEPPointerOrArray(a, index, TypeStoreSizeInBits(styp.Element()))
			typ = styp.Element()

		case *types.Array:
			v.handleGEPPointerOrArray(a, index, TypeStoreSizeInBits(styp.Element()))
			typ = styp.Element()

		case *types.Struct:
			i := index.(*ssa.IntLiteral).LiteralValue().(uint64)
			v.wop("addq $%d, #rax", newStructLayout(styp).fieldOffsetBits(int(i))/8)

			typ = styp.Fields()[i]

		default:
			panic("unim")
		}
	}

	v.wop("movq #rax, %s", a.valStr(instr))
}

// assumes current GEP pointer is in %rax
func (v Target) handleGEPPointerOrArray(a *allocator, index ssa.Value, storesz int) {
	if lit, ok := index.(*ssa.IntLiteral); ok {
		litval := int64(lit.LiteralValue().(uint64))

		if litval == 0 {
			return
		}
	}

	v.moveIntToReg(a, index, "r11")
	if isPow2(storesz) && storesz < 64 {
		v.wop("leaq (#rax,#r11,%d), #rax", storesz/8)
	} else {
		v.wop("imulq $%d, #r11", storesz/8)
		v.wop("addq #r11, #rax")
	}
}

func (v Target) genCall(a *allocator, instr *ssa.Call) {
	v.genLoadCallArguments(a, instr)
	v.wop("call %s", ssa.GetOperands(instr)[0].Name())

	if v.Platform == platform.Windows { // TODO move this
		totalMem := winTotalMemSizeBits(ssa.GetOperands(instr)[0].Type().(*types.Signature).Parameters())
		if totalMem > 0 {
			v.wop("addq $%d, #rsp", totalMem/8)
		}
	}

	v.genSaveReturnValue(a, instr)
}

func (v Target) genConvert(a *allocator, instr *ssa.Convert) {
	op := ssa.GetOperands(instr)[0]

	checkTypeSupported(op.Type())
	checkTypeSupported(instr.Type())

	targetStoresz := TypeStoreSizeInBits(instr.Type())
	targetsz := TypeSizeInBits(instr.Type())
	opsz := TypeSizeInBits(op.Type())

	switch instr.ConvertType() {
	case ssa.ConvertBitcast:
		v.moveValToVal(a, op, instr)

	case ssa.ConvertTrunc:
		v.moveIntToReg(a, op, "rax")
		if !isRegSizeBits(targetsz) {
			v.wop("andq $%d, #rax", (1<<uint(targetsz))-1)
		}
		v.wop("mov%s #%s, %s", sizeSuffixBits(targetStoresz), regToSize("rax", targetStoresz), a.valStr(instr))

	case ssa.ConvertZExt:
		v.moveIntToReg(a, op, "rax")
		v.moveRegToVal(a, "rax", instr)

	case ssa.ConvertSExt:
		v.genSExt(a, instr, op, opsz, targetStoresz, targetsz)

	case ssa.ConvertIntToPtr, ssa.ConvertPtrToInt:
		v.moveIntToReg(a, op, "rax")
		v.moveRegToVal(a, "rax", instr)

	default:
		panic("unim")
	}
}

func (v Target) genSExt(a *allocator, instr *ssa.Convert, op ssa.Value, opsz, targetStoresz, targetsz int) {
	v.moveIntToReg(a, op, "rax")

	if opsz == 1 {
		v.wop("xorq #rdx, #rdx")
		v.wop("movq $0x%X, #rcx", uint64((1<<uint(64))-1))
		v.wop("cmpb $1, #al")
		v.wop("cmoveq #rcx, #rdx")
		v.wop("movq #rdx, #rax")
	} else {
		extops := []string{"cbw", "cwde", "cdqe"}

		szToI := func(sz int) int {
			if sz == 8 {
				return 0
			} else if sz == 16 {
				return 1
			} else if sz == 32 {
				return 2
			} else if sz == 64 {
				return 3
			}
			panic("unim")
		}

		opi := szToI(opsz)
		targeti := szToI(targetsz)

		for opi < targeti {
			v.wop("%s", extops[opi])
			opi++
		}
	}

	v.moveRegToVal(a, "rax", instr)
}

func (v Target) genICmp(a *allocator, instr *ssa.ICmp) {
	ops := ssa.GetOperands(instr)

	v.wop("xorq #rcx, #rcx")
	v.wop("movq $1, #rdx")

	sz := TypeStoreSizeInBits(ops[0].Type())

	rax := regToSize("rax", sz)
	rbx := regToSize("rbx", sz)

	v.moveIntToReg(a, ops[1], rbx)
	v.moveIntToReg(a, ops[0], rax)
	v.wop("cmp%s #%s, #%s", sizeSuffixBits(sz), rbx, rax)

	moveType := ""

	switch instr.Predicate() {
	case ssa.IntEQ:
		moveType = "e"
	case ssa.IntNEQ:
		moveType = "ne"
	case ssa.IntUGT:
		moveType = "a" // above
	case ssa.IntUGE:
		moveType = "ae" // above or equal
	case ssa.IntULT:
		moveType = "b" // below
	case ssa.IntULE:
		moveType = "be" // below or equal
	case ssa.IntSGT:
		moveType = "g" // greater
	case ssa.IntSGE:
		moveType = "ge" // greater or equal
	case ssa.IntSLT:
		moveType = "l" // less
	case ssa.IntSLE:
		moveType = "le" // less or equal
	default:
		panic("unimplemented int predicate")
	}

	v.wop("cmov%sq #rdx, #rcx", moveType)
	v.wop("movb #cl, %s", a.valStr(instr))
}

func (v Target) genBr(a *allocator, instr *ssa.Br, blockLabelMap map[*ssa.Block]string) {
	target := ssa.GetOperands(instr)[0].(*ssa.Block)

	v.handleBrPhi(a, instr.Block(), target)

	v.wop("jmp %s", blockLabelMap[target])
}

func (v Target) genCondBr(a *allocator, instr *ssa.CondBr, blockLabelMap map[*ssa.Block]string) {
	ops := ssa.GetOperands(instr)
	trueTarget := ops[1].(*ssa.Block)
	falseTarget := ops[2].(*ssa.Block)

	v.handleBrPhi(a, instr.Block(), trueTarget)
	v.handleBrPhi(a, instr.Block(), falseTarget)

	v.moveIntToReg(a, ops[0], "al")
	v.wop("testb #al, #al")
	v.wop("jne %s", blockLabelMap[trueTarget])
	v.wop("je %s", blockLabelMap[falseTarget])
}

func (v Target) handleBrPhi(a *allocator, brBlock, targetBlock *ssa.Block) {
	type phiInc struct {
		phi   *ssa.Phi
		val   ssa.Value
		block *ssa.Block
	}
	phiIncs := []*phiInc{}

	for _, targetBlockInstr := range targetBlock.Instrs() {
		if phi, ok := targetBlockInstr.(*ssa.Phi); ok {
			for i := 0; i < phi.NumIncoming(); i++ {
				incVal, incBlock := phi.GetIncoming(i)

				if incBlock == brBlock {
					phiIncs = append(phiIncs, &phiInc{
						phi:   phi,
						val:   incVal,
						block: incBlock,
					})
				}
			}
		}
	}

	dep := false // check if we have a circular dependency chain among phis
	for i, inc := range phiIncs {
		for j, inc2 := range phiIncs {
			if i != j {
				if inc2.val == inc.phi {
					dep = true
				}
			}
		}
	}

	if dep {
		for _, inc := range phiIncs {
			v.wop("subq $%d, #rsp", TypeStoreSizeInBits(inc.val.Type())/8)
			v.moveValToMem(a, inc.val, "rsp", 0)
		}

		for i := len(phiIncs) - 1; i >= 0; i-- {
			v.moveMemToMem("rsp", "rbp", 0, -a.valOffset(phiIncs[i].phi), TypeStoreSizeInBits(phiIncs[i].phi.Type())/8)
			v.wop("addq $%d, #rsp", TypeStoreSizeInBits(phiIncs[i].phi.Type())/8)
		}
	} else {
		for _, inc := range phiIncs {
			v.moveValToVal(a, inc.val, inc.phi)
		}
	}
}

func (v Target) genRet(a *allocator, instr *ssa.Ret) {
	retVal := ssa.GetOperands(instr)[0]
	v.genLoadReturnValue(a, retVal)

	v.wop("movq #rbp, #rsp")
	v.wop("popq #r15")
	v.wop("popq #rbx")
	v.wop("popq #rbp")
	v.wop("retq")
}

var (
	binOpIntOps = []ssa.BinOpType{ssa.BinOpAdd, ssa.BinOpSub, ssa.BinOpMul,
		ssa.BinOpSDiv, ssa.BinOpUDiv, ssa.BinOpSRem, ssa.BinOpURem,
		ssa.BinOpShl, ssa.BinOpLShr, ssa.BinOpAShr, ssa.BinOpAnd, ssa.BinOpOr,
		ssa.BinOpXor}
	binOpIntOpStrs = []string{"add", "sub", "mul", "idiv", "div", "idiv",
		"div", "shl", "shr", "sar", "and", "or", "xor"}
)

func (v Target) genBinOp(a *allocator, instr *ssa.BinOp) {
	checkTypeSupported(instr.Type())

	//floatOps := []ssa.BinOpType{ssa.BinOpFAdd, ssa.BinOpFSub, ssa.BinOpFMul, ssa.BinOpFDiv, ssa.BinOpFRem}

	for i, intOp := range binOpIntOps {
		if instr.BinOpType() == intOp {
			v.genIntBinOp(a, instr, i)
		}
	}
}

func (v Target) genIntBinOp(a *allocator, instr *ssa.BinOp, opIndex int) {
	ops := ssa.GetOperands(instr)
	sz := TypeSizeInBits(ops[0].Type())
	suffix := sizeSuffixBits(sz)

	rax := regToSize("rax", sz)
	rcx := regToSize("rcx", sz)
	v.moveIntToReg(a, ops[0], rax)
	v.moveIntToReg(a, ops[1], rcx)

	switch instr.BinOpType() {
	case ssa.BinOpAdd, ssa.BinOpSub, ssa.BinOpAnd, ssa.BinOpOr, ssa.BinOpXor:
		v.wop("%s%s #%s, #%s", binOpIntOpStrs[opIndex], suffix, rcx, rax)

	case ssa.BinOpMul:
		v.wop("mul%s #%s", suffix, rcx)

	case ssa.BinOpShl, ssa.BinOpAShr, ssa.BinOpLShr:
		v.wop("%s%s #cl, #%s", binOpIntOpStrs[opIndex], suffix, rax)

	default:
		panic("unim")
	}

	if sz == 1 {
		v.wop("andq $1, #%s", rax)
	}
	v.wop("mov%s #%s, %s", suffix, rax, a.valStr(instr))

}
