package bplustree

import (
	"testing"
)

// TestRangeQuery tests the range query functionality of the generic B+ tree
func TestRangeQuery(t *testing.T) {
	// Create a tree with integers
	tree := NewGenericBPlusTree[int](
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

	// Test various range queries
	testCases := []struct {
		name     string
		start    int
		end      int
		expected []int
	}{
		{
			name:     "Full range",
			start:    0,
			end:      110,
			expected: []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
		},
		{
			name:     "Partial range",
			start:    25,
			end:      75,
			expected: []int{30, 40, 50, 60, 70},
		},
		{
			name:     "Exact bounds",
			start:    10,
			end:      100,
			expected: []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
		},
		{
			name:     "Empty range",
			start:    31,
			end:      39,
			expected: []int{},
		},
		{
			name:     "Single value",
			start:    50,
			end:      50,
			expected: []int{50},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tree.RangeQuery(tc.start, tc.end)

			// Check length
			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d values in range [%d, %d], got %d",
					len(tc.expected), tc.start, tc.end, len(result))
				t.Errorf("Expected: %v, Got: %v", tc.expected, result)
				return
			}

			// Check values
			for i, v := range tc.expected {
				if result[i] != v {
					t.Errorf("Expected %d at position %d, got %d", v, i, result[i])
				}
			}
		})
	}
}

// TestSetRange tests the Range method of Set
func TestSetRange(t *testing.T) {
	// Create a set
	set := NewSet(4)

	// Add some values
	values := []uint64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	for _, v := range values {
		set.Add(v)
	}

	// Test range query
	result := set.Range(25, 75)

	// Check result
	expected := []uint64{30, 40, 50, 60, 70}
	if len(result) != len(expected) {
		t.Errorf("Expected %d values in range [25, 75], got %d", len(expected), len(result))
		t.Errorf("Expected: %v, Got: %v", expected, result)
		return
	}

	// Check values
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("Expected %d at position %d, got %d", v, i, result[i])
		}
	}
}

// TestGenericSetRangeWithStrings tests the Range method with string values
func TestGenericSetRangeWithStrings(t *testing.T) {
	// Create a string set
	set := NewStringSet(4)

	// Add some values
	set.Add("apple")
	set.Add("banana")
	set.Add("cherry")
	set.Add("date")
	set.Add("elderberry")
	set.Add("fig")
	set.Add("grape")

	// Test range query
	result := set.Range("blueberry", "fig")

	// Check result
	expected := []string{"cherry", "date", "elderberry", "fig"}
	if len(result) != len(expected) {
		t.Errorf("Expected %d values in range [blueberry, fig], got %d", len(expected), len(result))
		t.Errorf("Expected: %v, Got: %v", expected, result)
		return
	}

	// Check values
	for i, v := range expected {
		if i < len(result) && result[i] != v {
			t.Errorf("Expected %s at position %d, got %s", v, i, result[i])
		}
	}
}
