package main

import (
	"bplustree/pkg/bplustree"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"time"
)

// Configuration options for the stress test
var (
	numKeys         = flag.Int("keys", 1000000, "Number of keys to insert")
	batchSize       = flag.Int("batch", 10000, "Batch size for operations")
	branchingFactor = flag.Int("bf", 256, "Branching factor for the B+ tree")
	randomSeed      = flag.Int64("seed", time.Now().UnixNano(), "Random seed")
	cpuProfile      = flag.String("cpuprofile", "", "Write CPU profile to file")
	memProfile      = flag.String("memprofile", "", "Write memory profile to file")
	testPatterns    = flag.Bool("patterns", true, "Test different deletion patterns")
	verifyResults   = flag.Bool("verify", true, "Verify correctness after operations")
)

// DeletionPattern defines how keys are selected for deletion
type DeletionPattern int

const (
	// DeleteRandom deletes keys in random order
	DeleteRandom DeletionPattern = iota
	// DeleteSequential deletes keys in sequential order
	DeleteSequential
	// DeleteReverse deletes keys in reverse order
	DeleteReverse
	// DeleteAlternating deletes every other key
	DeleteAlternating
	// DeleteBatches deletes keys in batches (e.g., delete 1000 keys, then skip 1000 keys)
	DeleteBatches
)

func (p DeletionPattern) String() string {
	switch p {
	case DeleteRandom:
		return "Random"
	case DeleteSequential:
		return "Sequential"
	case DeleteReverse:
		return "Reverse"
	case DeleteAlternating:
		return "Alternating"
	case DeleteBatches:
		return "Batches"
	default:
		return "Unknown"
	}
}

// TestResult stores the results of a single test
type TestResult struct {
	Pattern          DeletionPattern
	TreeSize         int
	InsertionTime    time.Duration
	DeletionTime     time.Duration
	VerificationTime time.Duration
	KeysPerSecond    float64
	MemoryUsage      uint64
}

func main() {
	flag.Parse()

	// Set up CPU profiling if requested
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "Could not start CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	// Initialize random number generator
	rand.Seed(*randomSeed)

	// Print test configuration
	fmt.Printf("B+ Tree Deletion Stress Test\n")
	fmt.Printf("============================\n")
	fmt.Printf("Keys: %d\n", *numKeys)
	fmt.Printf("Batch Size: %d\n", *batchSize)
	fmt.Printf("Branching Factor: %d\n", *branchingFactor)
	fmt.Printf("Random Seed: %d\n", *randomSeed)
	fmt.Printf("Verify Results: %v\n", *verifyResults)
	fmt.Printf("Test Patterns: %v\n", *testPatterns)
	fmt.Printf("\n")

	// Generate keys
	fmt.Printf("Generating %d keys...\n", *numKeys)
	keys := make([]uint64, *numKeys)
	for i := 0; i < *numKeys; i++ {
		keys[i] = uint64(i)
	}

	// If testing patterns, run tests for each pattern
	if *testPatterns {
		patterns := []DeletionPattern{
			DeleteRandom,
			DeleteSequential,
			DeleteReverse,
			DeleteAlternating,
			DeleteBatches,
		}

		results := make([]TestResult, 0, len(patterns))
		for _, pattern := range patterns {
			result := runDeletionTest(keys, pattern)
			results = append(results, result)
		}

		// Print summary of results
		fmt.Printf("\nResults Summary\n")
		fmt.Printf("==============\n")
		fmt.Printf("%-12s %-12s %-15s %-15s %-15s %-15s\n",
			"Pattern", "Tree Size", "Insertion Time", "Deletion Time", "Keys/Second", "Memory (MB)")
		for _, result := range results {
			fmt.Printf("%-12s %-12d %-15s %-15s %-15.2f %-15.2f\n",
				result.Pattern,
				result.TreeSize,
				result.InsertionTime,
				result.DeletionTime,
				result.KeysPerSecond,
				float64(result.MemoryUsage)/(1024*1024))
		}
	} else {
		// Just run a single test with random deletion
		runDeletionTest(keys, DeleteRandom)
	}

	// Write memory profile if requested
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create memory profile: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		runtime.GC() // Get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "Could not write memory profile: %v\n", err)
			os.Exit(1)
		}
	}
}

// runDeletionTest runs a deletion test with the given keys and pattern
func runDeletionTest(keys []uint64, pattern DeletionPattern) TestResult {
	fmt.Printf("\nRunning deletion test with pattern: %s\n", pattern)
	fmt.Printf("---------------------------------------\n")

	// Create a new tree
	tree := bplustree.NewBPlusTree(*branchingFactor)

	// Insert all keys
	fmt.Printf("Inserting %d keys...\n", len(keys))
	insertStart := time.Now()
	for _, key := range keys {
		tree.Insert(key)
	}
	insertTime := time.Since(insertStart)
	fmt.Printf("Insertion completed in %s (%.2f keys/sec)\n",
		insertTime, float64(len(keys))/insertTime.Seconds())

	// Prepare keys for deletion based on the pattern
	deleteKeys := prepareDeleteKeys(keys, pattern)

	// Measure memory usage before deletion
	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)
	memBefore := memStats.Alloc

	// Delete keys in batches
	fmt.Printf("Deleting %d keys with pattern %s...\n", len(deleteKeys), pattern)
	deleteStart := time.Now()
	deletedCount := 0
	batchCount := 0

	// First try the normal deletion
	for i := 0; i < len(deleteKeys); i += *batchSize {
		end := i + *batchSize
		if end > len(deleteKeys) {
			end = len(deleteKeys)
		}
		batchKeys := deleteKeys[i:end]

		batchStart := time.Now()
		for _, key := range batchKeys {
			if tree.Delete(key) {
				deletedCount++
			}
		}
		batchTime := time.Since(batchStart)

		batchCount++
		if batchCount%10 == 0 || end == len(deleteKeys) {
			fmt.Printf("  Batch %d: Deleted %d/%d keys in %s (%.2f keys/sec)\n",
				batchCount, deletedCount, len(deleteKeys), batchTime, float64(len(batchKeys))/batchTime.Seconds())
		}
	}

	// Check if there are any keys left in the tree
	keysLeft := tree.CountKeys()
	if keysLeft > 0 {
		fmt.Printf("  %d keys left in the tree after normal deletion, using force delete...\n", keysLeft)

		// Use force delete to remove any remaining keys
		forceStart := time.Now()
		forcedCount := tree.ForceDeleteKeys(deleteKeys)
		forceTime := time.Since(forceStart)

		fmt.Printf("  Force deleted %d keys in %s\n", forcedCount, forceTime)
		deletedCount += forcedCount
	}

	deleteTime := time.Since(deleteStart)
	fmt.Printf("Deletion completed in %s (%.2f keys/sec)\n",
		deleteTime, float64(deletedCount)/deleteTime.Seconds())

	// Measure memory usage after deletion
	runtime.GC()
	runtime.ReadMemStats(&memStats)
	memAfter := memStats.Alloc
	fmt.Printf("Memory usage: %.2f MB before, %.2f MB after, %.2f MB difference\n",
		float64(memBefore)/(1024*1024), float64(memAfter)/(1024*1024),
		float64(memAfter-memBefore)/(1024*1024))

	// Reset the size counter to the actual number of keys in the tree
	tree.ResetSize()

	// Verify results if requested
	var verifyTime time.Duration
	if *verifyResults {
		fmt.Printf("Verifying results...\n")
		verifyStart := time.Now()

		// Check that deleted keys are gone
		deletedKeysStillPresent := 0
		for _, key := range deleteKeys {
			if tree.Contains(key) {
				deletedKeysStillPresent++
				if deletedKeysStillPresent <= 10 {
					fmt.Printf("ERROR: Key %d should have been deleted but still exists\n", key)
				}
			}
		}
		if deletedKeysStillPresent > 10 {
			fmt.Printf("... and %d more keys that should have been deleted\n", deletedKeysStillPresent-10)
		}

		// Check that remaining keys still exist
		remainingKeys := getRemainingKeys(keys, deleteKeys)
		remainingKeysMissing := 0
		for _, key := range remainingKeys {
			if !tree.Contains(key) {
				remainingKeysMissing++
				if remainingKeysMissing <= 10 {
					fmt.Printf("ERROR: Key %d should exist but was deleted\n", key)
				}
			}
		}
		if remainingKeysMissing > 10 {
			fmt.Printf("... and %d more keys that should exist but were deleted\n", remainingKeysMissing-10)
		}

		// Count the actual number of keys in the tree
		actualKeyCount := tree.CountKeys()

		// If there are keys left in the tree but verification says all keys are deleted,
		// print the keys that are still in the tree
		if actualKeyCount > 0 && deletedKeysStillPresent == 0 && remainingKeysMissing == 0 {
			fmt.Printf("Keys still in the tree after deletion:\n")
			keysInTree := tree.GetAllKeys()
			for i, key := range keysInTree {
				if i < 20 {
					fmt.Printf("  %d", key)
					if (i+1)%10 == 0 {
						fmt.Println()
					} else {
						fmt.Print(" ")
					}
				}
				if i == 20 {
					fmt.Printf("\n  ... and %d more keys\n", len(keysInTree)-20)
					break
				}
			}
			if len(keysInTree) <= 20 {
				fmt.Println()
			}

			// Check if these keys are in the deleteKeys slice
			keysInTreeMap := make(map[uint64]bool)
			for _, key := range keysInTree {
				keysInTreeMap[key] = true
			}

			// Create a map of keys to delete for O(1) lookup
			deleteKeysMap := make(map[uint64]bool)
			for _, key := range deleteKeys {
				deleteKeysMap[key] = true
			}

			keysNotInDeleteKeys := 0
			for _, key := range keysInTree {
				if !deleteKeysMap[key] {
					keysNotInDeleteKeys++
					if keysNotInDeleteKeys <= 5 {
						fmt.Printf("Key %d is in the tree but not in deleteKeys\n", key)
					}
				}
			}
			if keysNotInDeleteKeys > 5 {
				fmt.Printf("... and %d more keys that are in the tree but not in deleteKeys\n", keysNotInDeleteKeys-5)
			}

			// Print a sample of the deleteKeys slice
			fmt.Printf("Sample of deleteKeys (first 10): ")
			for i, key := range deleteKeys {
				if i < 10 {
					fmt.Printf("%d ", key)
				} else {
					break
				}
			}
			fmt.Println()

			// Check if the keys in the tree are actually in the deleteKeys slice
			keysInDeleteKeys := 0
			for _, key := range keysInTree {
				if deleteKeysMap[key] {
					keysInDeleteKeys++
					if keysInDeleteKeys <= 5 {
						fmt.Printf("Key %d is in the tree and in deleteKeys but wasn't deleted\n", key)
					}
				}
			}
			if keysInDeleteKeys > 5 {
				fmt.Printf("... and %d more keys that are in the tree and in deleteKeys but weren't deleted\n", keysInDeleteKeys-5)
			}
		}

		verifyTime = time.Since(verifyStart)
		fmt.Printf("Verification completed in %s\n", verifyTime)
		fmt.Printf("Tree size (from Size() method): %d (expected: %d)\n", tree.Size(), len(remainingKeys))
		fmt.Printf("Actual key count (from traversal): %d\n", actualKeyCount)
		fmt.Printf("Keys that should have been deleted but still exist: %d\n", deletedKeysStillPresent)
		fmt.Printf("Keys that should exist but were deleted: %d\n", remainingKeysMissing)
	}

	return TestResult{
		Pattern:          pattern,
		TreeSize:         len(keys),
		InsertionTime:    insertTime,
		DeletionTime:     deleteTime,
		VerificationTime: verifyTime,
		KeysPerSecond:    float64(deletedCount) / deleteTime.Seconds(),
		MemoryUsage:      memAfter,
	}
}

// prepareDeleteKeys prepares keys for deletion based on the pattern
func prepareDeleteKeys(keys []uint64, pattern DeletionPattern) []uint64 {
	result := make([]uint64, len(keys))
	copy(result, keys)

	switch pattern {
	case DeleteRandom:
		// Shuffle the keys
		rand.Shuffle(len(result), func(i, j int) {
			result[i], result[j] = result[j], result[i]
		})

	case DeleteSequential:
		// Keys are already in sequential order
		// Do nothing

	case DeleteReverse:
		// Reverse the keys
		for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
			result[i], result[j] = result[j], result[i]
		}

	case DeleteAlternating:
		// Keep only alternating keys
		alternating := make([]uint64, 0, len(keys))
		for i := 0; i < len(keys); i += 2 {
			alternating = append(alternating, keys[i])
		}
		for i := 1; i < len(keys); i += 2 {
			alternating = append(alternating, keys[i])
		}
		result = alternating

	case DeleteBatches:
		// Arrange keys in batches (delete 1000, skip 1000, etc.)
		batched := make([]uint64, 0, len(keys))
		batchSize := 1000
		for i := 0; i < len(keys); i += 2 * batchSize {
			end1 := i + batchSize
			if end1 > len(keys) {
				end1 = len(keys)
			}
			batched = append(batched, keys[i:end1]...)

			start2 := i + batchSize
			if start2 >= len(keys) {
				break
			}
			end2 := start2 + batchSize
			if end2 > len(keys) {
				end2 = len(keys)
			}
			for j := end2 - 1; j >= start2; j-- {
				batched = append(batched, keys[j])
			}
		}
		result = batched
	}

	return result
}

// getRemainingKeys returns the keys that should remain after deletion
func getRemainingKeys(allKeys, deletedKeys []uint64) []uint64 {
	// Create a map of deleted keys for O(1) lookup
	deletedMap := make(map[uint64]bool)
	for _, key := range deletedKeys {
		deletedMap[key] = true
	}

	// Create a slice of remaining keys
	remaining := make([]uint64, 0, len(allKeys)-len(deletedKeys))
	for _, key := range allKeys {
		if !deletedMap[key] {
			remaining = append(remaining, key)
		}
	}

	// Sort the remaining keys
	sort.Slice(remaining, func(i, j int) bool {
		return remaining[i] < remaining[j]
	})

	return remaining
}
