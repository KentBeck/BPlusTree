# Generic B+ Tree Implementation

This package provides a generic B+ tree implementation that can work with any type of key, not just uint64. The implementation is based on the original B+ tree implementation but has been refactored to use Go's generics.

## Features

- **Generic Keys**: The B+ tree can work with any type of key, not just uint64.
- **Type Safety**: The generic implementation provides compile-time type checking for keys.
- **Range Queries**: The B+ tree supports range queries, which return all keys in a given range.
- **Bloom Filter**: The B+ tree uses a Bloom filter to optimize lookups of non-existent keys.

## Usage

### Creating a B+ Tree

```go
// Create a B+ tree with uint64 keys
tree := NewGenericBPlusTree[uint64](
    256, // branching factor
    func(a, b uint64) bool { return a < b }, // less function
    func(a, b uint64) bool { return a == b }, // equal function
    func(v uint64) uint64 { return v }, // hash function
)

// Create a B+ tree with int keys
tree := NewGenericBPlusTree[int](
    256, // branching factor
    func(a, b int) bool { return a < b }, // less function
    func(a, b int) bool { return a == b }, // equal function
    func(v int) uint64 { return uint64(v) }, // hash function
)

// Create a B+ tree with string keys
tree := NewGenericBPlusTree[string](
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

## Performance

The generic implementation is slightly slower than the original implementation due to the overhead of function values for comparisons. However, the difference is not significant for most use cases.

Here are some benchmark results:

```
BenchmarkOriginalBPlusTreeInsert-16                 	 7715008	       142.0 ns/op	       2 B/op	       0 allocs/op
BenchmarkGenericBPlusTreeInsert-16                  	 6375740	       180.1 ns/op	       2 B/op	       0 allocs/op
BenchmarkOriginalBPlusTreeContains-16               	 2710710	       438.3 ns/op	     168 B/op	      21 allocs/op
BenchmarkGenericBPlusTreeContains-16                	 2225780	       530.1 ns/op	     168 B/op	      21 allocs/op
BenchmarkOriginalBPlusTreeContainsNonExistent-16    	 3351202	       355.2 ns/op	     168 B/op	      21 allocs/op
BenchmarkGenericBPlusTreeContainsNonExistent-16     	 2063302	       580.1 ns/op	     168 B/op	      21 allocs/op
BenchmarkGenericBPlusTreeRangeQuery-16              	    2438	    429129 ns/op	 2049983 B/op	      24 allocs/op
BenchmarkGenericSetWithStrings-16                   	 7425721	       156.1 ns/op	       0 B/op	       0 allocs/op
```

## Implementation Details

The generic B+ tree implementation consists of the following components:

- **GenericNode[K]**: A generic interface for nodes in the B+ tree.
- **GenericLeafNode[K]**: A leaf node that stores keys of type K.
- **GenericBranchNode[K]**: An internal node that stores keys of type K and pointers to child nodes.
- **GenericBPlusTree[K]**: The B+ tree itself, which uses the generic nodes.
- **GenericSet[K]**: A high-level interface for using the B+ tree as a set.

The implementation uses Go's generics to provide type safety and flexibility. The B+ tree can work with any type of key, as long as you provide functions for comparing keys and hashing them for the Bloom filter.

## Benefits over the Original Implementation

1. **Type Safety**: The generic implementation provides compile-time type checking for keys.
2. **Flexibility**: The B+ tree can work with any type of key, not just uint64.
3. **Range Queries**: The B+ tree supports range queries, which return all keys in a given range.
4. **No Conversion Overhead**: The generic implementation doesn't need to convert between the key type and uint64, which can be expensive for complex types.
5. **Maintainability**: The code is cleaner and more maintainable without conversion functions.

## Limitations

1. **Performance**: The generic implementation is slightly slower than the original implementation due to the overhead of function values for comparisons.
2. **Bloom Filter**: The Bloom filter still uses uint64 hashes, so you need to provide a hash function for your key type.
3. **Memory Usage**: The generic implementation may use more memory for complex key types.
