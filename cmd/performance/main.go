package main

import (
	"bplustree/pkg/bplustree"
	"fmt"
	"math/rand"
	"time"
)

func main() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
	
	// Parameters
	numKeys := 1000000
	numQueries := 1000000
	branchingFactor := 256
	
	fmt.Printf("Performance Test (numKeys=%d, numQueries=%d, branchingFactor=%d)\n", 
		numKeys, numQueries, branchingFactor)
	
	// Create trees with and without Bloom filter
	treeWithBloom := bplustree.NewBPlusTree(branchingFactor)
	
	// Generate random keys for insertion
	fmt.Println("Generating random keys...")
	keys := make([]uint64, numKeys)
	for i := 0; i < numKeys; i++ {
		keys[i] = uint64(rand.Intn(numKeys * 10)) // Use a larger range to have some duplicates
	}
	
	// Measure insertion time
	fmt.Println("Inserting keys...")
	startTime := time.Now()
	for _, key := range keys {
		treeWithBloom.Insert(key)
	}
	insertTime := time.Since(startTime)
	fmt.Printf("Insertion Time: %v (%.2f keys/sec)\n", 
		insertTime, float64(numKeys)/insertTime.Seconds())
	
	// Generate random keys for queries (50% existing, 50% non-existing)
	fmt.Println("Generating query keys...")
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
	
	// Force Bloom filter computation
	treeWithBloom.Contains(keys[0])
	
	// Measure query time
	fmt.Println("Querying keys...")
	startTime = time.Now()
	hits := 0
	for _, key := range queryKeys {
		if treeWithBloom.Contains(key) {
			hits++
		}
	}
	queryTime := time.Since(startTime)
	
	// Print results
	fmt.Printf("Query Time: %v (%.2f queries/sec)\n", 
		queryTime, float64(numQueries)/queryTime.Seconds())
	fmt.Printf("Hits: %d (%.2f%%)\n", hits, float64(hits*100)/float64(numQueries))
	
	// Test different Bloom filter sizes
	fmt.Println("\nTesting different Bloom filter sizes...")
	
	// Test different false positive rates
	falsePositiveRates := []float64{0.1, 0.01, 0.001, 0.0001}
	
	for _, fpr := range falsePositiveRates {
		// Calculate optimal Bloom filter size for this false positive rate
		size, hashFunctions := bplustree.OptimalBloomFilterSize(numKeys, fpr)
		
		// Create a tree with this Bloom filter size
		tree := bplustree.NewBPlusTree(branchingFactor)
		tree.SetBloomFilterParams(size, hashFunctions)
		
		// Insert keys
		for _, key := range keys {
			tree.Insert(key)
		}
		
		// Force Bloom filter computation
		tree.Contains(keys[0])
		
		// Measure query time
		startTime := time.Now()
		hits := 0
		for _, key := range queryKeys {
			if tree.Contains(key) {
				hits++
			}
		}
		queryTime := time.Since(startTime)
		
		fmt.Printf("False Positive Rate: %.4f, Bloom Filter Size: %d bits, Hash Functions: %d\n",
			fpr, size, hashFunctions)
		fmt.Printf("  Query Time: %v (%.2f queries/sec)\n", 
			queryTime, float64(numQueries)/queryTime.Seconds())
		fmt.Printf("  Hits: %d (%.2f%%)\n", hits, float64(hits*100)/float64(numQueries))
	}
}
