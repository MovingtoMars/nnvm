package amd64

import (
	"github.com/MovingtoMars/nnvm/ssa"
	"github.com/MovingtoMars/nnvm/types"
)

func winIsInteger(typ types.Type) bool {
	if winIsFloat(typ) {
		return false
	}

	storesz := TypeStoreSizeInBits(typ)
	return storesz <= 64 && isPow2(storesz)
}

func winIsFloat(typ types.Type) bool {
	_, ok := typ.(*types.Float)
	return ok
}

func winIsMemory(typ types.Type) bool {
	if _, ok := typ.(types.Void); ok {
		return false
	}

	return !(winIsInteger(typ) || winIsFloat(typ))
}

/*type winPar struct {
	isMem bool
	mem   string    // if isMem is true
	val   ssa.Value // if isMem is false
}*/

func winTotalMemSizeBits(pars []types.Type) int {
	total := 0
	for _, par := range pars {
		for total%(16*8) != 0 { // lol
			total++
		}

		if winIsMemory(par) {
			total += TypeStoreSizeInBits(par)
		}
	}
	return total
}

func winTotalCallStackSizeBits(pars []types.Type) int {
	mem := winTotalMemSizeBits(pars)

	if len(pars) <= 4 {
		return mem + 32*8
	}

	return mem + (len(pars)-4)*64 + 32*8
}

var (
	winIntRegs = []string{"rcx", "rdx", "r8", "r9"}
	winSseRegs = []string{"xmm0", "xmm1", "xmm2", "xmm3"}
)

func (v Target) winLoadCallArgument(a *allocator, arg ssa.Value, i int) {
	reg := i < 4

	if !reg {
		v.wop("subq $8, #rsp")
	}

	if winIsMemory(arg.Type()) {
		storesz := TypeStoreSizeInBits(arg.Type())
		v.moveValToMem(a, arg, "r11", 0)

		if reg {
			v.wop("movq #r11, #%s", winIntRegs[i])
		} else {
			v.wop("movq #r11, #rsp")
		}

		addbits := storesz
		for addbits%(16*8) != 0 { // TODO do this better
			addbits++
		}
		v.wop("addq $%d, #r11", addbits/8)

	} else if winIsInteger(arg.Type()) {
		if reg {
			v.moveIntToReg(a, arg, winIntRegs[i])
		} else {
			v.moveValToMem(a, arg, "rsp", 0)
		}
	} else if winIsFloat(arg.Type()) {
		if reg {
			v.moveIntToReg(a, arg, winIntRegs[i])
			v.moveFloatToSSEReg(a, arg, winSseRegs[i])
		} else {
			v.moveValToMem(a, arg, "rsp", 0)
		}
	} else {
		panic(arg)
	}
}

func (v Target) winSaveFunctionParameter(a *allocator, par ssa.Value, i int) {
	reg := i < 4
	storesz := TypeStoreSizeInBits(par.Type())

	if winIsMemory(par.Type()) {
		if reg {
			v.moveMemToMem(winIntRegs[i], "rbp", 0, -a.valOffset(par), storesz/8)
		} else {
			v.wop("movq $%d(#rbp), #r11", 16+i*8)
			v.moveMemToMem("r11", "rbp", 0, -a.valOffset(par), storesz/8)
		}
	} else if winIsInteger(par.Type()) {
		if reg {
			v.wop("mov%s #%s, %d(#rbp)", sizeSuffixBits(storesz), regToSize(winIntRegs[i], storesz), -a.valOffset(par))
		} else {
			v.wop("movq $%d(#rbp), #r11", 16+i*8)
			v.wop("mov%s #%s, %d(#rbp)", sizeSuffixBits(storesz), regToSize("r11", storesz), -a.valOffset(par))
		}
	} else if winIsFloat(par.Type()) {
		if reg {
			v.moveSSERegToFloat(a, winSseRegs[i], par)
		} else {
			//v.moveValToMem(a, arg, "rsp", 0)
			panic("unim")
		}
	} else {
		panic("unim")
	}
}

func (v Target) winLoadCallArguments(a *allocator, vals []ssa.Value, sig *types.Signature) {
	numVals := len(vals)

	v.wop("andq $-16, #rsp")

	if numVals > 4 && numVals%2 != 0 {
		v.wop("pushq $0")
	}

	totalMem := winTotalMemSizeBits(sig.Parameters())
	if totalMem > 0 {
		v.wop("subq $%d, #rsp", totalMem/8)
	}
	v.wop("movq #rsp, #r11")

	for i, arg := range vals {
		v.winLoadCallArgument(a, arg, i)
	}

	v.wop("subq $32, #rsp")
}

func (v Target) winSaveFunctionParameters(a *allocator, vals []ssa.Value, sig *types.Signature) {
	for i, par := range vals {
		v.winSaveFunctionParameter(a, par, i)
	}
}

// non-float types of sz <= 64, where sz is pow 2 are returned in rax
// float types are returned in xmm0
// everything else is returned via memory, location is returned in rax
// assumes memory return location is saved in %r15
func (v Target) winCopyReturnValue(a *allocator, val ssa.Value, saving bool) {
	if _, ok := val.Type().(types.Void); ok {
		return
	}

	storesz := TypeStoreSizeInBits(val.Type())

	if _, ok := val.Type().(*types.Float); ok {
		if saving {
			v.moveSSERegToFloat(a, "xmm0", val)
		} else {
			v.moveFloatToSSEReg(a, val, "xmm0")
		}
		return
	}

	if storesz <= 64 && isPow2(storesz) {
		if saving {
			v.wop("mov%s #%s, %s", sizeSuffixBits(storesz), regToSize("rax", storesz), a.valStr(val))
		} else {
			v.moveIntToReg(a, val, "rax")
		}
		return
	}

	if saving {
		v.moveMemToMem("rax", "rbp", 0, -a.valOffset(val), storesz/8)
	} else {
		v.wop("movq #r15, #rax")
		v.moveValToMem(a, val, "rax", 0)
	}
}
