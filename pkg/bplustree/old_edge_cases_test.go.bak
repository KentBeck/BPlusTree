package bplustree

import (
	"testing"
)

// TestBranchNodeRemoveChild tests the RemoveChild method for branch nodes
func TestBranchNodeRemoveChild(t *testing.T) {
	// Create a branch node
	node := NewBranch()

	// Add some children
	child1 := NewLeaf()
	child2 := NewLeaf()
	child3 := NewLeaf()

	node.children = append(node.children, child1, child2, child3)
	node.keys = append(node.keys, 10, 20)

	// Remove the middle child
	node.RemoveChild(1)

	// Verify the child was removed
	if len(node.children) != 2 {
		t.Errorf("Expected 2 children after removal, got %d", len(node.children))
	}

	// Note: RemoveChild doesn't update keys, it only removes the child
	// In a real B+ tree, the keys would be updated by the balancing methods

	// Try to remove with invalid index (should be ignored)
	node.RemoveChild(5)

	// Verify nothing changed
	if len(node.children) != 2 {
		t.Errorf("Expected 2 children after invalid removal, got %d", len(node.children))
	}
}

// TestBranchNodeSetChild tests the SetChild method for branch nodes
func TestBranchNodeSetChild(t *testing.T) {
	// Create a branch node
	node := NewBranch()

	// Add some children
	child1 := NewLeaf()
	child2 := NewLeaf()

	node.children = append(node.children, child1)

	// Set a child at an existing index
	node.SetChild(0, child2)

	// Verify the child was set
	if node.children[0] != child2 {
		t.Errorf("Expected child at index 0 to be child2")
	}

	// Note: SetChild only sets a child at an existing index
	// It doesn't expand the slice, so we can't test setting at index 2
	// In a real B+ tree, the children array would be managed by other methods
}

// TestTreeEdgeCases tests edge cases in the tree implementation
func TestTreeEdgeCases(t *testing.T) {
	// We'll test edge cases that can be safely tested
	tree := NewBPlusTree(4)

	// Test findLeaf with a non-leaf, non-branch node (shouldn't happen in practice)
	// We can't directly test with nil as it would cause a panic

	// Test countKeysInNode with a non-leaf, non-branch node (shouldn't happen in practice)
	// We can simulate this by creating a custom node type that implements the Node interface

	// Test deleteAndBalance with edge cases
	deleted, _ := tree.deleteAndBalance(tree.root, nil, -1, 999) // Key doesn't exist
	if deleted {
		t.Errorf("Expected deleteAndBalance to return false for non-existent key")
	}
}

// TestTryBorrowFromLeftEdgeCases tests edge cases in the tryBorrowFromLeft method
func TestTryBorrowFromLeftEdgeCases(t *testing.T) {
	// Create a tree with a small branching factor
	tree := NewBPlusTree(3)

	// Insert some keys
	for i := uint64(10); i <= 50; i += 10 {
		tree.Insert(i)
	}

	// Create a test scenario for tryBorrowFromLeft
	parent := NewBranch()
	leftNode := NewBranch()
	rightNode := NewBranch()

	// Set up left node with minimum keys
	leftNode.keys = append(leftNode.keys, 10)
	leftNode.children = append(leftNode.children, NewLeaf(), NewLeaf())

	// Set up right node with minimum keys
	rightNode.keys = append(rightNode.keys, 30)
	rightNode.children = append(rightNode.children, NewLeaf(), NewLeaf())

	// Set up parent node
	parent.keys = append(parent.keys, 20)
	parent.children = append(parent.children, leftNode, rightNode)

	// Try to borrow from left when left node has minimum keys (should fail)
	if rightNode.tryBorrowFromLeft(parent, 1, 3) {
		t.Errorf("Expected tryBorrowFromLeft to return false when left node has minimum keys")
	}

	// Add more keys to left node
	leftNode.keys = append(leftNode.keys, 15)
	leftNode.children = append(leftNode.children, NewLeaf())

	// Try to borrow from left when left node has enough keys (should succeed)
	if !rightNode.tryBorrowFromLeft(parent, 1, 3) {
		t.Errorf("Expected tryBorrowFromLeft to return true when left node has enough keys")
	}

	// Verify right node now has the borrowed key
	if len(rightNode.keys) != 2 {
		t.Errorf("Expected right node to have 2 keys after borrowing, got %d", len(rightNode.keys))
	}
}

// TestTryBorrowFromRightEdgeCases tests edge cases in the tryBorrowFromRight method
func TestTryBorrowFromRightEdgeCases(t *testing.T) {
	// Create a tree with a small branching factor
	tree := NewBPlusTree(3)

	// Insert some keys
	for i := uint64(10); i <= 50; i += 10 {
		tree.Insert(i)
	}

	// Create a test scenario for tryBorrowFromRight
	parent := NewBranch()
	leftNode := NewBranch()
	rightNode := NewBranch()

	// Set up left node with minimum keys
	leftNode.keys = append(leftNode.keys, 10)
	leftNode.children = append(leftNode.children, NewLeaf(), NewLeaf())

	// Set up right node with minimum keys
	rightNode.keys = append(rightNode.keys, 30)
	rightNode.children = append(rightNode.children, NewLeaf(), NewLeaf())

	// Set up parent node
	parent.keys = append(parent.keys, 20)
	parent.children = append(parent.children, leftNode, rightNode)

	// Try to borrow from right when right node has minimum keys (should fail)
	if leftNode.tryBorrowFromRight(parent, 0, 3) {
		t.Errorf("Expected tryBorrowFromRight to return false when right node has minimum keys")
	}

	// Add more keys to right node
	rightNode.keys = append(rightNode.keys, 35)
	rightNode.children = append(rightNode.children, NewLeaf())

	// Try to borrow from right when right node has enough keys (should succeed)
	if !leftNode.tryBorrowFromRight(parent, 0, 3) {
		t.Errorf("Expected tryBorrowFromRight to return true when right node has enough keys")
	}

	// Verify left node now has the borrowed key
	if len(leftNode.keys) != 2 {
		t.Errorf("Expected left node to have 2 keys after borrowing, got %d", len(leftNode.keys))
	}
}
