package ssa

import "github.com/MovingtoMars/nnvm/types"

type insertPointType int

const (
	insertUndefined insertPointType = iota
	insertAfterInstr
	insertBeforeInstr
	insertBlockEnd
)

type Builder struct {
	insertType  insertPointType
	insertInstr Instruction
	insertBlock *Block
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (v *Builder) insert(i Instruction) {
	switch v.insertType {
	case insertAfterInstr:
		b := v.insertInstr.Block()
		b.instrs = append(b.instrs, nil)
		instrLoc := b.InstrIndex(v.insertInstr) + 1
		copy(b.instrs[instrLoc+1:], b.instrs[instrLoc:])
		b.instrs[instrLoc] = i

		v.insertInstr = i
	case insertBeforeInstr:
		b := v.insertInstr.Block()
		b.instrs = append(b.instrs, nil)
		instrLoc := b.InstrIndex(v.insertInstr)
		copy(b.instrs[instrLoc+1:], b.instrs[instrLoc:])
		b.instrs[instrLoc] = i

		v.insertInstr = i
	case insertBlockEnd:
		v.insertBlock.instrs = append(v.insertBlock.instrs, i)
	case insertUndefined:
		panic("undefined insert point")
	default:
		panic("erroneous insert type")
	}
}

func (v *Builder) SetInsertAfterInstr(i Instruction) {
	v.insertType = insertAfterInstr
	v.insertInstr = i
}

func (v *Builder) SetInsertBeforeInstr(i Instruction) {
	v.insertType = insertBeforeInstr
	v.insertInstr = i
}

func (v *Builder) SetInsertAtBlockEnd(b *Block) {
	v.insertType = insertBlockEnd
	v.insertBlock = b
}

func (v *Builder) SetInsertAtBlockStart(b *Block) {
	if b.NumInstrs() > 0 {
		v.SetInsertBeforeInstr(b.FirstInstr())
	}

	v.SetInsertAtBlockEnd(b)
}

func (v *Builder) currentBlock() *Block {
	switch v.insertType {
	case insertAfterInstr, insertBeforeInstr:
		return v.insertInstr.Block()
	case insertBlockEnd:
		return v.insertBlock
	case insertUndefined:
		return nil
	default:
		panic("erroneous insert type")
	}
}

func (v *Builder) setupInstr(i Instruction, name string) {
	if v.insertType == insertUndefined {
		panic("uninitialised builder")
	}

	i.setBlock(v.currentBlock())
	v.insert(i)

	for _, op := range i.operands() {
		if *op != nil {
			(*op).addReference(i)
		}
	}

	if name != "" {
		i.(Value).SetName(name)
	}
}

func (v *Builder) CreateRet(returnValue Value) *Ret {
	i := newRet(returnValue)
	v.setupInstr(i, "")
	return i
}

func (v *Builder) CreateBinOp(x, y Value, binOpType BinOpType, name string) *BinOp {
	i := newBinOp(x, y, binOpType)
	v.setupInstr(i, name)
	return i
}

func (v *Builder) CreateUnreachable() *Unreachable {
	i := newUnreachable()
	v.setupInstr(i, "")
	return i
}

func (v *Builder) CreateICmp(x, y Value, predicate IntPredicate, name string) *ICmp {
	i := newICmp(x, y, predicate)
	v.setupInstr(i, name)
	return i
}

func (v *Builder) CreateBr(target *Block) *Br {
	i := newBr(target)
	v.setupInstr(i, "")
	return i
}

func (v *Builder) CreateCondBr(cond Value, trueTarget, falseTarget *Block) *CondBr {
	i := newCondBr(cond, trueTarget, falseTarget)
	v.setupInstr(i, "")
	return i
}

func (v *Builder) CreateCall(fn *Function, args []Value, name string) *Call {
	i := newCall(fn, args)
	v.setupInstr(i, name)
	return i
}

func (v *Builder) CreateConvert(x Value, target types.Type, convertType ConvertType, name string) *Convert {
	i := newConvert(x, target, convertType)
	v.setupInstr(i, name)
	return i
}

func (v *Builder) CreateLoad(location Value, name string) *Load {
	i := newLoad(location)
	v.setupInstr(i, name)
	return i
}

func (v *Builder) CreateStore(location, value Value) *Store {
	i := newStore(location, value)
	v.setupInstr(i, "")
	return i
}

func (v *Builder) CreateAlloc(typ types.Type, name string) *Alloc {
	i := newAlloc(typ)
	v.setupInstr(i, name)
	return i
}

func (v *Builder) CreateGEP(value Value, indexes []Value, name string) *GEP {
	i := newGEP(value, indexes)
	v.setupInstr(i, name)
	return i
}

func (v *Builder) CreatePhi(t types.Type, name string) *Phi {
	i := newPhi(t)
	v.setupInstr(i, name)
	return i
}
