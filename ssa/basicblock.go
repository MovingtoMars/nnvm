package ssa

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/MovingtoMars/nnvm/types"
)

type Block struct {
	ReferenceHandler
	NameHandler

	function *Function
	instrs   []Instruction
}

func newBlock(name string) *Block {
	return &Block{NameHandler: NameHandler{name: name}}
}

func (_ Block) Type() types.Type {
	return types.NewLabel()
}

func (v *Block) IsEntry() bool {
	return v == v.function.blocks[0]
}

func (v Block) String() string {
	bytesBuf := bytes.NewBuffer(nil)
	buf := bufio.NewWriter(bytesBuf)
	v.string(buf)
	buf.Flush()
	return bytesBuf.String()
}

func (v Block) string(out *bufio.Writer) {
	out.WriteString(v.name)
	out.WriteString(":")

	const minCommentCol = 40
	commentCol := utf8.RuneCountInString(v.name) + 2
	if commentCol < minCommentCol {
		commentCol = minCommentCol
	}

	out.WriteString(strings.Repeat(" ", commentCol-(utf8.RuneCountInString(v.name)+2)))

	refs := v.References()

	out.WriteString("; preds = ")
	wpred := false
	for _, ref := range refs {
		switch ref.(type) {
		case *Br, *CondBr:
			if wpred {
				out.WriteString(", ")
			}
			wpred = true

			out.WriteString(ref.Block().Name())
		}
	}

	out.WriteByte('\n')

	for _, instr := range v.instrs {
		out.WriteString("    ")

		value, ok := instr.(Value)
		if ok {
			out.WriteByte('%')
			out.WriteString(value.Name())
			out.WriteString(" = ")
		}

		out.WriteString(instr.String())

		writtenComment := false
		for _, op := range instr.operands() {
			if floatLit, ok := (*op).(*FloatLiteral); ok {
				if !writtenComment {
					writtenComment = true
					out.WriteString("     ; Float literals:")
				}

				out.WriteString(fmt.Sprintf(" %f", floatLit.Float64()))
			}
		}

		out.WriteByte('\n')
	}
}

func (v Block) InstrIndex(needle Instruction) int {
	for i, instr := range v.instrs {
		if instr == needle {
			return i
		}
	}

	return -1
}

func (v Block) NumInstrs() int {
	return len(v.instrs)
}

func (v Block) FirstInstr() Instruction {
	if len(v.instrs) > 0 {
		return v.instrs[0]
	}
	return nil
}

func (v Block) LastInstr() Instruction {
	if len(v.instrs) > 0 {
		return v.instrs[len(v.instrs)-1]
	}
	return nil
}

// Returns nil for index out of range.
func (v Block) InstrAtIndex(i int) Instruction {
	if i < 0 || i >= len(v.instrs) {
		return nil
	}
	return v.instrs[i]
}

func (v Block) Instrs() []Instruction {
	return v.instrs
}

func (v Block) Function() *Function {
	return v.function
}
