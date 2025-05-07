package bplustree

import (
	"sort"
)

// GenericSet represents a set of values of type K implemented using a generic B+ tree
// K must be a comparable type
type GenericSet[K comparable] struct {
	tree *GenericBPlusTree[K]
}

// NewGenericSet creates a new set with the given branching factor
// and comparison functions
func NewGenericSet[K comparable](
	branchingFactor int,
	less func(a, b K) bool,
	equal func(a, b K) bool,
	hashFunc func(K) uint64,
) *GenericSet[K] {
	return &GenericSet[K]{
		tree: NewGenericBPlusTree(branchingFactor, less, equal, hashFunc),
	}
}

// NewUint64Set creates a new set for uint64 values
func NewUint64Set(branchingFactor int) *GenericSet[uint64] {
	return NewGenericSet[uint64](
		branchingFactor,
		func(a, b uint64) bool { return a < b },
		func(a, b uint64) bool { return a == b },
		func(v uint64) uint64 { return v }, // Simple identity hash function
	)
}

// NewIntSet creates a new set for int values
func NewIntSet(branchingFactor int) *GenericSet[int] {
	return NewGenericSet[int](
		branchingFactor,
		func(a, b int) bool { return a < b },
		func(a, b int) bool { return a == b },
		func(v int) uint64 { return uint64(v) }, // Convert to uint64 for hashing
	)
}

// NewStringSet creates a new set for string values
func NewStringSet(branchingFactor int) *GenericSet[string] {
	// Simple hash function for strings
	stringHash := func(s string) uint64 {
		var hash uint64
		for i := 0; i < len(s); i++ {
			hash = hash*31 + uint64(s[i])
		}
		return hash
	}

	return NewGenericSet[string](
		branchingFactor,
		func(a, b string) bool { return a < b },
		func(a, b string) bool { return a == b },
		stringHash,
	)
}

// Add adds a value to the set
// Returns true if the value was added, false if it already existed
func (s *GenericSet[K]) Add(value K) bool {
	return s.tree.Insert(value)
}

// Contains returns true if the set contains the value
func (s *GenericSet[K]) Contains(value K) bool {
	return s.tree.Contains(value)
}

// Delete removes a value from the set
// Returns true if the value was removed, false if it didn't exist
func (s *GenericSet[K]) Delete(value K) bool {
	return s.tree.Delete(value)
}

// Size returns the number of elements in the set
func (s *GenericSet[K]) Size() int {
	return s.tree.Size()
}

// IsEmpty returns true if the set is empty
func (s *GenericSet[K]) IsEmpty() bool {
	return s.Size() == 0
}

// Clear removes all elements from the set
func (s *GenericSet[K]) Clear() {
	branchingFactor := s.tree.branchingFactor
	less := s.tree.less
	equal := s.tree.equal
	hashFunc := s.tree.hashFunc
	s.tree = NewGenericBPlusTree(branchingFactor, less, equal, hashFunc)
}

// GetAll returns all elements in the set
func (s *GenericSet[K]) GetAll() []K {
	return s.tree.GetAllKeys()
}

// SortedSlice returns all elements in the set as a sorted slice
func (s *GenericSet[K]) SortedSlice() []K {
	result := s.GetAll()
	sort.Slice(result, func(i, j int) bool {
		return s.tree.less(result[i], result[j])
	})
	return result
}

// Range returns all elements in the range [start, end]
func (s *GenericSet[K]) Range(start, end K) []K {
	return s.tree.RangeQuery(start, end)
}
