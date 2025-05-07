package bplustree

import (
	"testing"
)

// TestCountKeys tests the CountKeys method
func TestCountKeys(t *testing.T) {
	// Create a tree with various numbers of keys
	tree := NewBPlusTree(4)
	
	// Empty tree should have 0 keys
	if count := tree.CountKeys(); count != 0 {
		t.Errorf("Expected empty tree to have 0 keys, got %d", count)
	}
	
	// Insert some keys
	keysToInsert := []uint64{10, 20, 30, 40, 50}
	for _, key := range keysToInsert {
		tree.Insert(key)
	}
	
	// Count should match the number of keys inserted
	if count := tree.CountKeys(); count != len(keysToInsert) {
		t.Errorf("Expected tree to have %d keys, got %d", len(keysToInsert), count)
	}
	
	// Delete a key and check count again
	tree.Delete(30)
	if count := tree.CountKeys(); count != len(keysToInsert)-1 {
		t.Errorf("Expected tree to have %d keys after deletion, got %d", len(keysToInsert)-1, count)
	}
}

// TestGetAllKeys tests the GetAllKeys method
func TestGetAllKeys(t *testing.T) {
	// Create a tree and insert keys
	tree := NewBPlusTree(4)
	keysToInsert := []uint64{50, 30, 70, 20, 40, 60, 80}
	for _, key := range keysToInsert {
		tree.Insert(key)
	}
	
	// Get all keys and verify
	allKeys := tree.GetAllKeys()
	
	// Check that we got the right number of keys
	if len(allKeys) != len(keysToInsert) {
		t.Errorf("Expected to get %d keys, got %d", len(keysToInsert), len(allKeys))
	}
	
	// Check that all inserted keys are present
	keyMap := make(map[uint64]bool)
	for _, key := range allKeys {
		keyMap[key] = true
	}
	
	for _, key := range keysToInsert {
		if !keyMap[key] {
			t.Errorf("Key %d is missing from GetAllKeys result", key)
		}
	}
}

// TestResetSize tests the ResetSize method
func TestResetSize(t *testing.T) {
	// Create a tree and insert keys
	tree := NewBPlusTree(4)
	for i := uint64(1); i <= 10; i++ {
		tree.Insert(i)
	}
	
	// Manually set the size to an incorrect value
	tree.size = 5
	
	// Reset the size
	tree.ResetSize()
	
	// Verify the size is corrected
	if tree.Size() != 10 {
		t.Errorf("Expected size to be reset to 10, got %d", tree.Size())
	}
}

// TestDeleteAll tests the DeleteAll method
func TestDeleteAll(t *testing.T) {
	// Create a tree and insert keys
	tree := NewBPlusTree(4)
	for i := uint64(1); i <= 10; i++ {
		tree.Insert(i)
	}
	
	// Delete all keys
	tree.DeleteAll()
	
	// Verify the tree is empty
	if tree.Size() != 0 {
		t.Errorf("Expected size to be 0 after DeleteAll, got %d", tree.Size())
	}
	
	// Verify height is reset to 1
	if tree.Height() != 1 {
		t.Errorf("Expected height to be 1 after DeleteAll, got %d", tree.Height())
	}
	
	// Verify no keys are present
	for i := uint64(1); i <= 10; i++ {
		if tree.Contains(i) {
			t.Errorf("Key %d should not exist after DeleteAll", i)
		}
	}
}

// TestForceDeleteKeys tests the ForceDeleteKeys method
func TestForceDeleteKeys(t *testing.T) {
	// Create a tree and insert keys
	tree := NewBPlusTree(4)
	for i := uint64(1); i <= 10; i++ {
		tree.Insert(i)
	}
	
	// Force delete some keys
	keysToDelete := []uint64{2, 4, 6, 8, 10}
	deletedCount := tree.ForceDeleteKeys(keysToDelete)
	
	// Verify the correct number of keys were deleted
	if deletedCount != len(keysToDelete) {
		t.Errorf("Expected to delete %d keys, got %d", len(keysToDelete), deletedCount)
	}
	
	// Verify the deleted keys are gone
	for _, key := range keysToDelete {
		if tree.Contains(key) {
			t.Errorf("Key %d should not exist after ForceDeleteKeys", key)
		}
	}
	
	// Verify the remaining keys still exist
	for _, key := range []uint64{1, 3, 5, 7, 9} {
		if !tree.Contains(key) {
			t.Errorf("Key %d should still exist after ForceDeleteKeys", key)
		}
	}
	
	// Verify the tree size
	if tree.Size() != 5 {
		t.Errorf("Expected tree size to be 5 after ForceDeleteKeys, got %d", tree.Size())
	}
	
	// Test deleting keys that don't exist
	nonExistentKeys := []uint64{100, 200, 300}
	deletedCount = tree.ForceDeleteKeys(nonExistentKeys)
	if deletedCount != 0 {
		t.Errorf("Expected to delete 0 non-existent keys, got %d", deletedCount)
	}
	
	// Test deleting all remaining keys
	remainingKeys := []uint64{1, 3, 5, 7, 9}
	deletedCount = tree.ForceDeleteKeys(remainingKeys)
	if deletedCount != len(remainingKeys) {
		t.Errorf("Expected to delete %d remaining keys, got %d", len(remainingKeys), deletedCount)
	}
	
	// Verify the tree is empty
	if tree.Size() != 0 {
		t.Errorf("Expected tree size to be 0 after deleting all keys, got %d", tree.Size())
	}
}

// TestString tests the String method
func TestString(t *testing.T) {
	// Create a tree
	tree := NewBPlusTree(4)
	
	// Test empty tree
	str := tree.String()
	if str == "" {
		t.Errorf("String() should not return an empty string")
	}
	
	// Insert some keys
	for i := uint64(1); i <= 5; i++ {
		tree.Insert(i)
	}
	
	// Test non-empty tree
	str = tree.String()
	if str == "" {
		t.Errorf("String() should not return an empty string")
	}
}

// TestSetBloomFilterParams tests the SetBloomFilterParams method
func TestSetBloomFilterParams(t *testing.T) {
	// Create a tree
	tree := NewBPlusTree(4)
	
	// Set custom Bloom filter parameters
	tree.SetBloomFilterParams(1000, 5)
	
	// Insert some keys
	for i := uint64(1); i <= 10; i++ {
		tree.Insert(i)
	}
	
	// Verify the keys can be found
	for i := uint64(1); i <= 10; i++ {
		if !tree.Contains(i) {
			t.Errorf("Key %d should exist after setting Bloom filter parameters", i)
		}
	}
}

// TestSetCustomBloomFilter tests the SetCustomBloomFilter method
func TestSetCustomBloomFilter(t *testing.T) {
	// Create a tree
	tree := NewBPlusTree(4)
	
	// Create a custom Bloom filter
	customFilter := NewBloomFilter(2000, 7)
	
	// Set the custom Bloom filter
	tree.SetCustomBloomFilter(customFilter)
	
	// Insert some keys
	for i := uint64(1); i <= 10; i++ {
		tree.Insert(i)
	}
	
	// Verify the keys can be found
	for i := uint64(1); i <= 10; i++ {
		if !tree.Contains(i) {
			t.Errorf("Key %d should exist after setting custom Bloom filter", i)
		}
	}
	
	// Test with an invalid filter type (should be ignored)
	tree.SetCustomBloomFilter("not a bloom filter")
	
	// Verify the keys can still be found
	for i := uint64(1); i <= 10; i++ {
		if !tree.Contains(i) {
			t.Errorf("Key %d should exist after attempting to set invalid Bloom filter", i)
		}
	}
}
