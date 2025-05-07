package bplustree

import (
	"testing"
)

// TestBranchNodeKeyCount tests the KeyCount method for branch nodes
func TestBranchNodeKeyCount(t *testing.T) {
	// Create a branch node
	node := NewBranch()

	// Empty node should have 0 keys
	if count := node.KeyCount(); count != 0 {
		t.Errorf("Expected empty branch node to have 0 keys, got %d", count)
	}

	// Add some keys with children
	node.keys = append(node.keys, 10, 20, 30)
	node.children = append(node.children, NewLeaf(), NewLeaf(), NewLeaf(), NewLeaf())

	// Check key count
	if count := node.KeyCount(); count != 3 {
		t.Errorf("Expected branch node to have 3 keys, got %d", count)
	}
}

// TestBranchNodeContains tests the Contains method for branch nodes
func TestBranchNodeContains(t *testing.T) {
	// Create a branch node
	node := NewBranch()

	// Empty node should not contain any key
	if node.Contains(10) {
		t.Errorf("Empty branch node should not contain key 10")
	}

	// Add some keys
	node.keys = append(node.keys, 10, 20, 30)
	node.children = append(node.children, NewLeaf(), NewLeaf(), NewLeaf(), NewLeaf())

	// Check contains
	if !node.Contains(10) {
		t.Errorf("Branch node should contain key 10")
	}
	if !node.Contains(20) {
		t.Errorf("Branch node should contain key 20")
	}
	if !node.Contains(30) {
		t.Errorf("Branch node should contain key 30")
	}
	if node.Contains(15) {
		t.Errorf("Branch node should not contain key 15")
	}
}

// TestBranchNodeInsertKey tests the InsertKey method for branch nodes
func TestBranchNodeInsertKey(t *testing.T) {
	// Create a branch node
	node := NewBranch()

	// InsertKey should return false for branch nodes (placeholder method)
	if node.InsertKey(10) {
		t.Errorf("InsertKey should return false for branch nodes")
	}
}

// TestLeafNodeKeyCount tests the KeyCount method for leaf nodes
func TestLeafNodeKeyCount(t *testing.T) {
	// Create a leaf node
	node := NewLeaf()

	// Empty node should have 0 keys
	if count := node.KeyCount(); count != 0 {
		t.Errorf("Expected empty leaf node to have 0 keys, got %d", count)
	}

	// Add some keys
	node.keys = append(node.keys, 10, 20, 30)

	// Check key count
	if count := node.KeyCount(); count != 3 {
		t.Errorf("Expected leaf node to have 3 keys, got %d", count)
	}
}

// TestLeafNodeNextAndSetNext tests the Next and SetNext methods for leaf nodes
func TestLeafNodeNextAndSetNext(t *testing.T) {
	// Create two leaf nodes
	node1 := NewLeaf()
	node2 := NewLeaf()

	// Initially, next should be nil
	if node1.Next() != nil {
		t.Errorf("Expected next to be nil initially")
	}

	// Set next
	node1.SetNext(node2)

	// Check next
	if node1.Next() != node2 {
		t.Errorf("Expected next to be node2")
	}
}

// TestBranchNodeBorrowFromLeft tests the BorrowFromLeft method for branch nodes
func TestBranchNodeBorrowFromLeft(t *testing.T) {
	// Create a parent node
	parent := NewBranch()

	// Create left and right siblings
	leftNode := NewBranch()
	rightNode := NewBranch()

	// Set up left node with keys and children
	leftNode.keys = append(leftNode.keys, 10, 20, 30)
	leftNode.children = append(leftNode.children, NewLeaf(), NewLeaf(), NewLeaf(), NewLeaf())

	// Set up right node with keys and children
	rightNode.keys = append(rightNode.keys, 50)
	rightNode.children = append(rightNode.children, NewLeaf(), NewLeaf())

	// Set up parent node
	parent.keys = append(parent.keys, 40)
	parent.children = append(parent.children, leftNode, rightNode)

	// Borrow from left
	rightNode.BorrowFromLeft(40, leftNode, 1, parent)

	// Verify right node now has the borrowed key
	if len(rightNode.keys) != 2 {
		t.Errorf("Expected right node to have 2 keys after borrowing, got %d", len(rightNode.keys))
	}

	// Verify the separator key was moved down
	if rightNode.keys[0] != 40 {
		t.Errorf("Expected first key in right node to be 40, got %d", rightNode.keys[0])
	}

	// Verify left node has one less key
	if len(leftNode.keys) != 2 {
		t.Errorf("Expected left node to have 2 keys after lending, got %d", len(leftNode.keys))
	}

	// Verify parent's separator key was updated
	if parent.keys[0] != 30 {
		t.Errorf("Expected parent's separator key to be 30, got %d", parent.keys[0])
	}
}

// TestLeafNodeBorrowFromLeft tests the BorrowFromLeft method for leaf nodes
func TestLeafNodeBorrowFromLeft(t *testing.T) {
	// Create a parent node
	parent := NewBranch()

	// Create left and right siblings
	leftNode := NewLeaf()
	rightNode := NewLeaf()

	// Set up left node with keys
	leftNode.keys = append(leftNode.keys, 10, 20, 30)

	// Set up right node with keys
	rightNode.keys = append(rightNode.keys, 50)

	// Set up parent node
	parent.keys = append(parent.keys, 40)
	parent.children = append(parent.children, leftNode, rightNode)

	// Borrow from left
	rightNode.BorrowFromLeft(leftNode)

	// Verify right node now has the borrowed key
	if len(rightNode.keys) != 2 {
		t.Errorf("Expected right node to have 2 keys after borrowing, got %d", len(rightNode.keys))
	}

	// Verify the borrowed key is at the beginning
	if rightNode.keys[0] != 30 {
		t.Errorf("Expected first key in right node to be 30, got %d", rightNode.keys[0])
	}

	// Verify left node has one less key
	if len(leftNode.keys) != 2 {
		t.Errorf("Expected left node to have 2 keys after lending, got %d", len(leftNode.keys))
	}

	// Update parent's separator key
	parent.keys[0] = rightNode.keys[0]

	// Verify parent's separator key was updated
	if parent.keys[0] != 30 {
		t.Errorf("Expected parent's separator key to be 30, got %d", parent.keys[0])
	}
}

// TestMergeWithRightInternal tests the mergeWithRightInternal method
func TestMergeWithRightInternal(t *testing.T) {
	// Create a tree with a small branching factor to force merges
	tree := NewBPlusTree(3)

	// Insert keys to create a structure with multiple levels
	for i := uint64(10); i <= 100; i += 10 {
		tree.Insert(i)
	}

	// Verify the tree has height > 1 (has internal nodes)
	if tree.Height() <= 1 {
		t.Fatalf("Expected tree to have height > 1, got %d", tree.Height())
	}

	// Create a test scenario where we need to merge with right internal node
	// This is a complex operation that requires setting up the right tree structure
	// For simplicity, we'll test the mergeWithRightInternal method directly

	// Create parent and two child nodes
	parent := NewBranch()
	leftChild := NewBranch()
	rightChild := NewBranch()

	// Set up left child
	leftChild.keys = append(leftChild.keys, 20)
	leftChild.children = append(leftChild.children, NewLeaf(), NewLeaf())

	// Set up right child
	rightChild.keys = append(rightChild.keys, 60)
	rightChild.children = append(rightChild.children, NewLeaf(), NewLeaf())

	// Set up parent
	parent.keys = append(parent.keys, 40)
	parent.children = append(parent.children, leftChild, rightChild)

	// Call the method directly
	if !tree.mergeWithRightInternal(leftChild, parent, 0) {
		t.Errorf("mergeWithRightInternal should return true")
	}

	// Verify left child now contains all keys
	if len(leftChild.keys) != 3 {
		t.Errorf("Expected left child to have 3 keys after merging, got %d", len(leftChild.keys))
	}

	// Verify the keys are in the correct order
	expectedKeys := []uint64{20, 40, 60}
	for i, key := range expectedKeys {
		if i < len(leftChild.keys) && leftChild.keys[i] != key {
			t.Errorf("Expected key at position %d to be %d, got %d", i, key, leftChild.keys[i])
		}
	}

	// Verify parent has removed the separator key and right child
	if len(parent.keys) != 0 {
		t.Errorf("Expected parent to have 0 keys after merging, got %d", len(parent.keys))
	}
	if len(parent.children) != 1 {
		t.Errorf("Expected parent to have 1 child after merging, got %d", len(parent.children))
	}
}
