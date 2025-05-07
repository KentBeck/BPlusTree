# Generic B+ Tree Implementation

This package provides a generic B+ tree implementation that can work with any type of key, not just uint64. The implementation uses Go's generics to provide type safety and flexibility.

## Features

- **Generic Keys**: The B+ tree can work with any type of key, not just uint64.
- **Type Safety**: The generic implementation provides compile-time type checking for keys.
- **Range Queries**: The B+ tree supports range queries, which return all keys in a given range.
- **Bloom Filter**: The B+ tree uses a Bloom filter to optimize lookups of non-existent keys.
- **Sorted Output**: The set implementation provides a method to get all elements in sorted order.

## Usage

### Creating a B+ Tree

```go
// Create a B+ tree with uint64 keys
tree := NewBPlusTree[uint64](
    256, // branching factor
    func(a, b uint64) bool { return a < b }, // less function
    func(a, b uint64) bool { return a == b }, // equal function
    func(v uint64) uint64 { return v }, // hash function
)

// Create a B+ tree with int keys
tree := NewBPlusTree[int](
    256, // branching factor
    func(a, b int) bool { return a < b }, // less function
    func(a, b int) bool { return a == b }, // equal function
    func(v int) uint64 { return uint64(v) }, // hash function
)

// Create a B+ tree with string keys
tree := NewBPlusTree[string](
    256, // branching factor
    func(a, b string) bool { return a < b }, // less function
    func(a, b string) bool { return a == b }, // equal function
    func(s string) uint64 { // hash function
        var hash uint64
        for i := 0; i < len(s); i++ {
            hash = hash*31 + uint64(s[i])
        }
        return hash
    },
)
```

### Using the Set Interface

```go
// Create a set of uint64 values
set := NewUint64Set(256)

// Create a set of int values
set := NewIntSet(256)

// Create a set of string values
set := NewStringSet(256)

// Add values to the set
set.Add(10)
set.Add(20)
set.Add(30)

// Check if the set contains a value
if set.Contains(20) {
    fmt.Println("Set contains 20")
}

// Delete a value from the set
set.Delete(20)

// Get all values in the set
values := set.GetAll()

// Get all values in a range
rangeValues := set.Range(15, 25)

// Get a sorted slice of all values
sortedValues := set.SortedSlice()
```

## Implementation Details

The generic B+ tree implementation consists of the following components:

- **Node[K]**: A generic interface for nodes in the B+ tree.
- **LeafNode[K]**: A leaf node that stores keys of type K.
- **BranchNode[K]**: An internal node that stores keys of type K and pointers to child nodes.
- **BPlusTree[K]**: The B+ tree itself, which uses the generic nodes.
- **Set[K]**: A high-level interface for using the B+ tree as a set.

The implementation uses Go's generics to provide type safety and flexibility. The B+ tree can work with any type of key, as long as you provide functions for comparing keys and hashing them for the Bloom filter.

## Benefits over a Non-Generic Implementation

1. **Type Safety**: The generic implementation provides compile-time type checking for keys.
2. **Flexibility**: The B+ tree can work with any type of key, not just uint64.
3. **No Conversion Overhead**: The generic implementation doesn't need to convert between the key type and uint64, which can be expensive for complex types.
4. **Range Queries**: The B+ tree supports range queries, which return all keys in a given range.
5. **Maintainability**: The code is cleaner and more maintainable without conversion functions.

## Performance Considerations

The generic implementation may have slightly more overhead than a specialized implementation for a specific type, but the difference is usually negligible. The Bloom filter optimization helps to minimize the performance impact of lookups for non-existent keys.

## Example

```go
package main

import (
    "fmt"
    "bplustree/pkg/genericbplustree"
)

func main() {
    // Create a set of integers
    set := genericbplustree.NewIntSet(4)

    // Add some values
    set.Add(10)
    set.Add(20)
    set.Add(30)

    // Check if the set contains a value
    fmt.Println(set.Contains(20)) // Output: true
    fmt.Println(set.Contains(40)) // Output: false

    // Get a range of values
    fmt.Println(set.Range(15, 35)) // Output: [20 30]

    // Delete a value
    set.Delete(20)
    fmt.Println(set.Contains(20)) // Output: false

    // Get a sorted slice of all values
    fmt.Println(set.SortedSlice()) // Output: [10 30]
}
```
