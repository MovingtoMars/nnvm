package amd64

import (
	"fmt"

	"github.com/MovingtoMars/nnvm/ssa"
	"github.com/MovingtoMars/nnvm/types"
)

type allocator struct {
	stackSize  int               // positive value
	valOffsets map[ssa.Value]int // positive values
}

func newAllocator() *allocator {
	return &allocator{valOffsets: make(map[ssa.Value]int)}
}

func (v *allocator) allocateValue(val ssa.Value) {
	if _, ok := val.Type().(types.Void); ok {
		return
	}

	if _, ok := v.valOffsets[val]; ok {
		panic("duplicate val")
	}

	sz := TypeStoreSizeInBits(val.Type()) / 8
	align := TypeAlignmentInBits(val.Type()) / 8

	v.stackSize += sz
	for v.stackSize%align != 0 {
		v.stackSize++
	}
	v.valOffsets[val] = v.stackSize
}

func (v *allocator) allocate(fn *ssa.Function) {
	for _, par := range fn.Parameters() {
		v.allocateValue(par)
	}

	for _, block := range fn.Blocks() {
		for _, instr := range block.Instrs() {
			if val, ok := instr.(ssa.Value); ok {
				v.allocateValue(val)
			}
		}
	}
}

// TODO this really should return a negative value
func (v allocator) valOffset(val ssa.Value) int {
	off := v.valOffsets[val]
	if off == 0 {
		panic("no: " + val.Name())
	}
	return off
}

func (v allocator) valStr(val ssa.Value) string {
	if global, ok := val.(*ssa.Global); ok {
		return fmt.Sprintf("$" + global.Name())
	}

	return fmt.Sprintf("-%d(#rbp)", v.valOffset(val))
}
