package bplustree

import (
	"fmt"
	"strings"
)

// PrintTree prints a visual representation of the tree
func PrintTree[K comparable](t *GenericBPlusTree[K]) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Tree(size=%d, height=%d, branching=%d)\n", t.size, t.Height(), t.branchingFactor))
	printNode(&sb, t.root, 0)
	return sb.String()
}

// printNode recursively prints a node and its children
func printNode[K any](sb *strings.Builder, node GenericNode[K], level int) {
	indent := strings.Repeat("  ", level)

	switch n := node.(type) {
	case *GenericLeafNode[K]:
		sb.WriteString(fmt.Sprintf("%sLeaf: %v\n", indent, n.Keys()))
	case *GenericBranchNode[K]:
		sb.WriteString(fmt.Sprintf("%sInternal: %v\n", indent, n.Keys()))
		for i, child := range n.Children() {
			if i > 0 {
				sb.WriteString(fmt.Sprintf("%s[Key: %v]\n", indent, n.Keys()[i-1]))
			}
			printNode(sb, child, level+1)
		}
	}
}
