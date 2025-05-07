// Package bplustree provides a generic implementation of a B+ tree data structure.
//
// A B+ tree is a self-balancing tree data structure that maintains sorted data
// and allows searches, sequential access, insertions, and deletions in logarithmic time.
// This implementation is generic and can work with any comparable type.
//
// The B+ tree is optimized with a bloom filter for faster lookups of non-existent keys.
package bplustree

import (
	"hash"
	"hash/fnv"
	"math"
)

// BloomFilterInterface defines the interface for a Bloom filter.
// A Bloom filter is a space-efficient probabilistic data structure that is used
// to test whether an element is a member of a set. False positives are possible,
// but false negatives are not.
type BloomFilterInterface interface {
	// Add adds a key to the Bloom filter.
	Add(key uint64)

	// Contains returns true if the key might be in the set.
	// Returns false if the key is definitely not in the set.
	// May return true for keys that are not in the set (false positives).
	Contains(key uint64) bool

	// Clear resets the Bloom filter, removing all elements.
	Clear()

	// SetValid marks the Bloom filter as valid.
	// A valid Bloom filter has been properly initialized with all keys.
	SetValid()

	// IsValid returns true if the Bloom filter is valid.
	IsValid() bool
}

// BloomFilter is a probabilistic data structure that tests whether an element is a member of a set.
// It is space-efficient but may produce false positives (indicating an element is in the set when it is not).
// It will never produce false negatives (indicating an element is not in the set when it is).
type BloomFilter struct {
	bits          []bool // The bit array
	size          int    // Size of the bit array
	hashFunctions int    // Number of hash functions
	valid         bool   // Whether the filter is valid
}

// NewBloomFilter creates a new Bloom filter with the given size and number of hash functions.
// Parameters:
//   - size: The size of the bit array. Larger sizes reduce false positives but use more memory.
//   - hashFunctions: The number of hash functions to use. More functions reduce false positives
//     but increase computation time.
//
// Returns a new Bloom filter initialized to empty (all bits set to false).
func NewBloomFilter(size int, hashFunctions int) *BloomFilter {
	return &BloomFilter{
		bits:          make([]bool, size),
		size:          size,
		hashFunctions: hashFunctions,
		valid:         false,
	}
}

// Add adds a key to the Bloom filter by setting bits at positions determined by the hash functions.
// Time complexity: O(k) where k is the number of hash functions.
func (bf *BloomFilter) Add(key uint64) {
	for i := 0; i < bf.hashFunctions; i++ {
		// Calculate bit position using the i-th hash function
		position := bf.hash(key, i) % uint64(bf.size)

		// Set the bit at that position
		bf.bits[position] = true
	}
}

// Contains returns true if the key might be in the set, false if it's definitely not in the set.
// This method may return false positives (true when the key is not actually in the set),
// but it will never return false negatives (false when the key is actually in the set).
// Time complexity: O(k) where k is the number of hash functions.
func (bf *BloomFilter) Contains(key uint64) bool {
	for i := 0; i < bf.hashFunctions; i++ {
		// Calculate bit position using the i-th hash function
		position := bf.hash(key, i) % uint64(bf.size)

		// If any bit is not set, the key is definitely not in the set
		if !bf.bits[position] {
			return false
		}
	}

	// All bits are set, so the key might be in the set
	return true
}

// Clear resets the Bloom filter by creating a new bit array and marking it as invalid.
// Time complexity: O(m) where m is the size of the bit array.
func (bf *BloomFilter) Clear() {
	bf.bits = make([]bool, bf.size)
	bf.valid = false
}

// SetValid marks the Bloom filter as valid.
// A valid Bloom filter has been properly initialized with all keys.
// Time complexity: O(1)
func (bf *BloomFilter) SetValid() {
	bf.valid = true
}

// IsValid returns true if the Bloom filter is valid.
// Time complexity: O(1)
func (bf *BloomFilter) IsValid() bool {
	return bf.valid
}

// hash generates a hash value for a key using the FNV-1a hash function
// with a seed based on the hash function index.
// Time complexity: O(1)
func (bf *BloomFilter) hash(key uint64, index int) uint64 {
	return hashWithSeed(key, uint64(index+1))
}

// hashWithSeed generates a hash value for a key with a given seed using FNV-1a.
// Time complexity: O(1)
func hashWithSeed(key uint64, seed uint64) uint64 {
	h := fnv.New64a()

	// Write seed bytes to the hash
	writeUint64ToHash(h, seed)

	// Write key bytes to the hash
	writeUint64ToHash(h, key)

	// Return the 64-bit hash value
	return h.Sum64()
}

// writeUint64ToHash writes a uint64 value to a hash.
// Time complexity: O(1)
func writeUint64ToHash(h hash.Hash, value uint64) {
	bytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		bytes[i] = byte(value >> (i * 8))
	}
	h.Write(bytes)
}

// OptimalBloomFilterSize calculates the optimal size and number of hash functions
// for a Bloom filter given the expected number of elements and desired false positive rate.
//
// Parameters:
//   - expectedElements: The expected number of elements to be inserted into the filter.
//   - falsePositiveRate: The desired false positive rate (between 0 and 1).
//
// Returns:
//   - size: The optimal size of the bit array.
//   - hashFunctions: The optimal number of hash functions.
//
// Time complexity: O(1)
func OptimalBloomFilterSize(expectedElements int, falsePositiveRate float64) (size int, hashFunctions int) {
	// Handle edge cases
	if expectedElements <= 0 {
		// For zero or negative elements, return a small default size
		return 10, 3
	}

	if falsePositiveRate <= 0 {
		falsePositiveRate = 0.01 // Default to 1% false positive rate
	} else if falsePositiveRate >= 1 {
		falsePositiveRate = 0.99 // Cap at 99% false positive rate
	}

	// Calculate optimal size (in bits)
	// Formula: m = -n*ln(p)/(ln(2)^2)
	// where m is size, n is expectedElements, p is falsePositiveRate
	size = int(math.Ceil(-float64(expectedElements) * math.Log(falsePositiveRate) / math.Pow(math.Log(2), 2)))

	// Calculate optimal number of hash functions
	// Formula: k = (m/n)*ln(2)
	// where k is hashFunctions, m is size, n is expectedElements
	hashFunctions = int(math.Ceil(float64(size) / float64(expectedElements) * math.Log(2)))

	// Ensure minimum values
	if size < 10 {
		size = 10 // Default minimum size
	}
	if hashFunctions < 1 {
		hashFunctions = 3 // Default minimum hash functions
	} else if hashFunctions > 20 {
		hashFunctions = 20 // Cap at 20 hash functions for performance
	}

	return size, hashFunctions
}

// NullBloomFilter is a null implementation of BloomFilterInterface that always answers "maybe".
// It's used when a bloom filter is not needed or desired, such as for small trees
// where the overhead of maintaining a bloom filter might exceed its benefits.
// This implementation is more efficient than a real bloom filter for small datasets.
type NullBloomFilter struct {
	// No fields needed - all methods have fixed behavior
}

// NewNullBloomFilter creates a new NullBloomFilter.
// This is a lightweight implementation that uses minimal memory.
func NewNullBloomFilter() *NullBloomFilter {
	return &NullBloomFilter{}
}

// Add is a no-op in NullBloomFilter.
// Time complexity: O(1)
func (bf *NullBloomFilter) Add(key uint64) {
	// Do nothing - this is intentionally a no-op
}

// Contains always returns true in NullBloomFilter (meaning "maybe").
// This means the tree will always need to check for the key's presence.
// Time complexity: O(1)
func (bf *NullBloomFilter) Contains(key uint64) bool {
	return true
}

// Clear is a no-op in NullBloomFilter.
// Time complexity: O(1)
func (bf *NullBloomFilter) Clear() {
	// Do nothing - this is intentionally a no-op
}

// SetValid is a no-op in NullBloomFilter since it's always valid.
// Time complexity: O(1)
func (bf *NullBloomFilter) SetValid() {
	// Do nothing - this is intentionally a no-op
}

// IsValid always returns true for NullBloomFilter.
// Time complexity: O(1)
func (bf *NullBloomFilter) IsValid() bool {
	return true
}
