package bplustree

// Set is an alias for GenericSet[uint64] for backward compatibility
type Set = GenericSet[uint64]

// NewSet creates a new set with the given branching factor
// This is kept for backward compatibility
func NewSet(branchingFactor int) *Set {
	return NewUint64Set(branchingFactor)
}
