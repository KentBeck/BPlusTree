package main

import (
	"bplustree/pkg/bplustree"
	"fmt"
	"math/rand"
	"time"
)

// FakeBloomFilter is a Bloom filter that always returns true for Contains
type FakeBloomFilter struct{}

func NewFakeBloomFilter() *FakeBloomFilter {
	return &FakeBloomFilter{}
}

func (f *FakeBloomFilter) Add(key uint64) {
	// Do nothing
}

func (f *FakeBloomFilter) Contains(key uint64) bool {
	return true // Always return true (maybe in set)
}

func (f *FakeBloomFilter) Clear() {
	// Do nothing
}

func (f *FakeBloomFilter) SetValid() {
	// Do nothing
}

func (f *FakeBloomFilter) IsValid() bool {
	return true
}

func main() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Parameters
	numKeys := 1000000
	numQueries := 1000000
	branchingFactor := 256

	fmt.Printf("Performance Test (numKeys=%d, numQueries=%d, branchingFactor=%d)\n",
		numKeys, numQueries, branchingFactor)

	// Generate random keys for insertion
	fmt.Println("Generating random keys...")
	keys := make([]uint64, numKeys)
	for i := 0; i < numKeys; i++ {
		keys[i] = uint64(rand.Intn(numKeys * 10)) // Use a larger range to have some duplicates
	}

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

	// Create a tree with a real Bloom filter
	fmt.Println("\nTesting with real Bloom filter...")
	treeWithRealBloom := bplustree.NewBPlusTree(branchingFactor)

	// Insert keys
	fmt.Println("Inserting keys...")
	startTime := time.Now()
	for _, key := range keys {
		treeWithRealBloom.Insert(key)
	}
	insertTimeReal := time.Since(startTime)
	fmt.Printf("Insertion Time: %v (%.2f keys/sec)\n",
		insertTimeReal, float64(numKeys)/insertTimeReal.Seconds())

	// Force Bloom filter computation
	treeWithRealBloom.Contains(keys[0])

	// Measure query time
	fmt.Println("Querying keys...")
	startTime = time.Now()
	hitsReal := 0
	for _, key := range queryKeys {
		if treeWithRealBloom.Contains(key) {
			hitsReal++
		}
	}
	queryTimeReal := time.Since(startTime)

	// Print results
	fmt.Printf("Query Time: %v (%.2f queries/sec)\n",
		queryTimeReal, float64(numQueries)/queryTimeReal.Seconds())
	fmt.Printf("Hits: %d (%.2f%%)\n", hitsReal, float64(hitsReal*100)/float64(numQueries))

	// Create a tree with a fake Bloom filter (always returns true)
	fmt.Println("\nTesting with fake Bloom filter (always returns true)...")
	treeWithFakeBloom := bplustree.NewBPlusTree(branchingFactor)
	treeWithFakeBloom.SetCustomBloomFilter(NewFakeBloomFilter())

	// Insert keys
	fmt.Println("Inserting keys...")
	startTime = time.Now()
	for _, key := range keys {
		treeWithFakeBloom.Insert(key)
	}
	insertTimeFake := time.Since(startTime)
	fmt.Printf("Insertion Time: %v (%.2f keys/sec)\n",
		insertTimeFake, float64(numKeys)/insertTimeFake.Seconds())

	// Measure query time
	fmt.Println("Querying keys...")
	startTime = time.Now()
	hitsFake := 0
	for _, key := range queryKeys {
		if treeWithFakeBloom.Contains(key) {
			hitsFake++
		}
	}
	queryTimeFake := time.Since(startTime)

	// Print results
	fmt.Printf("Query Time: %v (%.2f queries/sec)\n",
		queryTimeFake, float64(numQueries)/queryTimeFake.Seconds())
	fmt.Printf("Hits: %d (%.2f%%)\n", hitsFake, float64(hitsFake*100)/float64(numQueries))

	// Create a tree with no Bloom filter (direct tree traversal)
	fmt.Println("\nTesting with no Bloom filter (direct tree traversal)...")
	treeWithNoBloom := bplustree.NewBPlusTree(branchingFactor)
	treeWithNoBloom.DisableBloomFilter()

	// Insert keys
	fmt.Println("Inserting keys...")
	startTime = time.Now()
	for _, key := range keys {
		treeWithNoBloom.Insert(key)
	}
	insertTimeNo := time.Since(startTime)
	fmt.Printf("Insertion Time: %v (%.2f keys/sec)\n",
		insertTimeNo, float64(numKeys)/insertTimeNo.Seconds())

	// Measure query time
	fmt.Println("Querying keys...")
	startTime = time.Now()
	hitsNo := 0
	for _, key := range queryKeys {
		if treeWithNoBloom.Contains(key) {
			hitsNo++
		}
	}
	queryTimeNo := time.Since(startTime)

	// Print results
	fmt.Printf("Query Time: %v (%.2f queries/sec)\n",
		queryTimeNo, float64(numQueries)/queryTimeNo.Seconds())
	fmt.Printf("Hits: %d (%.2f%%)\n", hitsNo, float64(hitsNo*100)/float64(numQueries))

	// Compare results
	fmt.Println("\nComparison:")
	fmt.Printf("Real Bloom Filter Query Time: %v\n", queryTimeReal)
	fmt.Printf("Fake Bloom Filter Query Time: %v\n", queryTimeFake)
	fmt.Printf("No Bloom Filter Query Time: %v\n", queryTimeNo)

	// Calculate speedups
	realVsNo := float64(queryTimeNo) / float64(queryTimeReal)
	fakeVsNo := float64(queryTimeNo) / float64(queryTimeFake)

	fmt.Printf("Real Bloom Filter Speedup vs No Bloom Filter: %.2fx\n", realVsNo)
	fmt.Printf("Fake Bloom Filter Speedup vs No Bloom Filter: %.2fx\n", fakeVsNo)

	// Calculate overhead of Bloom filter
	bloomOverhead := float64(queryTimeReal)/float64(queryTimeFake) - 1.0
	fmt.Printf("Bloom Filter Overhead: %.2f%%\n", bloomOverhead*100)

	// Create a tree with a fake Bloom filter (always returns true)
	fmt.Println("\nTesting with fake Bloom filter (always returns true)...")
	treeWithFakeBloom := bplustree.NewBPlusTree(branchingFactor)
	treeWithFakeBloom.SetCustomBloomFilter(NewFakeBloomFilter())

	// Insert keys
	fmt.Println("Inserting keys...")
	startTime = time.Now()
	for _, key := range keys {
		treeWithFakeBloom.Insert(key)
	}
	insertTimeFake := time.Since(startTime)
	fmt.Printf("Insertion Time: %v (%.2f keys/sec)\n",
		insertTimeFake, float64(numKeys)/insertTimeFake.Seconds())

	// Measure query time
	fmt.Println("Querying keys...")
	startTime = time.Now()
	hitsFake := 0
	for _, key := range queryKeys {
		if treeWithFakeBloom.Contains(key) {
			hitsFake++
		}
	}
	queryTimeFake := time.Since(startTime)

	// Print results
	fmt.Printf("Query Time: %v (%.2f queries/sec)\n",
		queryTimeFake, float64(numQueries)/queryTimeFake.Seconds())
	fmt.Printf("Hits: %d (%.2f%%)\n", hitsFake, float64(hitsFake*100)/float64(numQueries))

	// Create a tree with no Bloom filter (direct tree traversal)
	fmt.Println("\nTesting with no Bloom filter (direct tree traversal)...")
	treeWithNoBloom := bplustree.NewBPlusTree(branchingFactor)
	treeWithNoBloom.DisableBloomFilter()

	// Insert keys
	fmt.Println("Inserting keys...")
	startTime = time.Now()
	for _, key := range keys {
		treeWithNoBloom.Insert(key)
	}
	insertTimeNo := time.Since(startTime)
	fmt.Printf("Insertion Time: %v (%.2f keys/sec)\n",
		insertTimeNo, float64(numKeys)/insertTimeNo.Seconds())

	// Measure query time
	fmt.Println("Querying keys...")
	startTime = time.Now()
	hitsNo := 0
	for _, key := range queryKeys {
		if treeWithNoBloom.Contains(key) {
			hitsNo++
		}
	}
	queryTimeNo := time.Since(startTime)

	// Print results
	fmt.Printf("Query Time: %v (%.2f queries/sec)\n",
		queryTimeNo, float64(numQueries)/queryTimeNo.Seconds())
	fmt.Printf("Hits: %d (%.2f%%)\n", hitsNo, float64(hitsNo*100)/float64(numQueries))

	// Compare results
	fmt.Println("\nComparison:")
	fmt.Printf("Real Bloom Filter Query Time: %v\n", queryTimeReal)
	fmt.Printf("Fake Bloom Filter Query Time: %v\n", queryTimeFake)
	fmt.Printf("No Bloom Filter Query Time: %v\n", queryTimeNo)

	// Calculate speedups
	realVsNo := float64(queryTimeNo) / float64(queryTimeReal)
	fakeVsNo := float64(queryTimeNo) / float64(queryTimeFake)

	fmt.Printf("Real Bloom Filter Speedup vs No Bloom Filter: %.2fx\n", realVsNo)
	fmt.Printf("Fake Bloom Filter Speedup vs No Bloom Filter: %.2fx\n", fakeVsNo)

	// Calculate overhead of Bloom filter
	bloomOverhead := float64(queryTimeReal)/float64(queryTimeFake) - 1.0
	fmt.Printf("Bloom Filter Overhead: %.2f%%\n", bloomOverhead*100)

	// Test with different percentages of non-existent keys
	fmt.Println("\nTesting with different percentages of non-existent keys...")
	nonExistentPercentages := []int{0, 25, 50, 75, 100}

	for _, percentage := range nonExistentPercentages {
		// Generate query keys with the specified percentage of non-existent keys
		fmt.Printf("\nTesting with %d%% non-existent keys...\n", percentage)
		testQueryKeys := make([]uint64, numQueries)
		for i := 0; i < numQueries; i++ {
			if rand.Intn(100) < percentage {
				// Non-existing key
				testQueryKeys[i] = uint64(numKeys*10 + rand.Intn(numKeys*10))
			} else {
				// Existing key
				testQueryKeys[i] = keys[rand.Intn(len(keys))]
			}
		}

		// Test with real Bloom filter
		startTime = time.Now()
		hitsReal = 0
		for _, key := range testQueryKeys {
			if treeWithRealBloom.Contains(key) {
				hitsReal++
			}
		}
		realTime := time.Since(startTime)

		// Test with fake Bloom filter
		startTime = time.Now()
		hitsFake = 0
		for _, key := range testQueryKeys {
			if treeWithFakeBloom.Contains(key) {
				hitsFake++
			}
		}
		fakeTime := time.Since(startTime)

		// Test with no Bloom filter
		startTime = time.Now()
		hitsNo = 0
		for _, key := range testQueryKeys {
			if treeWithNoBloom.Contains(key) {
				hitsNo++
			}
		}
		noTime := time.Since(startTime)

		// Print results
		fmt.Printf("Real Bloom Filter: %v, Hits: %d (%.2f%%)\n",
			realTime, hitsReal, float64(hitsReal*100)/float64(numQueries))
		fmt.Printf("Fake Bloom Filter: %v, Hits: %d (%.2f%%)\n",
			fakeTime, hitsFake, float64(hitsFake*100)/float64(numQueries))
		fmt.Printf("No Bloom Filter: %v, Hits: %d (%.2f%%)\n",
			noTime, hitsNo, float64(hitsNo*100)/float64(numQueries))

		// Calculate speedups
		realVsNo = float64(noTime) / float64(realTime)
		fakeVsNo = float64(noTime) / float64(fakeTime)

		fmt.Printf("Real Bloom Filter Speedup: %.2fx\n", realVsNo)
		fmt.Printf("Fake Bloom Filter Speedup: %.2fx\n", fakeVsNo)

		// Calculate overhead
		bloomOverhead = float64(realTime)/float64(fakeTime) - 1.0
		fmt.Printf("Bloom Filter Overhead: %.2f%%\n", bloomOverhead*100)
	}
}
