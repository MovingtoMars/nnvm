package ssa

import (
	"bufio"
	"bytes"

	"github.com/MovingtoMars/nnvm/types"
)

type Module struct {
	name string

	functions []*Function
	globals   []*Global
}

func NewModule(name string) *Module {
	return &Module{
		name: name,
	}
}

func (v Module) Name() string {
	return v.name
}

func (v *Module) NewGlobal(typ types.Type, init Initialiser, name string) *Global {
	glob := newGlobal(typ, init, name)
	v.globals = append(v.globals, glob)
	return glob
}

// Convenience method
func (v *Module) NewGlobalString(val string, nullTerminate bool, name string) *Global {
	strLit := NewStringLiteral(val, nullTerminate)
	return v.NewGlobal(strLit.Type(), NewLiteralInitialiser(strLit), name)
}

func (v Module) Globals() []*Global {
	return v.globals
}

func (v *Module) UpdateNames() {
	nameMap := make(map[string]int)
	nameMap[""] = 1

	for _, glob := range v.globals {
		updateValueName(nameMap, glob)
	}

	for _, fn := range v.functions {
		fn.UpdateNames()
	}
}

// Automatically calls UpdateNames
func (v *Module) String() string {
	v.UpdateNames()

	bytesBuf := bytes.NewBuffer(make([]byte, 0, 12+len(v.globals)*16+len(v.functions)*64))
	buf := bufio.NewWriter(bytesBuf)

	buf.WriteString("; Module '")
	buf.WriteString(v.name)
	buf.WriteString("'\n")

	for _, glob := range v.globals {
		buf.WriteString(glob.String())
		buf.WriteByte('\n')
	}

	for _, fn := range v.functions {
		buf.WriteByte('\n')
		fn.string(buf)
	}

	buf.Flush()
	return bytesBuf.String()
}

func (v Module) FunctionNamed(name string) *Function {
	for _, fn := range v.functions {
		if fn.Name() == name {
			return fn
		}
	}

	return nil
}

func (v Module) GlobalNamed(name string) *Global {
	for _, glob := range v.globals {
		if glob.Name() == name {
			return glob
		}
	}

	return nil
}

func (v Module) Functions() []*Function {
	return v.functions
}

// Created a new function with the specified type and name then inserts it into the module, returning the new function afterwards.
// If a function with the same name already exists, does nothing and returns nil.
func (v *Module) NewFunction(typ *types.Signature, name string) *Function {
	if v.FunctionNamed(name) != nil {
		return nil
	}

	fn := newFunction(typ, name)
	v.functions = append(v.functions, fn)

	return fn
}

// Automatically called by validate.Validate()
/*func (v *Module) UpdateValueReferences() {
	for _, fn := range v.functions {
		for _, block := range fn.blocks {
			for _, instr := range block.instrs {
				for _, op := range instr.operands() {
					(*op).clearReferences()
				}
			}
		}
	}

	for _, fn := range v.functions {
		for _, block := range fn.blocks {
			for _, instr := range block.instrs {
				for _, op := range instr.operands() {
					(*op).addReference(instr)
				}
			}
		}
	}
}*/
