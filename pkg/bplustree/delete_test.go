package bplustree

import (
	"testing"
)

// TestDeleteWithMergeLeafNodes tests deleting keys that causes leaf nodes to merge
func TestDeleteWithMergeLeafNodes(t *testing.T) {
	// Create a tree with small branching factor to force merges
	tree := NewBPlusTree(3) // Minimum keys in leaf: 1, Max: 3

	// Insert keys in a specific order to create a predictable structure
	keys := []uint64{10, 20, 30, 40, 50}
	for _, key := range keys {
		inserted := tree.Insert(key)
		t.Logf("Inserted key %d: %v", key, inserted)
	}

	t.Logf("Initial tree:\n%s", tree.PrintTree())

	// Delete keys to force leaf node merging
	t.Logf("Deleting key 20: %v", tree.Delete(20))
	t.Logf("After deleting 20:\n%s", tree.PrintTree())

	// Verify all remaining keys are still accessible
	for _, key := range []uint64{10, 30, 40, 50} {
		if !tree.Contains(key) {
			t.Errorf("Key %d should exist after merging", key)
		}
	}

	// Verify deleted keys are gone
	if tree.Contains(20) {
		t.Errorf("Key 20 should not exist after deletion")
	}

	// Verify tree size
	if tree.Size() != 4 {
		t.Errorf("Expected tree size 4, got %d", tree.Size())
	}
}

// TestDeleteWithRedistributeLeafNodes tests deleting keys that causes leaf nodes to redistribute
func TestDeleteWithRedistributeLeafNodes(t *testing.T) {
	// Create a tree with small branching factor
	tree := NewBPlusTree(3) // Minimum keys in leaf: 1, Max: 3

	// Insert keys to create a specific structure for redistribution
	keys := []uint64{10, 20, 30, 40, 50}
	for _, key := range keys {
		tree.Insert(key)
	}

	// Delete a key that should cause redistribution rather than merging
	tree.Delete(20)

	// Verify all remaining keys are still accessible
	for _, key := range []uint64{10, 30, 40, 50} {
		if !tree.Contains(key) {
			t.Errorf("Key %d should exist after redistribution", key)
		}
	}

	// Verify tree size
	if tree.Size() != 4 {
		t.Errorf("Expected tree size 4, got %d", tree.Size())
	}
}

// TestDeleteWithMergeInternalNodes tests deleting keys that causes internal nodes to merge
func TestDeleteWithMergeInternalNodes(t *testing.T) {
	// Create a tree with small branching factor
	tree := NewBPlusTree(3) // Minimum keys in internal node: 1, Max: 2

	// Insert keys in a specific order to create a predictable structure
	keys := []uint64{10, 20, 30, 40, 50}
	for _, key := range keys {
		inserted := tree.Insert(key)
		t.Logf("Inserted key %d: %v", key, inserted)
	}

	t.Logf("Initial tree:\n%s", tree.PrintTree())

	// Delete a key
	deleted := tree.Delete(20)
	t.Logf("Deleting key 20: %v", deleted)
	t.Logf("After deleting 20:\n%s", tree.PrintTree())

	// Verify all remaining keys are still accessible
	remainingKeys := []uint64{10, 30, 40, 50}
	for _, key := range remainingKeys {
		if !tree.Contains(key) {
			t.Errorf("Key %d should exist after deletion", key)
		}
	}

	// Verify deleted key is gone
	if tree.Contains(20) {
		t.Errorf("Key 20 should not exist after deletion")
	}

	// Verify tree size
	expectedSize := len(remainingKeys)
	if tree.Size() != expectedSize {
		t.Errorf("Expected tree size %d, got %d", expectedSize, tree.Size())
	}
}

// TestDeleteWithRedistributeInternalNodes tests deleting keys that causes internal nodes to redistribute
func TestDeleteWithRedistributeInternalNodes(t *testing.T) {
	// Create a tree with small branching factor
	tree := NewBPlusTree(3) // Minimum keys in internal node: 1, Max: 2

	// Insert keys to create a specific structure for redistribution
	keys := []uint64{10, 20, 30, 40, 50}
	for _, key := range keys {
		inserted := tree.Insert(key)
		t.Logf("Inserted key %d: %v", key, inserted)
	}

	t.Logf("Initial tree:\n%s", tree.PrintTree())

	// Delete a key
	deleted := tree.Delete(20)
	t.Logf("Deleting key 20: %v", deleted)
	t.Logf("After deleting 20:\n%s", tree.PrintTree())

	// Verify all remaining keys are still accessible
	remainingKeys := []uint64{10, 30, 40, 50}
	for _, key := range remainingKeys {
		if !tree.Contains(key) {
			t.Errorf("Key %d should exist after deletion", key)
		}
	}

	// Verify deleted key is gone
	if tree.Contains(20) {
		t.Errorf("Key 20 should not exist after deletion")
	}

	// Verify tree size
	expectedSize := len(remainingKeys)
	if tree.Size() != expectedSize {
		t.Errorf("Expected tree size %d, got %d", expectedSize, tree.Size())
	}
}

// TestDeleteRootCollapse tests that the tree height decreases when the root has only one child
func TestDeleteRootCollapse(t *testing.T) {
	// Create a tree with small branching factor
	tree := NewBPlusTree(3)

	// Insert keys to create a tree with height > 1
	keys := []uint64{10, 20, 30, 40, 50, 60}
	for _, key := range keys {
		tree.Insert(key)
	}

	initialHeight := tree.Height()
	if initialHeight < 2 {
		t.Fatalf("Expected initial tree height to be at least 2, got %d", initialHeight)
	}

	// Delete keys until the root should collapse
	for _, key := range []uint64{10, 30, 50} {
		tree.Delete(key)
	}

	// Verify the height decreased
	if tree.Height() >= initialHeight {
		t.Errorf("Expected tree height to decrease from %d, got %d", initialHeight, tree.Height())
	}

	// Verify remaining keys are still accessible
	for _, key := range []uint64{20, 40, 60} {
		if !tree.Contains(key) {
			t.Errorf("Key %d should exist after root collapse", key)
		}
	}
}

// TestDeleteEmptyTree tests deleting from an empty tree
func TestDeleteEmptyTree(t *testing.T) {
	tree := NewBPlusTree(3)

	// Try to delete from empty tree
	if tree.Delete(10) {
		t.Errorf("Delete should return false for empty tree")
	}

	// Verify size is still 0
	if tree.Size() != 0 {
		t.Errorf("Expected size 0, got %d", tree.Size())
	}
}
