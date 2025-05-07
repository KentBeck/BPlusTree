package bplustree

// GenericSet represents a set of values of type V implemented using a B+ tree
// V must be a type that can be used as a map key (comparable)
type GenericSet[V any] struct {
	tree *BPlusTree
	// We need a way to convert between the generic type V and uint64
	// which is used internally by the B+ tree
	toUint64   func(V) uint64
	fromUint64 func(uint64) V
}

// NewGenericSet creates a new set with the given branching factor
// and conversion functions between V and uint64
func NewGenericSet[V any](
	branchingFactor int,
	toUint64 func(V) uint64,
	fromUint64 func(uint64) V,
) *GenericSet[V] {
	return &GenericSet[V]{
		tree:       NewBPlusTree(branchingFactor),
		toUint64:   toUint64,
		fromUint64: fromUint64,
	}
}

// NewUint64Set creates a new set for uint64 values
func NewUint64Set(branchingFactor int) *GenericSet[uint64] {
	return NewGenericSet[uint64](
		branchingFactor,
		func(v uint64) uint64 { return v },
		func(v uint64) uint64 { return v },
	)
}

// NewIntSet creates a new set for int values
func NewIntSet(branchingFactor int) *GenericSet[int] {
	return NewGenericSet[int](
		branchingFactor,
		func(v int) uint64 { return uint64(v) },
		func(v uint64) int { return int(v) },
	)
}

// Add adds a value to the set
// Returns true if the value was added, false if it already existed
func (s *GenericSet[V]) Add(value V) bool {
	return s.tree.Insert(s.toUint64(value))
}

// Contains returns true if the set contains the value
func (s *GenericSet[V]) Contains(value V) bool {
	return s.tree.Contains(s.toUint64(value))
}

// Delete removes a value from the set
// Returns true if the value was removed, false if it didn't exist
func (s *GenericSet[V]) Delete(value V) bool {
	return s.tree.Delete(s.toUint64(value))
}

// Size returns the number of elements in the set
func (s *GenericSet[V]) Size() int {
	return s.tree.Size()
}

// IsEmpty returns true if the set is empty
func (s *GenericSet[V]) IsEmpty() bool {
	return s.Size() == 0
}

// Clear removes all elements from the set
func (s *GenericSet[V]) Clear() {
	s.tree = NewBPlusTree(s.tree.branchingFactor)
}

// GetAll returns all elements in the set
func (s *GenericSet[V]) GetAll() []V {
	keys := s.tree.GetAllKeys()
	result := make([]V, len(keys))
	for i, key := range keys {
		result[i] = s.fromUint64(key)
	}
	return result
}
