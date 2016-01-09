package ssa

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/MovingtoMars/nnvm/types"
)

type Function struct {
	NameHandler
	ReferenceHandler

	typ        *types.Signature
	parameters []*Parameter

	blocks []*Block
}

func newFunction(typ *types.Signature, name string) *Function {
	var parameters []*Parameter

	for _, parType := range typ.Parameters() {
		parameters = append(parameters, newParameter(parType, ""))
	}

	return &Function{
		typ:         typ,
		NameHandler: NameHandler{name: name},
		parameters:  parameters,
	}
}

func (v Function) IsPrototype() bool {
	return len(v.blocks) == 0
}

func (v Function) Parameters() []*Parameter {
	return v.parameters
}

func (v Function) Blocks() []*Block {
	return v.blocks
}

func (v Function) EntryBlock() *Block {
	if len(v.blocks) > 0 {
		return v.blocks[0]
	}

	return nil
}

func (v Function) Type() types.Type {
	return v.typ
}

func updateValueName(nameMap map[string]int, thing Value) {
	name := thing.Name()
	num := nameMap[name]
	if num > 0 {
		thing.SetName(fmt.Sprintf("%s%d", name, num))
	}

	nameMap[name]++
}

func (v *Function) UpdateNames() {
	nameMap := make(map[string]int)
	nameMap[""] = 1

	for _, par := range v.parameters {
		updateValueName(nameMap, par)
	}

	for _, block := range v.blocks {
		updateValueName(nameMap, block)
	}

	for _, block := range v.blocks {
		for _, instr := range block.Instrs() {
			if val, ok := instr.(Value); ok {
				updateValueName(nameMap, val)
			}
		}
	}
}

func (v Function) String() string {
	bytesBuf := bytes.NewBuffer(nil)
	buf := bufio.NewWriter(bytesBuf)
	v.string(buf)
	buf.Flush()
	return bytesBuf.String()
}

func (v Function) SignatureString() string {
	str := "func " + v.typ.ReturnType().String() + " @" + v.name + "("

	for i, par := range v.parameters {
		str += ValueString(par)

		if v.typ.Variadic() || i < len(v.Parameters())-1 {
			str += ", "
		}
	}

	if v.typ.Variadic() {
		str += "..."
	}

	str += ")"
	return str
}

func (v Function) string(out *bufio.Writer) {
	out.WriteString(v.SignatureString())

	if len(v.blocks) > 0 {
		out.WriteString(" {\n")
		for i, block := range v.blocks {
			block.string(out)

			if i < len(v.blocks)-1 {
				out.WriteByte('\n')
			}
		}

		out.WriteByte('}')
	}

	out.WriteByte('\n')
}

func (v *Function) AddBlockAtStart(name string) *Block {
	b := newBlock(name)
	b.function = v
	v.blocks = append([]*Block{b}, v.blocks...)
	return b
}

func (v *Function) AddBlockAtEnd(name string) *Block {
	b := newBlock(name)
	b.function = v
	v.blocks = append(v.blocks, b)
	return b
}
