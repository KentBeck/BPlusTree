package bplustree

// Set represents a set of uint64 values implemented using a B+ tree
type Set struct {
	tree *BPlusTree
}

// NewSet creates a new set with the given branching factor
func NewSet(branchingFactor int) *Set {
	return &Set{
		tree: NewBPlusTree(branchingFactor),
	}
}

// Add adds a value to the set
// Returns true if the value was added, false if it already existed
func (s *Set) Add(value uint64) bool {
	return s.tree.Insert(value)
}

// Contains returns true if the set contains the value
func (s *Set) Contains(value uint64) bool {
	return s.tree.Contains(value)
}

// Delete removes a value from the set
// Returns true if the value was removed, false if it didn't exist
func (s *Set) Delete(value uint64) bool {
	return s.tree.Delete(value)
}

// Size returns the number of elements in the set
func (s *Set) Size() int {
	return s.tree.Size()
}

// IsEmpty returns true if the set is empty
func (s *Set) IsEmpty() bool {
	return s.Size() == 0
}

// Clear removes all elements from the set
func (s *Set) Clear() {
	s.tree = NewBPlusTree(s.tree.branchingFactor)
}
