package bplustree

import (
	"testing"
)

func TestSetAdd(t *testing.T) {
	set := NewSet(4)
	
	// Add some values
	if !set.Add(10) {
		t.Errorf("Failed to add value 10")
	}
	if !set.Add(20) {
		t.Errorf("Failed to add value 20")
	}
	if !set.Add(5) {
		t.Errorf("Failed to add value 5")
	}
	
	// Try to add a duplicate
	if set.Add(10) {
		t.Errorf("Added duplicate value 10")
	}
	
	// Check size
	if set.Size() != 3 {
		t.Errorf("Expected size 3, got %d", set.Size())
	}
}

func TestSetContains(t *testing.T) {
	set := NewSet(4)
	
	// Add some values
	set.Add(10)
	set.Add(20)
	set.Add(5)
	
	// Check contains
	if !set.Contains(10) {
		t.Errorf("Expected to contain value 10")
	}
	if !set.Contains(20) {
		t.Errorf("Expected to contain value 20")
	}
	if !set.Contains(5) {
		t.Errorf("Expected to contain value 5")
	}
	if set.Contains(15) {
		t.Errorf("Expected not to contain value 15")
	}
}

func TestSetDelete(t *testing.T) {
	set := NewSet(4)
	
	// Add some values
	set.Add(10)
	set.Add(20)
	set.Add(5)
	
	// Delete a value
	if !set.Delete(10) {
		t.Errorf("Failed to delete value 10")
	}
	
	// Check contains
	if set.Contains(10) {
		t.Errorf("Expected not to contain value 10 after deletion")
	}
	
	// Try to delete a non-existent value
	if set.Delete(15) {
		t.Errorf("Deleted non-existent value 15")
	}
	
	// Check size
	if set.Size() != 2 {
		t.Errorf("Expected size 2, got %d", set.Size())
	}
}

func TestSetIsEmpty(t *testing.T) {
	set := NewSet(4)
	
	// Check empty set
	if !set.IsEmpty() {
		t.Errorf("Expected empty set")
	}
	
	// Add a value
	set.Add(10)
	
	// Check non-empty set
	if set.IsEmpty() {
		t.Errorf("Expected non-empty set")
	}
}

func TestSetClear(t *testing.T) {
	set := NewSet(4)
	
	// Add some values
	set.Add(10)
	set.Add(20)
	set.Add(5)
	
	// Clear the set
	set.Clear()
	
	// Check size
	if set.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", set.Size())
	}
	
	// Check empty
	if !set.IsEmpty() {
		t.Errorf("Expected empty set after clear")
	}
}
