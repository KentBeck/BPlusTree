package bplustree

import (
	"hash"
	"hash/fnv"
	"math"
)

// BloomFilterInterface defines the interface for a Bloom filter
type BloomFilterInterface interface {
	// Add adds a key to the Bloom filter
	Add(key uint64)
	// Contains returns true if the key might be in the set
	Contains(key uint64) bool
	// Clear resets the Bloom filter
	Clear()
	// SetValid marks the Bloom filter as valid
	SetValid()
	// IsValid returns true if the Bloom filter is valid
	IsValid() bool
}

// BloomFilter is a probabilistic data structure that tests whether an element is a member of a set
type BloomFilter struct {
	bits          []bool
	size          int
	hashFunctions int
	valid         bool
}

// NewBloomFilter creates a new Bloom filter with the given size and number of hash functions
func NewBloomFilter(size int, hashFunctions int) *BloomFilter {
	return &BloomFilter{
		bits:          make([]bool, size),
		size:          size,
		hashFunctions: hashFunctions,
		valid:         false,
	}
}

// Add adds a key to the Bloom filter
func (bf *BloomFilter) Add(key uint64) {
	for i := 0; i < bf.hashFunctions; i++ {
		position := bf.hash(key, i) % uint64(bf.size)
		bf.bits[position] = true
	}
}

// Contains returns true if the key might be in the set
// Returns false if the key is definitely not in the set
func (bf *BloomFilter) Contains(key uint64) bool {
	for i := 0; i < bf.hashFunctions; i++ {
		position := bf.hash(key, i) % uint64(bf.size)
		if !bf.bits[position] {
			return false
		}
	}
	return true
}

// Clear resets the Bloom filter
func (bf *BloomFilter) Clear() {
	bf.bits = make([]bool, bf.size)
	bf.valid = false
}

// SetValid marks the Bloom filter as valid
func (bf *BloomFilter) SetValid() {
	bf.valid = true
}

// IsValid returns true if the Bloom filter is valid
func (bf *BloomFilter) IsValid() bool {
	return bf.valid
}

// hash generates a hash value for a key using the FNV-1a hash function
// with a seed based on the hash function index
func (bf *BloomFilter) hash(key uint64, index int) uint64 {
	return hashWithSeed(key, uint64(index+1))
}

// hashWithSeed generates a hash value for a key with a given seed
func hashWithSeed(key uint64, seed uint64) uint64 {
	h := fnv.New64a()

	// Write seed bytes
	writeUint64ToHash(h, seed)

	// Write key bytes
	writeUint64ToHash(h, key)

	return h.Sum64()
}

// writeUint64ToHash writes a uint64 value to a hash
func writeUint64ToHash(h hash.Hash, value uint64) {
	bytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		bytes[i] = byte(value >> (i * 8))
	}
	h.Write(bytes)
}

// OptimalBloomFilterSize calculates the optimal size for a Bloom filter
// given the expected number of elements and desired false positive rate
func OptimalBloomFilterSize(expectedElements int, falsePositiveRate float64) (size int, hashFunctions int) {
	// Calculate optimal size (in bits)
	size = int(math.Ceil(-float64(expectedElements) * math.Log(falsePositiveRate) / math.Pow(math.Log(2), 2)))

	// Calculate optimal number of hash functions
	hashFunctions = int(math.Ceil(float64(size) / float64(expectedElements) * math.Log(2)))

	// Ensure minimum values
	if size < 1 {
		size = 1024 // Default minimum size
	}
	if hashFunctions < 1 {
		hashFunctions = 3 // Default minimum hash functions
	}

	return size, hashFunctions
}

// NullBloomFilter is a null implementation of BloomFilterInterface that always answers "maybe"
// It's used to replace explicit nil checks in the tree implementation
type NullBloomFilter struct {
	valid bool
}

// NewNullBloomFilter creates a new NullBloomFilter
func NewNullBloomFilter() *NullBloomFilter {
	return &NullBloomFilter{
		valid: true,
	}
}

// Add is a no-op in NullBloomFilter
func (bf *NullBloomFilter) Add(key uint64) {
	// Do nothing
}

// Contains always returns true in NullBloomFilter (meaning "maybe")
func (bf *NullBloomFilter) Contains(key uint64) bool {
	return true
}

// Clear is a no-op in NullBloomFilter
func (bf *NullBloomFilter) Clear() {
	// Do nothing
}

// SetValid marks the NullBloomFilter as valid
func (bf *NullBloomFilter) SetValid() {
	bf.valid = true
}

// IsValid returns true if the NullBloomFilter is valid
func (bf *NullBloomFilter) IsValid() bool {
	return bf.valid
}
