package analysis

import (
	"fmt"
	"io"

	"github.com/MovingtoMars/nnvm/ssa"
)

func PrintValues(mod *ssa.Module, out io.Writer) {
	printf := func(val string, stuff ...interface{}) {
		fmt.Fprintf(out, val, stuff...)
	}

	printf("Values in module `%s`\n", mod.Name())
	printf("\n")

	printf("Globals:\n")
	for _, glob := range mod.Globals() {
		printf("  %s (references: %d)\n", ssa.ValueString(glob), len(glob.References()))
	}
	printf("\n")

	printf("Functions:\n")
	for _, fn := range mod.Functions() {
		printf("  %s (references: %d)\n", ssa.ValueString(fn), len(fn.References()))
		if fn.IsPrototype() {
			printf("\n")
			continue
		}

		printf("    Parameters:\n")

		for _, par := range fn.Parameters() {
			printf("      %s (references: %d)\n", ssa.ValueString(par), len(par.References()))
		}

		printf("    Instruction values:\n")

		for _, block := range fn.Blocks() {
			for _, instr := range block.Instrs() {
				if val, ok := instr.(ssa.Value); ok {
					printf("      %s (references: %d)\n", ssa.ValueString(val), len(val.References()))
				}
			}
		}

		printf("\n")
	}
}
