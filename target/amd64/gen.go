package amd64

import (
	"fmt"

	"github.com/MovingtoMars/nnvm/ssa"
	"github.com/MovingtoMars/nnvm/types"
)

func (v Target) moveValToMem(a *allocator, src ssa.Value, memReg string, memOffset int) {
	/*if !src.Type().Equals(dest.Type()) {
		panic("internal error")
	}*/

	checkTypeSupported(src.Type())

	storesz := TypeStoreSizeInBits(src.Type())

	switch src.Type().(type) {
	case *types.Int, *types.Pointer:
		if lit, ok := src.(*ssa.IntLiteral); ok {
			v.wop("mov%s $%d, %d(#%s)", sizeSuffixBits(storesz), lit.LiteralValue(), memOffset, memReg)
		} else {
			v.wop("mov%s %s, #%s", sizeSuffixBits(storesz), a.valStr(src), regToSize("rax", storesz))
			v.wop("mov%s #%s, %d(#%s)", sizeSuffixBits(storesz), regToSize("rax", storesz), memOffset, memReg)
		}

	default:
		panic("unim")
	}
}

// don't use rax as a src or dest reg
func (v Target) moveMemToMem(srcMemReg, destMemReg string, srcMemOffset, destMemOffset, bytes int) {
	if bytes <= 8 && isPow2(bytes) {
		tmp := regToSize("rax", bytes*8)
		v.wop("mov%s %d(#%s), #%s", sizeSuffixBits(bytes*8), srcMemOffset, srcMemReg, tmp)
		v.wop("mov%s #%s, %d(#%s)", sizeSuffixBits(bytes*8), tmp, destMemOffset, destMemReg)
	} else {
		panic("unim")
	}
}

func (v Target) moveValToVal(a *allocator, src, dest ssa.Value) {
	v.moveValToMem(a, src, "rbp", -a.valOffset(dest))
}

/*
if signed {
		switch sz {
		case 8:
			e.asm("movq (%%%s), %%rax\n", reg)
		case 4:
			e.asm("movslq (%%%s), %%rax\n", reg)
		case 2:
			e.asm("movswq (%%%s), %%rax\n", reg)
		case 1:
			e.asm("movsbq (%%%s), %%rax\n", reg)
		default:
			panic("internal error")
		}
	} else {
		switch sz {
		case 8:
			e.asm("movq (%%%s), %%rax\n", reg)
		case 4:
			e.asm("movzlq (%%%s), %%rax\n", reg)
		case 2:
			e.asm("movzwq (%%%s), %%rax\n", reg)
		case 1:
			e.asm("movzbq (%%%s), %%rax\n", reg)
		default:
			panic("internal error")
		}
	}
*/

// TODO rename
func (v Target) moveIntToReg(a *allocator, val ssa.Value, reg string) {
	storesz := TypeStoreSizeInBits(val.Type())
	reg = regToSize(reg, storesz)
	reg64 := regToSize(reg, 64)

	if storesz > 64 {
		panic("unim")
	}
	if !isPow2(storesz) {
		panic("unim")
	}

	if storesz != 64 {
		v.wop("xorq #%s, #%s", reg64, reg64)
	}

	if lit, ok := val.(*ssa.IntLiteral); ok {
		v.wop("mov%s $%d, #%s", sizeSuffixBits(storesz), lit.LiteralValue(), reg)
	} else {
		v.wop("mov%s %s, #%s", sizeSuffixBits(storesz), a.valStr(val), reg)
	}
}

func (v Target) moveRegToVal(a *allocator, reg string, val ssa.Value) {
	storesz := TypeStoreSizeInBits(val.Type())
	reg = regToSize(reg, storesz)

	if storesz > 64 {
		panic("unim")
	}
	if !isPow2(storesz) {
		panic("unim")
	}

	v.wop("mov%s #%s, %s", sizeSuffixBits(storesz), reg, a.valStr(val))
}

func (v Target) moveFloatToSSEReg(a *allocator, val ssa.Value, reg string) {
	float := val.Type().(*types.Float)
	com := moveInstrForFloatType(float.Type())

	if lit, ok := val.(*ssa.FloatLiteral); ok {
		v.wop("pushq $0x%x", lit.LiteralValue())
		v.wop("%s (#rsp), %s", com, reg)
		v.wop("addq $8, #rsp")
	} else {
		v.wop("%s %s, #%s", com, a.valStr(val), reg)
	}
}

func (v Target) moveSSERegToFloat(a *allocator, reg string, val ssa.Value) {
	float := val.Type().(*types.Float)
	com := moveInstrForFloatType(float.Type())

	v.wop("%s #%s, %s", com, reg, a.valStr(val))
}

func moveInstrForFloatType(typ types.FloatType) string {
	switch typ {
	case types.Float32:
		return "movpls"
	case types.Float64:
		return "movlpd"
	default:
		panic("unim")
	}
}

func checkTypeSupported(typ types.Type) {
	switch typ := typ.(type) {
	case *types.Int:
		switch typ.Width() {
		case 1, 8, 16, 32, 64:
			// all good
		default:
			goto unsupported
		}

	case *types.Array:
		checkTypeSupported(typ.Element())

	case *types.Struct:
		for _, field := range typ.Fields() {
			checkTypeSupported(field)
		}
	}

	return

unsupported:
	panic(fmt.Sprintf("unsupported type: %s", typ))
}

func (v Target) genGlobals() {
	for _, global := range v.mod.Globals() {
		checkTypeSupported(global.Type())

		v.wop(".globl %s", global.Name())
		v.wlabel(global.Name())

		switch typ := global.Type().(*types.Pointer).Element().(type) {
		case *types.Array:
			v.genArrayGlobal(global)

		case *types.Pointer:
			v.genPointerGlobal(global)

		default:
			_ = typ
			panic("unim")
		}
	}
}

func (v Target) genPointerGlobal(global *ssa.Global) {
	init := global.Initialiser()

	switch init := init.(type) {
	case *ssa.GlobalPointerInitialiser:
		v.wop(".quad %s", init.Global().Name())

	default:
		panic("unim")
	}
}

func (v Target) genArrayGlobal(global *ssa.Global) {
	init := global.Initialiser()

	switch init := init.(type) {
	case *ssa.LiteralInitialiser:
		lit := init.Literal()
		if str, ok := lit.(*ssa.StringLiteral); ok {
			v.wop(".ascii \"%s\"", ssa.EscapeString(str.LiteralValue().(string)))
		} else {
			panic("unim")
		}

	default:
		panic("unim")
	}
}

func (v *Target) gen() {
	v.wop(".data")

	v.genGlobals()

	v.wnl()
	v.wop(".text")

	for _, fn := range v.mod.Functions() {
		v.genFunction(fn)
	}
}

func (v *Target) nextLabelName() string {
	v.labelID++
	return fmt.Sprintf(".L%d", v.labelID)
}

func (v *Target) genFunction(fn *ssa.Function) {
	if fn.IsPrototype() {
		return
	}

	allocator := newAllocator()
	allocator.allocate(fn)

	blockLabelMap := make(map[*ssa.Block]string)

	for _, block := range fn.Blocks() {
		blockLabelMap[block] = v.nextLabelName()
	}

	v.wnl()
	v.wop(".globl %s", fn.Name())
	v.wop(".align 16, 0x90") // pad with NOPs

	if v.Platform.IsUnixLike() {
		v.wop(".type %s,@function", fn.Name())
	}

	v.wlabel(fn.Name())

	v.wop("pushq #rbp")
	v.wop("pushq #rbx")
	v.wop("pushq #r15")
	v.wop("movq #rsp, #rbp")
	v.wop("subq $%d, #rsp", (allocator.stackSize|0xF)+1)

	retType := fn.Type().(*types.Signature).ReturnType()
	if sysVClassifyType(retType)[0] == sysVClassMEMORY {
		v.wop("movq #rdi, #r15")
	} else if winIsMemory(retType) {
		v.wop("movq #rdx, #r15")
	}

	v.genSaveFunctionParameters(allocator, fn)

	for _, block := range fn.Blocks() {
		v.wlabel(blockLabelMap[block])

		for _, instr := range block.Instrs() {
			v.genInstr(allocator, instr, blockLabelMap)
		}
	}

}
