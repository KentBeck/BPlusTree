package bplustree

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// BenchmarkBPlusTreeInsert benchmarks the insertion of keys into the B+ tree
func BenchmarkBPlusTreeInsert(b *testing.B) {
	tree := NewBPlusTree(256)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := uint64(rand.Intn(1000000))
		tree.Insert(key)
	}
}

// BenchmarkBPlusTreeContains benchmarks the lookup of keys in the B+ tree
func BenchmarkBPlusTreeContains(b *testing.B) {
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

// BenchmarkBPlusTreeContainsNonExistent benchmarks the lookup of non-existent keys
func BenchmarkBPlusTreeContainsNonExistent(b *testing.B) {
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

// TestPerformance is a comprehensive performance test that measures insertion and lookup times
func TestPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}
	
	// Parameters
	numKeys := 1000000
	numQueries := 1000000
	branchingFactor := 256
	
	// Create trees with and without Bloom filter
	treeWithBloom := NewBPlusTree(branchingFactor)
	
	// Create a tree without Bloom filter by modifying the Contains method
	treeWithoutBloom := NewBPlusTree(branchingFactor)
	
	// Generate random keys for insertion
	keys := make([]uint64, numKeys)
	for i := 0; i < numKeys; i++ {
		keys[i] = uint64(rand.Intn(numKeys * 10)) // Use a larger range to have some duplicates
	}
	
	// Measure insertion time for tree with Bloom filter
	startTime := time.Now()
	for _, key := range keys {
		treeWithBloom.Insert(key)
	}
	insertTimeWithBloom := time.Since(startTime)
	
	// Measure insertion time for tree without Bloom filter
	startTime = time.Now()
	for _, key := range keys {
		treeWithoutBloom.Insert(key)
	}
	insertTimeWithoutBloom := time.Since(startTime)
	
	// Generate random keys for queries (50% existing, 50% non-existing)
	queryKeys := make([]uint64, numQueries)
	for i := 0; i < numQueries; i++ {
		if rand.Intn(2) == 0 {
			// Existing key
			queryKeys[i] = keys[rand.Intn(len(keys))]
		} else {
			// Non-existing key
			queryKeys[i] = uint64(numKeys*10 + rand.Intn(numKeys*10))
		}
	}
	
	// Force Bloom filter computation for the tree with Bloom filter
	treeWithBloom.Contains(keys[0])
	
	// Measure query time for tree with Bloom filter
	startTime = time.Now()
	for _, key := range queryKeys {
		treeWithBloom.Contains(key)
	}
	queryTimeWithBloom := time.Since(startTime)
	
	// Measure query time for tree without Bloom filter by directly using findLeaf
	startTime = time.Now()
	for _, key := range queryKeys {
		treeWithoutBloom.findLeaf(treeWithoutBloom.root, key)
	}
	queryTimeWithoutBloom := time.Since(startTime)
	
	// Print results
	fmt.Printf("Performance Test Results (numKeys=%d, numQueries=%d, branchingFactor=%d):\n", 
		numKeys, numQueries, branchingFactor)
	fmt.Printf("Insertion Time (with Bloom filter): %v\n", insertTimeWithBloom)
	fmt.Printf("Insertion Time (without Bloom filter): %v\n", insertTimeWithoutBloom)
	fmt.Printf("Query Time (with Bloom filter): %v\n", queryTimeWithBloom)
	fmt.Printf("Query Time (without Bloom filter): %v\n", queryTimeWithoutBloom)
	fmt.Printf("Query Speedup with Bloom filter: %.2fx\n", 
		float64(queryTimeWithoutBloom)/float64(queryTimeWithBloom))
	
	// Log results to the test output
	t.Logf("Performance Test Results (numKeys=%d, numQueries=%d, branchingFactor=%d):", 
		numKeys, numQueries, branchingFactor)
	t.Logf("Insertion Time (with Bloom filter): %v", insertTimeWithBloom)
	t.Logf("Insertion Time (without Bloom filter): %v", insertTimeWithoutBloom)
	t.Logf("Query Time (with Bloom filter): %v", queryTimeWithBloom)
	t.Logf("Query Time (without Bloom filter): %v", queryTimeWithoutBloom)
	t.Logf("Query Speedup with Bloom filter: %.2fx", 
		float64(queryTimeWithoutBloom)/float64(queryTimeWithBloom))
}

// TestPerformanceWithDifferentBloomFilterSizes tests the performance with different Bloom filter sizes
func TestPerformanceWithDifferentBloomFilterSizes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}
	
	// Create a standalone performance tester that can be run from the command line
	fmt.Println("Running performance test with different Bloom filter sizes...")
	
	// Parameters
	numKeys := 1000000
	numQueries := 100000
	branchingFactor := 256
	
	// Generate random keys for insertion
	keys := make([]uint64, numKeys)
	for i := 0; i < numKeys; i++ {
		keys[i] = uint64(rand.Intn(numKeys * 10))
	}
	
	// Generate random keys for queries (50% existing, 50% non-existing)
	queryKeys := make([]uint64, numQueries)
	for i := 0; i < numQueries; i++ {
		if rand.Intn(2) == 0 {
			// Existing key
			queryKeys[i] = keys[rand.Intn(len(keys))]
		} else {
			// Non-existing key
			queryKeys[i] = uint64(numKeys*10 + rand.Intn(numKeys*10))
		}
	}
	
	// Test different false positive rates
	falsePositiveRates := []float64{0.1, 0.01, 0.001, 0.0001}
	
	for _, fpr := range falsePositiveRates {
		// Calculate optimal Bloom filter size for this false positive rate
		size, hashFunctions := OptimalBloomFilterSize(numKeys, fpr)
		
		// Create a tree with this Bloom filter size
		tree := NewBPlusTree(branchingFactor)
		tree.bloomFilter = NewBloomFilter(size, hashFunctions)
		
		// Insert keys
		for _, key := range keys {
			tree.Insert(key)
		}
		
		// Force Bloom filter computation
		tree.Contains(keys[0])
		
		// Measure query time
		startTime := time.Now()
		for _, key := range queryKeys {
			tree.Contains(key)
		}
		queryTime := time.Since(startTime)
		
		fmt.Printf("False Positive Rate: %.4f, Bloom Filter Size: %d bits, Hash Functions: %d, Query Time: %v\n",
			fpr, size, hashFunctions, queryTime)
		t.Logf("False Positive Rate: %.4f, Bloom Filter Size: %d bits, Hash Functions: %d, Query Time: %v",
			fpr, size, hashFunctions, queryTime)
	}
}

// TestMain is a standalone performance tester that can be run from the command line
func TestMain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}
	
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
	
	// Run the performance tests
	TestPerformance(t)
	TestPerformanceWithDifferentBloomFilterSizes(t)
}
