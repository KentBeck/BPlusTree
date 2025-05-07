package bplustree

import (
	"testing"
)

func TestGenericSetUint64(t *testing.T) {
	set := NewUint64Set(4)

	// Add some values
	if !set.Add(uint64(10)) {
		t.Errorf("Failed to add value 10")
	}
	if !set.Add(uint64(20)) {
		t.Errorf("Failed to add value 20")
	}
	if !set.Add(uint64(5)) {
		t.Errorf("Failed to add value 5")
	}

	// Try to add a duplicate
	if set.Add(uint64(10)) {
		t.Errorf("Added duplicate value 10")
	}

	// Check size
	if set.Size() != 3 {
		t.Errorf("Expected size 3, got %d", set.Size())
	}

	// Check contains
	if !set.Contains(uint64(10)) {
		t.Errorf("Expected to contain value 10")
	}
	if !set.Contains(uint64(20)) {
		t.Errorf("Expected to contain value 20")
	}
	if !set.Contains(uint64(5)) {
		t.Errorf("Expected to contain value 5")
	}
	if set.Contains(uint64(15)) {
		t.Errorf("Expected not to contain value 15")
	}

	// Delete a value
	if !set.Delete(uint64(10)) {
		t.Errorf("Failed to delete value 10")
	}

	// Check contains after deletion
	if set.Contains(uint64(10)) {
		t.Errorf("Expected not to contain value 10 after deletion")
	}

	// Try to delete a non-existent value
	if set.Delete(uint64(15)) {
		t.Errorf("Deleted non-existent value 15")
	}

	// Check size after deletion
	if set.Size() != 2 {
		t.Errorf("Expected size 2 after deletion, got %d", set.Size())
	}

	// Clear the set
	set.Clear()

	// Check size after clear
	if set.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", set.Size())
	}

	// Check empty
	if !set.IsEmpty() {
		t.Errorf("Expected empty set after clear")
	}
}

func TestGenericSetInt(t *testing.T) {
	set := NewIntSet(4)

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

	// Delete a value
	if !set.Delete(10) {
		t.Errorf("Failed to delete value 10")
	}

	// Check contains after deletion
	if set.Contains(10) {
		t.Errorf("Expected not to contain value 10 after deletion")
	}

	// Try to delete a non-existent value
	if set.Delete(15) {
		t.Errorf("Deleted non-existent value 15")
	}

	// Check size after deletion
	if set.Size() != 2 {
		t.Errorf("Expected size 2 after deletion, got %d", set.Size())
	}

	// Get all values
	values := set.GetAll()
	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}

	// Check values
	found5 := false
	found20 := false
	for _, v := range values {
		if v == 5 {
			found5 = true
		}
		if v == 20 {
			found20 = true
		}
	}
	if !found5 {
		t.Errorf("Expected to find value 5 in GetAll")
	}
	if !found20 {
		t.Errorf("Expected to find value 20 in GetAll")
	}

	// Clear the set
	set.Clear()

	// Check size after clear
	if set.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", set.Size())
	}

	// Check empty
	if !set.IsEmpty() {
		t.Errorf("Expected empty set after clear")
	}
}

// TestGenericSetString tests the GenericSet with string values
func TestGenericSetString(t *testing.T) {
	// Create a string set with custom conversion functions
	set := NewGenericSet[string](
		4,
		func(s string) uint64 {
			// Simple hash function for strings
			var hash uint64
			for i := 0; i < len(s); i++ {
				hash = hash*31 + uint64(s[i])
			}
			return hash
		},
		func(hash uint64) string {
			// We can't convert back from hash to string
			// This is just a placeholder
			return ""
		},
		func(a, b string) bool { return a < b },
		func(a, b string) bool { return a == b },
	)

	// Add some values
	if !set.Add("apple") {
		t.Errorf("Failed to add value 'apple'")
	}
	if !set.Add("banana") {
		t.Errorf("Failed to add value 'banana'")
	}
	if !set.Add("cherry") {
		t.Errorf("Failed to add value 'cherry'")
	}

	// Try to add a duplicate
	if set.Add("apple") {
		t.Errorf("Added duplicate value 'apple'")
	}

	// Check size
	if set.Size() != 3 {
		t.Errorf("Expected size 3, got %d", set.Size())
	}

	// Check contains
	if !set.Contains("apple") {
		t.Errorf("Expected to contain value 'apple'")
	}
	if !set.Contains("banana") {
		t.Errorf("Expected to contain value 'banana'")
	}
	if !set.Contains("cherry") {
		t.Errorf("Expected to contain value 'cherry'")
	}
	if set.Contains("date") {
		t.Errorf("Expected not to contain value 'date'")
	}

	// Delete a value
	if !set.Delete("apple") {
		t.Errorf("Failed to delete value 'apple'")
	}

	// Check contains after deletion
	if set.Contains("apple") {
		t.Errorf("Expected not to contain value 'apple' after deletion")
	}

	// Check size after deletion
	if set.Size() != 2 {
		t.Errorf("Expected size 2 after deletion, got %d", set.Size())
	}
}
