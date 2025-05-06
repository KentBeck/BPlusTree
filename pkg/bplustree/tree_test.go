package bplustree

import (
	"testing"
)

func TestBPlusTreeInsert(t *testing.T) {
	tree := NewBPlusTree(4) // Branching factor of 4
	
	// Insert some keys
	if !tree.Insert(10) {
		t.Errorf("Failed to insert key 10")
	}
	if !tree.Insert(20) {
		t.Errorf("Failed to insert key 20")
	}
	if !tree.Insert(5) {
		t.Errorf("Failed to insert key 5")
	}
	
	// Try to insert a duplicate
	if tree.Insert(10) {
		t.Errorf("Inserted duplicate key 10")
	}
	
	// Check size
	if tree.Size() != 3 {
		t.Errorf("Expected size 3, got %d", tree.Size())
	}
}

func TestBPlusTreeContains(t *testing.T) {
	tree := NewBPlusTree(4)
	
	// Insert some keys
	tree.Insert(10)
	tree.Insert(20)
	tree.Insert(5)
	
	// Check contains
	if !tree.Contains(10) {
		t.Errorf("Expected to contain key 10")
	}
	if !tree.Contains(20) {
		t.Errorf("Expected to contain key 20")
	}
	if !tree.Contains(5) {
		t.Errorf("Expected to contain key 5")
	}
	if tree.Contains(15) {
		t.Errorf("Expected not to contain key 15")
	}
}

func TestBPlusTreeDelete(t *testing.T) {
	tree := NewBPlusTree(4)
	
	// Insert some keys
	tree.Insert(10)
	tree.Insert(20)
	tree.Insert(5)
	
	// Delete a key
	if !tree.Delete(10) {
		t.Errorf("Failed to delete key 10")
	}
	
	// Check contains
	if tree.Contains(10) {
		t.Errorf("Expected not to contain key 10 after deletion")
	}
	
	// Try to delete a non-existent key
	if tree.Delete(15) {
		t.Errorf("Deleted non-existent key 15")
	}
	
	// Check size
	if tree.Size() != 2 {
		t.Errorf("Expected size 2, got %d", tree.Size())
	}
}

func TestBPlusTreeSplitting(t *testing.T) {
	tree := NewBPlusTree(3) // Small branching factor to force splits
	
	// Insert keys to force splitting
	for i := 1; i <= 10; i++ {
		tree.Insert(uint64(i))
	}
	
	// Check that all keys are present
	for i := 1; i <= 10; i++ {
		if !tree.Contains(uint64(i)) {
			t.Errorf("Expected to contain key %d after splitting", i)
		}
	}
	
	// Check size
	if tree.Size() != 10 {
		t.Errorf("Expected size 10, got %d", tree.Size())
	}
	
	// Check height (should be greater than 1 due to splitting)
	if tree.Height() <= 1 {
		t.Errorf("Expected height > 1, got %d", tree.Height())
	}
}
