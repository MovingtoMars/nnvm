package analysis

import (
	"fmt"
	"strings"

	"github.com/MovingtoMars/nnvm/ssa"
)

type DominatorTree struct {
	nodes []*DominatorTreeNode
	cfg   *CFG
}

func (v DominatorTree) BlockCFG() *CFG {
	return v.cfg
}

func (v DominatorTree) Nodes() []*DominatorTreeNode {
	return v.nodes
}

// return nil if unfound
func (v DominatorTree) NodeForBlock(b *ssa.Block) *DominatorTreeNode {
	for _, node := range v.nodes {
		if node.block == b {
			return node
		}
	}
	return nil
}

func (v DominatorTree) String() string {
	str := "BlockDominatorTree\n[Node: Children]\n"

	for _, node := range v.nodes {
		str += node.block.Name() + ": "
		for i, child := range node.children {
			str += child.block.Name()

			if i < len(node.children)-1 {
				str += ", "
			}
		}
		str += "\n"
	}

	return str
}

// Warning: will overwrite images.
// Always uses svg.
// Requires `dot` command from the graphviz package be available.
func (v DominatorTree) SaveImage(filename string) error {
	names := make(map[*DominatorTreeNode]string, len(v.nodes))
	i := 0
	for _, node := range v.nodes {
		names[node] = fmt.Sprintf("name%d", i)
		i++
	}

	contents := ""
	for _, node := range v.nodes {
		contents += "  " + names[node] + " -> {"
		for _, next := range node.children {
			contents += " " + names[next]
		}
		contents += " };\n"
	}

	for node, name := range names {
		contents += fmt.Sprintf("  %s [label=\"%s\"];\n", name, strings.Replace(node.String(), "\"", "", -1))
	}

	return saveDotImage(filename, contents)
}

// "strict" refers to strict dominators.
type DominatorTreeNode struct {
	parent   *DominatorTreeNode // nil for start node
	children []*DominatorTreeNode
	block    *ssa.Block
}

func (v DominatorTreeNode) String() string {
	return v.block.Name()
}

func NewBlockDominatorTree(cfg *CFG) *DominatorTree {
	v := &DominatorTree{}
	v.construct(cfg)
	v.cfg = cfg
	return v
}

// Returns nil for entry node
func (v DominatorTreeNode) ImmediateDominator() *DominatorTreeNode {
	return v.parent
}

func (v *DominatorTreeNode) DominatedBy(node *DominatorTreeNode, strict bool) bool {
	if !strict && v == node {
		return true
	}
	for dom := v.parent; dom != nil; dom = dom.parent {
		if dom == node {
			return true
		}
	}

	return false
}

func (v *DominatorTreeNode) Dominators(strict bool) []*DominatorTreeNode {
	doms := make([]*DominatorTreeNode, 0, 8)
	var dom *DominatorTreeNode
	if strict {
		dom = v.parent
	} else {
		dom = v
	}
	for ; dom != nil; dom = dom.parent {
		doms = append(doms, dom)
	}
	return doms
}

// Returns true if every node from entry to dominatee contains dominator.
// Non-strict
func blockDominatesBlock(dominator, dominatee *CFGNode) bool {
	if dominator == dominatee {
		return true
	}

	if len(dominatee.Prev()) == 0 {
		return false
	}

	for _, prev := range dominatee.Prev() {
		if prev == dominatee {
			continue
		}
		if dom := blockDominatesBlock(dominator, prev); !dom {
			return false
		}
	}

	return true
}

// slow, TODO use faster algorithm
func (v *DominatorTree) construct(cfg *CFG) {
	if len(cfg.Nodes()) == 0 {
		return
	}

	blockDomsMap := make(map[*ssa.Block][]*ssa.Block, len(cfg.Nodes()))
	blockNodeMap := make(map[*ssa.Block]*DominatorTreeNode)

	for _, node := range cfg.Nodes() {
		doms := []*ssa.Block{}

		for _, otherNode := range cfg.Nodes() {
			if blockDominatesBlock(otherNode, node) {
				doms = append(doms, otherNode.block)
			}
		}

		blockDomsMap[node.block] = doms
	}

	for _, node := range cfg.Nodes() {
		domNode := &DominatorTreeNode{
			block: node.block,
		}
		blockNodeMap[node.block] = domNode
		v.nodes = append(v.nodes, domNode)
	}

	removeBlockDom := func(node, removeDom *ssa.Block) bool {
		doms := blockDomsMap[node]
		for i, dom := range doms {
			if dom == removeDom {
				copy(doms[i:], doms[i+1:])
				doms = doms[:len(doms)-1]
				blockDomsMap[node] = doms
				return true
			}
		}

		return false
	}

	var do func(*ssa.Block)
	do = func(root *ssa.Block) {
		for _, node := range cfg.Nodes() {
			if removeBlockDom(node.block, root) && len(blockDomsMap[node.block]) == 1 {
				blockNodeMap[root].children = append(blockNodeMap[root].children, blockNodeMap[node.block])
				blockNodeMap[node.block].parent = blockNodeMap[root]
			}
		}

		for _, node := range cfg.Nodes() {
			block := node.block
			if len(blockDomsMap[node.block]) == 1 {
				do(block)
			}
		}
	}

	do(cfg.Nodes()[0].block)
}
