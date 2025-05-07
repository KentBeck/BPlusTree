package bplustree

import (
	"math/rand"
	"testing"
)

// BenchmarkOriginalBPlusTreeInsert benchmarks the insertion of keys into the original B+ tree
func BenchmarkOriginalBPlusTreeInsert(b *testing.B) {
	tree := NewBPlusTree(256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := uint64(rand.Intn(1000000))
		tree.Insert(key)
	}
}

// BenchmarkGenericBPlusTreeInsert benchmarks the insertion of keys into the generic B+ tree
func BenchmarkGenericBPlusTreeInsert(b *testing.B) {
	tree := NewGenericBPlusTree[uint64](
		256,
		func(a, b uint64) bool { return a < b },
		func(a, b uint64) bool { return a == b },
		func(v uint64) uint64 { return v },
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := uint64(rand.Intn(1000000))
		tree.Insert(key)
	}
}

// BenchmarkOriginalBPlusTreeContains benchmarks the lookup of keys in the original B+ tree
func BenchmarkOriginalBPlusTreeContains(b *testing.B) {
	tree := NewBPlusTree(256)

	// Insert a set of keys
	keys := make([]uint64, 1000000)
	for i := 0; i < 1000000; i++ {
		keys[i] = uint64(i)
		tree.Insert(keys[i])
	}

	// Force the Bloom filter to be computed
	tree.Contains(keys[0])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[rand.Intn(len(keys))]
		tree.Contains(key)
	}
}

// BenchmarkGenericBPlusTreeContains benchmarks the lookup of keys in the generic B+ tree
func BenchmarkGenericBPlusTreeContains(b *testing.B) {
	tree := NewGenericBPlusTree[uint64](
		256,
		func(a, b uint64) bool { return a < b },
		func(a, b uint64) bool { return a == b },
		func(v uint64) uint64 { return v },
	)

	// Insert a set of keys
	keys := make([]uint64, 1000000)
	for i := 0; i < 1000000; i++ {
		keys[i] = uint64(i)
		tree.Insert(keys[i])
	}

	// Force the Bloom filter to be computed
	tree.Contains(keys[0])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[rand.Intn(len(keys))]
		tree.Contains(key)
	}
}

// BenchmarkOriginalBPlusTreeContainsNonExistent benchmarks the lookup of non-existent keys in the original B+ tree
func BenchmarkOriginalBPlusTreeContainsNonExistent(b *testing.B) {
	tree := NewBPlusTree(256)

	// Insert a set of keys
	for i := 0; i < 1000000; i++ {
		tree.Insert(uint64(i))
	}

	// Force the Bloom filter to be computed
	tree.Contains(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := uint64(1000000 + rand.Intn(1000000)) // Keys that don't exist
		tree.Contains(key)
	}
}

// BenchmarkGenericBPlusTreeContainsNonExistent benchmarks the lookup of non-existent keys in the generic B+ tree
func BenchmarkGenericBPlusTreeContainsNonExistent(b *testing.B) {
	tree := NewGenericBPlusTree[uint64](
		256,
		func(a, b uint64) bool { return a < b },
		func(a, b uint64) bool { return a == b },
		func(v uint64) uint64 { return v },
	)

	// Insert a set of keys
	for i := 0; i < 1000000; i++ {
		tree.Insert(uint64(i))
	}

	// Force the Bloom filter to be computed
	tree.Contains(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := uint64(1000000 + rand.Intn(1000000)) // Keys that don't exist
		tree.Contains(key)
	}
}

// BenchmarkGenericBPlusTreeRangeQuery benchmarks the range query in the generic B+ tree
func BenchmarkGenericBPlusTreeRangeQuery(b *testing.B) {
	tree := NewGenericBPlusTree[uint64](
		256,
		func(a, b uint64) bool { return a < b },
		func(a, b uint64) bool { return a == b },
		func(v uint64) uint64 { return v },
	)

	// Insert a set of keys
	for i := 0; i < 1000000; i++ {
		tree.Insert(uint64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := uint64(rand.Intn(900000))
		end := start + uint64(rand.Intn(100000))
		tree.RangeQuery(start, end)
	}
}

// BenchmarkGenericSetWithStrings benchmarks the generic set with string values
func BenchmarkGenericSetWithStrings(b *testing.B) {
	set := NewStringSet(256)

	// Generate random strings
	strings := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		strings[i] = randomString(10)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Add(strings[rand.Intn(len(strings))])
	}
}

// Helper function to generate random strings
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
