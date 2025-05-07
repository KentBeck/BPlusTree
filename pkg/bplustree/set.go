package bplustree

// Set represents a set of uint64 values implemented using a B+ tree
// This is a backward-compatible wrapper around GenericSet[uint64]
type Set struct {
	genericSet *GenericSet[uint64]
}

// NewSet creates a new set with the given branching factor
func NewSet(branchingFactor int) *Set {
	return &Set{
		genericSet: NewUint64Set(branchingFactor),
	}
}

// Add adds a value to the set
// Returns true if the value was added, false if it already existed
func (s *Set) Add(value uint64) bool {
	return s.genericSet.Add(value)
}

// Contains returns true if the set contains the value
func (s *Set) Contains(value uint64) bool {
	return s.genericSet.Contains(value)
}

// Delete removes a value from the set
// Returns true if the value was removed, false if it didn't exist
func (s *Set) Delete(value uint64) bool {
	return s.genericSet.Delete(value)
}

// Size returns the number of elements in the set
func (s *Set) Size() int {
	return s.genericSet.Size()
}

// IsEmpty returns true if the set is empty
func (s *Set) IsEmpty() bool {
	return s.genericSet.IsEmpty()
}

// Clear removes all elements from the set
func (s *Set) Clear() {
	s.genericSet.Clear()
}

// GetAll returns all elements in the set
func (s *Set) GetAll() []uint64 {
	return s.genericSet.GetAll()
}
