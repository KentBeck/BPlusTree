package bplustree

import (
	"fmt"
	"strings"
)

// PrintTree prints a visual representation of the tree
func (t *BPlusTree) PrintTree() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Tree(size=%d, height=%d, branching=%d)\n", t.size, t.height, t.branchingFactor))
	t.printNode(&sb, t.root, 0)
	return sb.String()
}

// printNode recursively prints a node and its children
func (t *BPlusTree) printNode(sb *strings.Builder, node Node, level int) {
	indent := strings.Repeat("  ", level)
	
	switch n := node.(type) {
	case *LeafNodeImpl:
		sb.WriteString(fmt.Sprintf("%sLeaf: %v\n", indent, n.Keys()))
	case *InternalNodeImpl:
		sb.WriteString(fmt.Sprintf("%sInternal: %v\n", indent, n.Keys()))
		for i, child := range n.Children() {
			if i > 0 {
				sb.WriteString(fmt.Sprintf("%s[Key: %d]\n", indent, n.Keys()[i-1]))
			}
			t.printNode(sb, child, level+1)
		}
	}
}
