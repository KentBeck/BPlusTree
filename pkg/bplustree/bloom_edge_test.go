package bplustree

import (
	"testing"
)

// TestOptimalBloomFilterSizeEdgeCases tests edge cases for the OptimalBloomFilterSize function
func TestOptimalBloomFilterSizeEdgeCases(t *testing.T) {
	// Test with very small expected elements
	size, hashFunctions := OptimalBloomFilterSize(1, 0.01)
	if size < 1 {
		t.Errorf("Size should be at least 1, got %d", size)
	}
	if hashFunctions < 1 {
		t.Errorf("Hash functions should be at least 1, got %d", hashFunctions)
	}

	// Test with zero expected elements (should use defaults)
	size, hashFunctions = OptimalBloomFilterSize(0, 0.01)
	if size < 100 {
		t.Errorf("Expected reasonable size for zero elements, got %d", size)
	}
	if hashFunctions < 1 {
		t.Errorf("Expected at least 1 hash function for zero elements, got %d", hashFunctions)
	}

	// Test with negative expected elements (should use defaults)
	size, hashFunctions = OptimalBloomFilterSize(-10, 0.01)
	if size < 100 {
		t.Errorf("Expected reasonable size for negative elements, got %d", size)
	}
	if hashFunctions < 1 {
		t.Errorf("Expected at least 1 hash function for negative elements, got %d", hashFunctions)
	}

	// Test with very high false positive rate (should result in small size)
	size, hashFunctions = OptimalBloomFilterSize(1000, 0.9)
	if size < 1 {
		t.Errorf("Size should be at least 1, got %d", size)
	}
	if hashFunctions < 1 {
		t.Errorf("Hash functions should be at least 1, got %d", hashFunctions)
	}

	// Test with very low false positive rate (should result in large size)
	size, hashFunctions = OptimalBloomFilterSize(1000, 0.0001)
	if size < 1000 {
		t.Errorf("Size should be large for low false positive rate, got %d", size)
	}
}

// TestNullBloomFilterMethods tests the methods of NullBloomFilter
func TestNullBloomFilterMethods(t *testing.T) {
	filter := NewNullBloomFilter()

	// Test Add method (no-op)
	filter.Add(123)

	// Test Clear method (no-op)
	filter.Clear()

	// Test SetValid method (no-op)
	filter.SetValid()

	// Test IsValid method (always returns true)
	if !filter.IsValid() {
		t.Errorf("IsValid should always return true for NullBloomFilter")
	}

	// Test Contains method (always returns true)
	if !filter.Contains(123) {
		t.Errorf("Contains should always return true for NullBloomFilter")
	}
}
