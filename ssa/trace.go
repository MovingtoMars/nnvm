package ssa

import "fmt"

func FunctionTrace(function *Function) string {
	return function.SignatureString()
}

func BlockTrace(block *Block) string {
	return fmt.Sprintf("`%s` ; func @%s", block.Name(), block.Function().Name())
}

func InstrTrace(instr Instruction) string {
	return fmt.Sprintf("`%s` ; instr index %d, block %%%s, func @%s", instr.String(), instr.Block().InstrIndex(instr), instr.Block().Name(), instr.Block().Function().Name())
}

func GlobalTrace(global *Global) string {
	return fmt.Sprintf("`%s`", global.String())
}

func InstrName(instr Instruction) string {
	i := 0
	r := ' '

	str := instr.String()
	for i, r = range str {
		if r == ' ' {
			break
		}
	}

	return str[:i]
}
