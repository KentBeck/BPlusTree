package bplustree

import (
	"testing"
)

// TestNullBloomFilter tests the NullBloomFilter implementation
func TestNullBloomFilter(t *testing.T) {
	// Create a null bloom filter
	nullFilter := NewNullBloomFilter()

	// Test that it always returns true for Contains (meaning "maybe")
	if !nullFilter.Contains(123) {
		t.Errorf("NullBloomFilter.Contains should always return true")
	}

	// Test that it's valid by default
	if !nullFilter.IsValid() {
		t.Errorf("NullBloomFilter should be valid by default")
	}

	// Test that Clear and SetValid don't cause errors
	nullFilter.Clear()
	nullFilter.SetValid()

	// Test that Add doesn't cause errors
	nullFilter.Add(456)
}

// TestNewBPlusTreeWithOptions tests the NewBPlusTreeWithOptions constructor
func TestNewBPlusTreeWithOptions(t *testing.T) {
	// Create a tree without a bloom filter
	tree := NewBPlusTreeWithOptions(3, false)

	// Insert some keys
	tree.Insert(10)
	tree.Insert(20)
	tree.Insert(30)

	// Verify that the tree works correctly
	if !tree.Contains(10) {
		t.Errorf("Tree should contain key 10")
	}

	if !tree.Contains(20) {
		t.Errorf("Tree should contain key 20")
	}

	if !tree.Contains(30) {
		t.Errorf("Tree should contain key 30")
	}

	if tree.Contains(40) {
		t.Errorf("Tree should not contain key 40")
	}
}
