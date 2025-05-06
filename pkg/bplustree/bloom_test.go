package bplustree

import (
	"testing"
)

// TestBloomFilterBasic tests the basic functionality of the Bloom filter
func TestBloomFilterBasic(t *testing.T) {
	bf := NewBloomFilter(1000, 3)

	// Add some keys
	bf.Add(10)
	bf.Add(20)
	bf.Add(30)

	// Check if keys exist
	if !bf.Contains(10) {
		t.Errorf("Bloom filter should contain key 10")
	}
	if !bf.Contains(20) {
		t.Errorf("Bloom filter should contain key 20")
	}
	if !bf.Contains(30) {
		t.Errorf("Bloom filter should contain key 30")
	}

	// Check a key that doesn't exist
	// Note: This might occasionally fail due to false positives
	if bf.Contains(40) {
		t.Logf("False positive detected for key 40 (this is expected occasionally)")
	}

	// Clear the filter
	bf.Clear()

	// Check that keys no longer exist
	if bf.Contains(10) {
		t.Errorf("Bloom filter should not contain key 10 after clearing")
	}
	if bf.Contains(20) {
		t.Errorf("Bloom filter should not contain key 20 after clearing")
	}
	if bf.Contains(30) {
		t.Errorf("Bloom filter should not contain key 30 after clearing")
	}
}

// TestBPlusTreeWithBloomFilter tests the B+ tree with the Bloom filter
func TestBPlusTreeWithBloomFilter(t *testing.T) {
	tree := NewBPlusTree(4)

	// Insert some keys
	tree.Insert(10)
	tree.Insert(20)
	tree.Insert(30)
	tree.Insert(40)
	tree.Insert(50)

	// Check if keys exist
	// This should compute the Bloom filter
	if !tree.Contains(10) {
		t.Errorf("Tree should contain key 10")
	}

	// Check if the Bloom filter is used for subsequent lookups
	// This is hard to test directly, but we can verify that the result is correct
	if !tree.Contains(20) {
		t.Errorf("Tree should contain key 20")
	}
	if !tree.Contains(30) {
		t.Errorf("Tree should contain key 30")
	}

	// Check a key that doesn't exist
	if tree.Contains(60) {
		t.Errorf("Tree should not contain key 60")
	}

	// Delete a key
	tree.Delete(20)

	// Check that the deleted key no longer exists
	if tree.Contains(20) {
		t.Errorf("Tree should not contain key 20 after deletion")
	}

	// Check that other keys still exist
	if !tree.Contains(10) {
		t.Errorf("Tree should still contain key 10")
	}
	if !tree.Contains(30) {
		t.Errorf("Tree should still contain key 30")
	}

	// Add a new key
	tree.Insert(60)

	// Check that the new key exists
	if !tree.Contains(60) {
		t.Errorf("Tree should contain key 60 after insertion")
	}
}

// TestBloomFilterAddDuringInsertion tests that keys are added to the Bloom filter during insertion
func TestBloomFilterAddDuringInsertion(t *testing.T) {
	tree := NewBPlusTree(4)

	// First, make sure the Bloom filter is valid by doing a lookup
	if tree.Contains(10) {
		t.Errorf("Empty tree should not contain key 10")
	}

	// Now insert a key - this should add it to the Bloom filter without invalidating
	tree.Insert(10)

	// Check that the key exists without having to recompute the Bloom filter
	if !tree.Contains(10) {
		t.Errorf("Tree should contain key 10 after insertion")
	}

	// Insert more keys
	tree.Insert(20)
	tree.Insert(30)

	// Check that all keys exist
	if !tree.Contains(10) {
		t.Errorf("Tree should contain key 10")
	}
	if !tree.Contains(20) {
		t.Errorf("Tree should contain key 20")
	}
	if !tree.Contains(30) {
		t.Errorf("Tree should contain key 30")
	}

	// Check a key that doesn't exist
	if tree.Contains(40) {
		t.Errorf("Tree should not contain key 40")
	}
}
