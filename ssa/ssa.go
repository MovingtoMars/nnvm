/*

Procedures required for working with SSA IR:
- finding all references to a specified value
- changing the value a reference refers to
- replacing all references to a value with another value

*/
package ssa

import "github.com/MovingtoMars/nnvm/types"

type Value interface {
	Type() types.Type
	Name() string // is empty for unset name
	SetName(string)

	// Duplicates may exist if the value is referenced more than once in an instruction.
	References() []Instruction

	// does not update the operand of the instruction
	addReference(Instruction)
	removeReference(Instruction)
	clearReferences()
}

type Instruction interface {
	// Returns the textual IR representation of the instruction.
	// If the instruction is also a value, the name of the value is not included in the string.
	String() string

	IsTerminating() bool

	Block() *Block
	setBlock(*Block)

	operands() []*Value
}

func ValueString(val Value) string {
	return val.Type().String() + " " + ValueIdentifier(val)
}

func ValueIdentifier(val Value) string {
	switch val.(type) {
	case *Global, *Function:
		return "@" + val.Name()
	case Literal:
		return val.Name()
	default:
		return "%" + val.Name()
	}
}

func valueListString(values []Value) string {
	str := ""
	for i, val := range values {
		str += ValueString(val)

		if i < len(values)-1 {
			str += ", "
		}
	}

	return str
}

func GetOperands(instr Instruction) []Value {
	var ops []Value

	for _, op := range instr.operands() {
		ops = append(ops, *op)
	}

	return ops
}

func ReplaceOperandFromValue(instr Instruction, value *Value, newValue Value) {
	(*value).removeReference(instr)
	(*value) = newValue
	(*value).addReference(instr)
}

func ReplaceOperandFromIndex(instr Instruction, opIndex int, newOp Value) {
	if newOp == nil {
		panic("ReplaceOperand: newOp == nil")
	}

	ops := instr.operands()
	if opIndex < 0 {
		panic("ReplaceOperand: opIndex < 0")
	} else if opIndex >= len(ops) {
		panic("ReplaceOperand: opIndex too high")
	}

	opPtr := ops[opIndex]
	ReplaceOperandFromValue(instr, opPtr, newOp)
}

// Make sure references are not out-of-date when calling this function.
func ReplaceAllValueReferences(original, replacement Value) {
	origRefs := original.References()
	for _, instr := range origRefs {
		for i, op := range instr.operands() {

			if *op == original {
				*op = replacement
				break
			}

			if i == len(instr.operands())-1 {
				panic("a value reference does not have the value as an operand")
			}
		}
	}
}
