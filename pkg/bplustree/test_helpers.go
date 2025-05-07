package bplustree

// This file contains helper functions for tests that were previously using
// the backward compatibility functions.

// NewBPlusTree creates a new B+ tree with uint64 keys for testing purposes
func NewBPlusTree(branchingFactor int) *GenericBPlusTree[uint64] {
	return NewGenericBPlusTree[uint64](
		branchingFactor,
		func(a, b uint64) bool { return a < b },
		func(a, b uint64) bool { return a == b },
		func(v uint64) uint64 { return v },
	)
}

// NewBPlusTreeWithOptions creates a new B+ tree with uint64 keys for testing purposes
func NewBPlusTreeWithOptions(branchingFactor int, useBloomFilter bool) *GenericBPlusTree[uint64] {
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
