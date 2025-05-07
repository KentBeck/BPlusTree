package bplustree

import (
	"testing"
)

// TestGenericBPlusTreeEmptyTree tests operations on an empty tree
func TestGenericBPlusTreeEmptyTree(t *testing.T) {
	tree := NewGenericBPlusTree(
		4,
		func(a, b int) bool { return a < b },
		func(a, b int) bool { return a == b },
		func(v int) uint64 { return uint64(v) },
	)

	// Test properties of an empty tree
	if !tree.IsEmpty() {
		t.Errorf("Expected new tree to be empty")
	}
	if tree.Size() != 0 {
		t.Errorf("Expected size 0, got %d", tree.Size())
	}
	if tree.Height() != 1 {
		t.Errorf("Expected height 1, got %d", tree.Height())
	}

	// Test operations on an empty tree
	if tree.Contains(10) {
		t.Errorf("Empty tree should not contain any values")
	}
	if tree.Delete(10) {
		t.Errorf("Delete on empty tree should return false")
	}

	// Test GetAllKeys on an empty tree
	keys := tree.GetAllKeys()
	if len(keys) != 0 {
		t.Errorf("Expected empty key list, got %d keys", len(keys))
	}

	// Test RangeQuery on an empty tree
	result := tree.RangeQuery(10, 20)
	if len(result) != 0 {
		t.Errorf("Expected empty range query result, got %d keys", len(result))
	}
}

// TestGenericBPlusTreeSingleElement tests operations on a tree with a single element
func TestGenericBPlusTreeSingleElement(t *testing.T) {
	tree := NewGenericBPlusTree(
		4,
		func(a, b int) bool { return a < b },
		func(a, b int) bool { return a == b },
		func(v int) uint64 { return uint64(v) },
	)

	// Insert a single element
	tree.Insert(10)

	// Test properties
	if tree.IsEmpty() {
		t.Errorf("Tree with one element should not be empty")
	}
	if tree.Size() != 1 {
		t.Errorf("Expected size 1, got %d", tree.Size())
	}
	if tree.Height() != 1 {
		t.Errorf("Expected height 1, got %d", tree.Height())
	}

	// Test Contains
	if !tree.Contains(10) {
		t.Errorf("Tree should contain the inserted value")
	}
	if tree.Contains(20) {
		t.Errorf("Tree should not contain values that weren't inserted")
	}

	// Test GetAllKeys
	keys := tree.GetAllKeys()
	if len(keys) != 1 || keys[0] != 10 {
		t.Errorf("Expected [10], got %v", keys)
	}

	// Test RangeQuery
	result := tree.RangeQuery(5, 15)
	if len(result) != 1 || result[0] != 10 {
		t.Errorf("Expected [10], got %v", result)
	}

	// Test Delete
	if !tree.Delete(10) {
		t.Errorf("Delete should return true for existing value")
	}
	if tree.Size() != 0 {
		t.Errorf("Expected size 0 after deletion, got %d", tree.Size())
	}
}

// TestGenericBPlusTreeSplitRoot tests the root splitting
func TestGenericBPlusTreeSplitRoot(t *testing.T) {
	tree := NewGenericBPlusTree(
		3, // Small branching factor to force splits
		func(a, b int) bool { return a < b },
		func(a, b int) bool { return a == b },
		func(v int) uint64 { return uint64(v) },
	)

	// Initial height should be 1
	if tree.Height() != 1 {
		t.Errorf("Expected initial height 1, got %d", tree.Height())
	}

	// Insert enough values to cause a root split
	if !tree.Insert(10) {
		t.Errorf("Failed to insert value 10")
	}
	if !tree.Insert(20) {
		t.Errorf("Failed to insert value 20")
	}
	if !tree.Insert(30) {
		t.Errorf("Failed to insert value 30")
	}

	// Insert more values to ensure we get a split
	if !tree.Insert(40) {
		t.Errorf("Failed to insert value 40")
	}
	if !tree.Insert(50) {
		t.Errorf("Failed to insert value 50")
	}

	// Check that all values are accessible
	for _, v := range []int{10, 20, 30, 40, 50} {
		if !tree.Contains(v) {
			t.Errorf("Tree should contain value %d after insertions", v)
		}
	}
}

// TestGenericBPlusTreeMultipleSplits tests multiple node splits
func TestGenericBPlusTreeMultipleSplits(t *testing.T) {
	tree := NewGenericBPlusTree(
		3, // Small branching factor to force splits
		func(a, b int) bool { return a < b },
		func(a, b int) bool { return a == b },
		func(v int) uint64 { return uint64(v) },
	)

	// Insert values in order to cause multiple splits
	values := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	for _, v := range values {
		if !tree.Insert(v) {
			t.Errorf("Failed to insert value %d", v)
		}
	}

	// Verify all values were inserted
	if tree.Size() != len(values) {
		t.Errorf("Expected size %d, got %d", len(values), tree.Size())
	}

	// Check that all values are accessible
	for _, v := range values {
		if !tree.Contains(v) {
			t.Errorf("Tree should contain %d after multiple splits", v)
		}
	}
}

// TestGenericBPlusTreeDeleteAndMerge tests deletion with node merging
func TestGenericBPlusTreeDeleteAndMerge(t *testing.T) {
	tree := NewGenericBPlusTree(
		3, // Small branching factor to force merges
		func(a, b int) bool { return a < b },
		func(a, b int) bool { return a == b },
		func(v int) uint64 { return uint64(v) },
	)

	// Insert a few values
	values := []int{10, 20, 30}
	for _, v := range values {
		if !tree.Insert(v) {
			t.Errorf("Failed to insert value %d", v)
		}
	}

	// Verify all values were inserted
	if tree.Size() != len(values) {
		t.Errorf("Expected size %d after insertions, got %d", len(values), tree.Size())
	}

	// Delete a value
	if !tree.Delete(20) {
		t.Errorf("Failed to delete value 20")
	}

	// Verify size after deletion
	if tree.Size() != len(values)-1 {
		t.Errorf("Expected size %d after deletion, got %d", len(values)-1, tree.Size())
	}

	// Verify the deleted value is gone
	if tree.Contains(20) {
		t.Errorf("Tree should not contain deleted value 20")
	}

	// Verify the remaining values are still accessible
	if !tree.Contains(10) {
		t.Errorf("Tree should still contain value 10")
	}
	if !tree.Contains(30) {
		t.Errorf("Tree should still contain value 30")
	}
}

// TestGenericBPlusTreeBorrowFromSibling tests borrowing during deletion
func TestGenericBPlusTreeBorrowFromSibling(t *testing.T) {
	tree := NewGenericBPlusTree(
		3, // Small branching factor to force borrowing
		func(a, b int) bool { return a < b },
		func(a, b int) bool { return a == b },
		func(v int) uint64 { return uint64(v) },
	)

	// Insert a few values
	values := []int{10, 20, 30}
	for _, v := range values {
		if !tree.Insert(v) {
			t.Errorf("Failed to insert value %d", v)
		}
	}

	// Verify all values were inserted
	for _, v := range values {
		if !tree.Contains(v) {
			t.Errorf("Tree should contain value %d after insertion", v)
		}
	}

	// Delete the middle value
	if !tree.Delete(20) {
		t.Errorf("Failed to delete value 20")
	}

	// Verify the deleted value is gone
	if tree.Contains(20) {
		t.Errorf("Tree should not contain deleted value 20")
	}

	// Verify the remaining values are still accessible
	if !tree.Contains(10) {
		t.Errorf("Tree should still contain value 10")
	}
	if !tree.Contains(30) {
		t.Errorf("Tree should still contain value 30")
	}
}

// TestGenericBPlusTreeStringType tests the tree with string keys
func TestGenericBPlusTreeStringType(t *testing.T) {
	tree := NewGenericBPlusTree(
		4,
		func(a, b string) bool { return a < b },
		func(a, b string) bool { return a == b },
		func(v string) uint64 {
			var hash uint64
			for i := 0; i < len(v); i++ {
				hash = hash*31 + uint64(v[i])
			}
			return hash
		},
	)

	// Insert string values
	tree.Insert("apple")
	tree.Insert("banana")
	tree.Insert("cherry")

	// Test Contains
	if !tree.Contains("apple") || !tree.Contains("banana") || !tree.Contains("cherry") {
		t.Errorf("Tree should contain all inserted string values")
	}
	if tree.Contains("date") {
		t.Errorf("Tree should not contain values that weren't inserted")
	}

	// Test Delete
	tree.Delete("banana")
	if tree.Contains("banana") {
		t.Errorf("Tree should not contain deleted string value")
	}

	// Test RangeQuery
	result := tree.RangeQuery("apple", "cherry")
	if len(result) != 2 || result[0] != "apple" || result[1] != "cherry" {
		t.Errorf("Expected [apple, cherry], got %v", result)
	}
}

// TestGenericBPlusTreeCustomType tests the tree with a custom struct type
func TestGenericBPlusTreeCustomType(t *testing.T) {
	type Person struct {
		ID   int
		Name string
	}

	tree := NewGenericBPlusTree(
		4,
		func(a, b Person) bool { return a.ID < b.ID },
		func(a, b Person) bool { return a.ID == b.ID },
		func(v Person) uint64 { return uint64(v.ID) },
	)

	// Insert custom values
	tree.Insert(Person{ID: 1, Name: "Alice"})
	tree.Insert(Person{ID: 2, Name: "Bob"})
	tree.Insert(Person{ID: 3, Name: "Charlie"})

	// Test Contains
	if !tree.Contains(Person{ID: 1, Name: ""}) { // Note: only ID is used for comparison
		t.Errorf("Tree should contain person with ID 1")
	}
	if tree.Contains(Person{ID: 4, Name: ""}) {
		t.Errorf("Tree should not contain person with ID 4")
	}

	// Test Delete
	tree.Delete(Person{ID: 2, Name: ""})
	if tree.Contains(Person{ID: 2, Name: ""}) {
		t.Errorf("Tree should not contain deleted person")
	}

	// Test GetAllKeys
	keys := tree.GetAllKeys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}
}
