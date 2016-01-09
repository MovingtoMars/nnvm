package ssa

type ReferenceHandler struct {
	references []Instruction
}

func (v ReferenceHandler) References() []Instruction {
	return v.references
}

func (v *ReferenceHandler) addReference(ref Instruction) {
	v.references = append(v.references, ref)
}

func (v *ReferenceHandler) removeReference(ref Instruction) {
	loc := -1

	for i, eref := range v.references {
		if eref == ref {
			loc = i
			break
		}
	}

	if loc == -1 {
		panic("tried to remove non-existant reference")
	}

	copy(v.references[loc:], v.references[loc+1:])
	v.references = v.references[:len(v.references)-1]
}

func (v *ReferenceHandler) clearReferences() {
	v.references = nil
}

type NameHandler struct {
	name string
}

func (v NameHandler) Name() string {
	return v.name
}

func (v *NameHandler) SetName(name string) {
	v.name = name
}

type BlockHandler struct {
	block *Block
}

func (v BlockHandler) Block() *Block {
	return v.block
}

func (v *BlockHandler) setBlock(b *Block) {
	v.block = b
}
