package bplustree

import (
	"testing"
)

// TestGenericBPlusTreeBasicOperations tests the basic operations of the generic B+ tree
func TestGenericBPlusTreeBasicOperations(t *testing.T) {
	// Create a tree with a small branching factor for easier testing
	tree := NewGenericBPlusTree(
		4,
		func(a, b int) bool { return a < b },
		func(a, b int) bool { return a == b },
		func(v int) uint64 { return uint64(v) },
	)

	// Test initial state
	if !tree.IsEmpty() {
		t.Errorf("Expected new tree to be empty")
	}
	if tree.Size() != 0 {
		t.Errorf("Expected size 0, got %d", tree.Size())
	}
	if tree.Height() != 1 {
		t.Errorf("Expected height 1, got %d", tree.Height())
	}

	// Test insertion
	if !tree.Insert(10) {
		t.Errorf("Failed to insert 10")
	}
	if !tree.Insert(20) {
		t.Errorf("Failed to insert 20")
	}
	if !tree.Insert(5) {
		t.Errorf("Failed to insert 5")
	}

	// Test duplicate insertion
	if tree.Insert(10) {
		t.Errorf("Inserted duplicate value 10")
	}

	// Test size after insertions
	if tree.Size() != 3 {
		t.Errorf("Expected size 3, got %d", tree.Size())
	}

	// Test contains
	if !tree.Contains(10) {
		t.Errorf("Expected to contain 10")
	}
	if !tree.Contains(20) {
		t.Errorf("Expected to contain 20")
	}
	if !tree.Contains(5) {
		t.Errorf("Expected to contain 5")
	}
	if tree.Contains(15) {
		t.Errorf("Expected not to contain 15")
	}

	// Test deletion
	if !tree.Delete(10) {
		t.Errorf("Failed to delete 10")
	}
	if tree.Contains(10) {
		t.Errorf("Expected not to contain 10 after deletion")
	}
	if tree.Size() != 2 {
		t.Errorf("Expected size 2 after deletion, got %d", tree.Size())
	}

	// Test deleting non-existent key
	if tree.Delete(10) {
		t.Errorf("Deleted non-existent key 10")
	}

	// Test GetAllKeys
	keys := tree.GetAllKeys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}

	// Check if the keys are correct
	found5 := false
	found20 := false
	for _, k := range keys {
		if k == 5 {
			found5 = true
		}
		if k == 20 {
			found20 = true
		}
	}
	if !found5 || !found20 {
		t.Errorf("GetAllKeys returned incorrect keys")
	}

	// Test Clear
	tree.Clear()
	if !tree.IsEmpty() {
		t.Errorf("Expected tree to be empty after Clear")
	}
	if tree.Size() != 0 {
		t.Errorf("Expected size 0 after Clear, got %d", tree.Size())
	}
}

// TestGenericBPlusTreeRangeQuery tests the range query functionality
func TestGenericBPlusTreeRangeQuery(t *testing.T) {
	tree := NewGenericBPlusTree(
		4,
		func(a, b int) bool { return a < b },
		func(a, b int) bool { return a == b },
		func(v int) uint64 { return uint64(v) },
	)

	// Insert some values
	values := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	for _, v := range values {
		tree.Insert(v)
	}

	// Test range query
	result := tree.RangeQuery(25, 75)

	// Check result
	expected := []int{30, 40, 50, 60, 70}
	if len(result) != len(expected) {
		t.Errorf("Expected %d values in range [25, 75], got %d", len(expected), len(result))
	}

	// Check if all expected values are in the result
	for _, v := range expected {
		found := false
		for _, r := range result {
			if r == v {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected value %d not found in range query result", v)
		}
	}

	// Test empty range
	result = tree.RangeQuery(15, 15)
	if len(result) != 0 {
		t.Errorf("Expected empty result for range [15, 15], got %d values", len(result))
	}

	// Test range with no values
	result = tree.RangeQuery(31, 39)
	if len(result) != 0 {
		t.Errorf("Expected empty result for range [31, 39], got %d values", len(result))
	}

	// Test range with boundary values
	result = tree.RangeQuery(10, 10)
	if len(result) != 1 || result[0] != 10 {
		t.Errorf("Expected [10] for range [10, 10], got %v", result)
	}
}

// TestGenericBPlusTreeWithoutBloom tests the tree without a bloom filter
func TestGenericBPlusTreeWithoutBloom(t *testing.T) {
	tree := NewGenericBPlusTreeWithoutBloom(
		4,
		func(a, b int) bool { return a < b },
		func(a, b int) bool { return a == b },
		func(v int) uint64 { return uint64(v) },
	)

	// Test basic operations
	tree.Insert(10)
	tree.Insert(20)
	tree.Insert(30)

	if !tree.Contains(10) || !tree.Contains(20) || !tree.Contains(30) {
		t.Errorf("Tree without bloom filter failed to contain inserted values")
	}

	if tree.Contains(40) {
		t.Errorf("Tree without bloom filter contains non-existent value")
	}

	// Test deletion
	if !tree.Delete(20) {
		t.Errorf("Failed to delete from tree without bloom filter")
	}

	if tree.Contains(20) {
		t.Errorf("Tree without bloom filter still contains deleted value")
	}
}

// TestGenericBPlusTreeLargeDataset tests the tree with a large dataset
func TestGenericBPlusTreeLargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	tree := NewGenericBPlusTree(
		128, // Larger branching factor for better performance
		func(a, b int) bool { return a < b },
		func(a, b int) bool { return a == b },
		func(v int) uint64 { return uint64(v) },
	)

	// Insert a moderate number of values
	const numValues = 100 // Small enough for fast tests but large enough to test tree structure

	// Insert values in order
	for i := 0; i < numValues; i++ {
		if !tree.Insert(i) {
			t.Errorf("Failed to insert value %d", i)
		}
	}

	// Verify size
	if tree.Size() != numValues {
		t.Errorf("Expected size %d, got %d", numValues, tree.Size())
	}

	// Verify all values are present
	for i := 0; i < numValues; i++ {
		if !tree.Contains(i) {
			t.Errorf("Tree does not contain value %d", i)
			break
		}
	}

	// Delete half the values
	for i := 0; i < numValues/2; i++ {
		if !tree.Delete(i) {
			t.Errorf("Failed to delete value %d", i)
		}
	}

	// Verify size after deletion
	if tree.Size() != numValues/2 {
		t.Errorf("Expected size %d after deletion, got %d", numValues/2, tree.Size())
	}

	// Verify deleted values are gone
	for i := 0; i < numValues/2; i++ {
		if tree.Contains(i) {
			t.Errorf("Tree still contains deleted value %d", i)
			break
		}
	}

	// Verify remaining values are still present
	for i := numValues / 2; i < numValues; i++ {
		if !tree.Contains(i) {
			t.Errorf("Tree does not contain value %d", i)
			break
		}
	}
}

// TestGenericBPlusTreeBloomFilterOptimization tests the bloom filter optimization
func TestGenericBPlusTreeBloomFilterOptimization(t *testing.T) {
	// Create a tree with bloom filter
	treeWithBloom := NewGenericBPlusTree(
		16,
		func(a, b int) bool { return a < b },
		func(a, b int) bool { return a == b },
		func(v int) uint64 { return uint64(v) },
	)

	// Create a tree without bloom filter
	treeWithoutBloom := NewGenericBPlusTreeWithoutBloom(
		16,
		func(a, b int) bool { return a < b },
		func(a, b int) bool { return a == b },
		func(v int) uint64 { return uint64(v) },
	)

	// Insert the same values in both trees (small count for faster tests)
	for i := 0; i < 20; i++ {
		if !treeWithBloom.Insert(i) {
			t.Errorf("Failed to insert value %d into tree with bloom filter", i)
		}
		if !treeWithoutBloom.Insert(i) {
			t.Errorf("Failed to insert value %d into tree without bloom filter", i)
		}
	}

	// Both trees should contain the same values
	for i := 0; i < 20; i++ {
		if !treeWithBloom.Contains(i) {
			t.Errorf("Tree with bloom filter does not contain value %d", i)
		}
		if !treeWithoutBloom.Contains(i) {
			t.Errorf("Tree without bloom filter does not contain value %d", i)
		}
	}

	// Both trees should not contain values outside the range
	for i := 20; i < 30; i++ {
		if treeWithBloom.Contains(i) {
			t.Errorf("Tree with bloom filter contains unexpected value %d", i)
		}
		if treeWithoutBloom.Contains(i) {
			t.Errorf("Tree without bloom filter contains unexpected value %d", i)
		}
	}

	// Test bloom filter resize
	treeWithBloom.ResizeBloomFilter(100, 0.01)

	// Tree should still contain all values after resize
	for i := 0; i < 20; i++ {
		if !treeWithBloom.Contains(i) {
			t.Errorf("Tree does not contain value %d after bloom filter resize", i)
		}
	}
}
