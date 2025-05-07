package bplustree

// Helper functions for minimum key calculations
func minInternalKeys(branchingFactor int) int {
	return (branchingFactor+1)/2 - 1
}

func minLeafKeys(branchingFactor int) int {
	return (branchingFactor + 1) / 2
}

// BPlusTree is an alias for GenericBPlusTree[uint64] for backward compatibility
type BPlusTree = GenericBPlusTree[uint64]

// NewBPlusTree creates a new B+ tree with the given branching factor
// This is kept for backward compatibility
func NewBPlusTree(branchingFactor int) *BPlusTree {
	return NewBPlusTreeWithOptions(branchingFactor, true)
}

// NewBPlusTreeWithOptions creates a new B+ tree with the given branching factor and bloom filter option
// This is kept for backward compatibility
func NewBPlusTreeWithOptions(branchingFactor int, useBloomFilter bool) *BPlusTree {
	if useBloomFilter {
		return NewGenericBPlusTree[uint64](
			branchingFactor,
			func(a, b uint64) bool { return a < b },
			func(a, b uint64) bool { return a == b },
			func(v uint64) uint64 { return v },
		)
	} else {
		// Use a NullBloomFilter by creating a tree with a null hash function
		tree := NewGenericBPlusTree[uint64](
			branchingFactor,
			func(a, b uint64) bool { return a < b },
			func(a, b uint64) bool { return a == b },
			func(v uint64) uint64 { return v },
		)
		tree.bloomFilter = NewNullBloomFilter()
		return tree
	}
}
