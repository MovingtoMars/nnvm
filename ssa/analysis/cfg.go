package analysis

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/MovingtoMars/nnvm/ssa"
)

type CFG struct {
	nodes []*CFGNode
}

func (v CFG) Nodes() []*CFGNode {
	return v.nodes
}

// nil if unfound
func (v CFG) NodeForBlock(b *ssa.Block) *CFGNode {
	for _, node := range v.nodes {
		if node.block == b {
			return node
		}
	}
	return nil
}

// Warning: will overwrite images.
// Always uses svg.
// Requires `dot` command from the graphviz package be available.
func (v CFG) SaveImage(filename string) error {
	names := make(map[*CFGNode]string, len(v.nodes))
	i := 0
	for _, node := range v.nodes {
		names[node] = fmt.Sprintf("name%d", i)
		i++
	}

	contents := ""
	for _, node := range v.nodes {
		contents += "  " + names[node] + " -> {"
		for _, next := range node.Next() {
			contents += " " + names[next]
		}
		contents += " };\n"
	}

	for node, name := range names {
		contents += fmt.Sprintf("  %s [label=\"%s\"];\n", name, strings.Replace(node.String(), "\"", "", -1))
	}

	return saveDotImage(filename, contents)
}

func saveDotImage(filename, graphContents string) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	dotCommand := exec.Command("dot", "-Tsvg")
	stdin, err := dotCommand.StdinPipe()
	if err != nil {
		return err
	}
	dotCommand.Stdout = file

	dotCommand.Start()
	fmt.Fprintf(stdin, "digraph {\n %s }\n", graphContents)
	stdin.Close()

	return dotCommand.Wait()
}

type CFGNode struct {
	block *ssa.Block
	prev  []*CFGNode
	next  []*CFGNode
}

func (v *CFGNode) Block() *ssa.Block {
	return v.block
}

func (v *CFG) Add(block *ssa.Block, prev, next []*CFGNode) {
	if prev == nil {
		panic("BlockCFG.Add: cannot add a new node with no previous nodes")
	}

	for _, node := range v.nodes {
		if node.block == block {
			panic("BlockCFG.Add: cannot add duplicate block")
		}
	}

	node := &CFGNode{block: block}

	for _, p := range prev {
		p.next = append(p.next, node)
	}

	for _, p := range next {
		p.prev = append(p.prev, node)
	}
}

func NewBlockCFG(fn *ssa.Function) *CFG {
	v := &CFG{}
	v.construct(fn)
	return v
}

func (v CFGNode) Next() []*CFGNode {
	return v.next
}

func (v CFGNode) Prev() []*CFGNode {
	return v.prev
}

func (v CFGNode) String() string {
	return v.block.Name()
}

func (v *CFG) construct(fn *ssa.Function) {
	blocks := fn.Blocks()
	nodes := make([]*CFGNode, len(blocks))
	blocksToNodes := make(map[*ssa.Block]*CFGNode)

	for i, block := range blocks {
		nodes[i] = &CFGNode{
			block: block,
		}
		blocksToNodes[block] = nodes[i]
	}

	for _, node := range nodes {
		switch term := node.block.LastInstr().(type) {
		case *ssa.Br:
			ops := ssa.GetOperands(term)
			node.next = []*CFGNode{
				blocksToNodes[ops[0].(*ssa.Block)],
			}

		case *ssa.CondBr:
			ops := ssa.GetOperands(term)
			node.next = []*CFGNode{
				blocksToNodes[ops[1].(*ssa.Block)],
				blocksToNodes[ops[2].(*ssa.Block)],
			}

		case *ssa.Unreachable, *ssa.Ret:
			// these lead to nowhere

		default:
			panic("unimplemented terminating instruction")
		}

		for _, next := range node.next {
			next.prev = append(next.prev, node)
		}
	}

	v.nodes = nodes
}
