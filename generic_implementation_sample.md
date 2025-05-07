# Sample Implementation of Generic B+ Tree Components

This document provides sample code snippets for the core components of the generic B+ tree implementation. These are not complete implementations but serve as a starting point to illustrate the approach.

## 1. Generic Node Interface

```go
package bplustree

// NodeType represents the type of node (leaf or branch)
type NodeType int

const (
    Leaf NodeType = iota
    Branch
)

// GenericNode is a generic interface for nodes in the B+ tree
type GenericNode[K any] interface {
    // Type returns the type of the node
    Type() NodeType

    // Keys returns the keys in the node
    Keys() []K

    // KeyCount returns the number of keys in the node
    KeyCount() int

    // IsFull returns true if the node is full
    IsFull(branchingFactor int) bool

    // IsUnderflow returns true if the node has too few keys
    IsUnderflow(branchingFactor int) bool

    // InsertKey inserts a key into the node
    InsertKey(key K, less func(a, b K) bool) bool

    // DeleteKey deletes a key from the node
    DeleteKey(key K, equal func(a, b K) bool) bool

    // FindKey returns the index of the key in the node, or -1 if not found
    FindKey(key K, equal func(a, b K) bool) int

    // Contains returns true if the node contains the key
    Contains(key K, equal func(a, b K) bool) bool
}
```

## 2. Generic Leaf Node

```go
// GenericLeafNode is a leaf node that stores keys of type K
type GenericLeafNode[K any] struct {
    keys []K
    next *GenericLeafNode[K] // Pointer to the next leaf node for range queries
}

// NewGenericLeafNode creates a new generic leaf node
func NewGenericLeafNode[K any]() *GenericLeafNode[K] {
    return &GenericLeafNode[K]{
        keys: make([]K, 0),
        next: nil,
    }
}

// Type returns the type of the node
func (n *GenericLeafNode[K]) Type() NodeType {
    return Leaf
}

// Keys returns the keys in the node
func (n *GenericLeafNode[K]) Keys() []K {
    return n.keys
}

// KeyCount returns the number of keys in the node
func (n *GenericLeafNode[K]) KeyCount() int {
    return len(n.keys)
}

// IsFull returns true if the node is full
func (n *GenericLeafNode[K]) IsFull(branchingFactor int) bool {
    return len(n.keys) >= branchingFactor
}

// IsUnderflow returns true if the node has too few keys
func (n *GenericLeafNode[K]) IsUnderflow(branchingFactor int) bool {
    // For leaf nodes, minimum number of keys is ceil(m/2)
    return len(n.keys) < (branchingFactor+1)/2
}

// InsertKey inserts a key into the node
func (n *GenericLeafNode[K]) InsertKey(key K, less func(a, b K) bool) bool {
    // Find position to insert using binary search
    pos := sort.Search(len(n.keys), func(i int) bool {
        return !less(n.keys[i], key)
    })

    // Check if key already exists
    if pos < len(n.keys) && !less(n.keys[pos], key) && !less(key, n.keys[pos]) {
        return false // Key already exists
    }

    // Insert key
    n.keys = append(n.keys, *new(K)) // Add zero value of K
    copy(n.keys[pos+1:], n.keys[pos:])
    n.keys[pos] = key
    return true
}

// DeleteKey deletes a key from the node
func (n *GenericLeafNode[K]) DeleteKey(key K, equal func(a, b K) bool) bool {
    pos := n.FindKey(key, equal)
    if pos == -1 {
        return false
    }

    // Remove key
    copy(n.keys[pos:], n.keys[pos+1:])
    n.keys = n.keys[:len(n.keys)-1]
    return true
}

// FindKey returns the index of the key in the node, or -1 if not found
func (n *GenericLeafNode[K]) FindKey(key K, equal func(a, b K) bool) int {
    for i, k := range n.keys {
        if equal(k, key) {
            return i
        }
    }
    return -1
}

// Contains returns true if the node contains the key
func (n *GenericLeafNode[K]) Contains(key K, equal func(a, b K) bool) bool {
    return n.FindKey(key, equal) != -1
}
```

## 3. Generic Branch Node

```go
// GenericBranchNode is an internal node that stores keys of type K
type GenericBranchNode[K any] struct {
    keys     []K
    children []GenericNode[K]
}

// NewGenericBranchNode creates a new generic branch node
func NewGenericBranchNode[K any]() *GenericBranchNode[K] {
    return &GenericBranchNode[K]{
        keys:     make([]K, 0),
        children: make([]GenericNode[K], 0),
    }
}

// Type returns the type of the node
func (n *GenericBranchNode[K]) Type() NodeType {
    return Branch
}

// Keys returns the keys in the node
func (n *GenericBranchNode[K]) Keys() []K {
    return n.keys
}

// Children returns the children of the node
func (n *GenericBranchNode[K]) Children() []GenericNode[K] {
    return n.children
}

// KeyCount returns the number of keys in the node
func (n *GenericBranchNode[K]) KeyCount() int {
    return len(n.keys)
}

// IsFull returns true if the node is full
func (n *GenericBranchNode[K]) IsFull(branchingFactor int) bool {
    return len(n.keys) >= branchingFactor-1
}

// IsUnderflow returns true if the node has too few keys
func (n *GenericBranchNode[K]) IsUnderflow(branchingFactor int) bool {
    // For internal nodes, minimum number of keys is ceil(m/2)-1
    return len(n.keys) < (branchingFactor+1)/2 - 1
}

// InsertKeyWithChild inserts a key and child into the node at the correct position
func (n *GenericBranchNode[K]) InsertKeyWithChild(key K, child GenericNode[K], less func(a, b K) bool) {
    // Find position to insert using binary search
    pos := sort.Search(len(n.keys), func(i int) bool {
        return !less(n.keys[i], key)
    })

    // Insert key
    n.keys = append(n.keys, *new(K)) // Add zero value of K
    copy(n.keys[pos+1:], n.keys[pos:])
    n.keys[pos] = key

    // Insert child (goes to the right of the key)
    n.children = append(n.children, nil)
    copy(n.children[pos+2:], n.children[pos+1:])
    n.children[pos+1] = child
}

// FindChildIndex returns the index of the child that should contain the key
func (n *GenericBranchNode[K]) FindChildIndex(key K, less func(a, b K) bool) int {
    // Find the position using binary search
    return sort.Search(len(n.keys), func(i int) bool {
        return !less(n.keys[i], key)
    })
}
```

## 4. Generic B+ Tree

```go
// GenericBPlusTree is a B+ tree that works with any comparable type
type GenericBPlusTree[K any] struct {
    root            GenericNode[K]
    branchingFactor int
    height          int
    size            int
    less            func(a, b K) bool
    equal           func(a, b K) bool
    bloomFilter     BloomFilterInterface
}

// NewGenericBPlusTree creates a new generic B+ tree
func NewGenericBPlusTree[K any](
    branchingFactor int,
    less func(a, b K) bool,
    equal func(a, b K) bool,
) *GenericBPlusTree[K] {
    if branchingFactor < 3 {
        branchingFactor = 3 // Minimum branching factor
    }

    // Create a Bloom filter with reasonable default parameters
    bloomSize, hashFunctions := OptimalBloomFilterSize(1000, 0.01)

    return &GenericBPlusTree[K]{
        root:            NewGenericLeafNode[K](),
        branchingFactor: branchingFactor,
        height:          1,
        size:            0,
        less:            less,
        equal:           equal,
        bloomFilter:     NewBloomFilter(bloomSize, hashFunctions),
    }
}

// Size returns the number of keys in the tree
func (t *GenericBPlusTree[K]) Size() int {
    return t.size
}

// Height returns the height of the tree
func (t *GenericBPlusTree[K]) Height() int {
    return t.height
}

// Insert inserts a key into the tree
func (t *GenericBPlusTree[K]) Insert(key K) bool {
    if t.root.IsFull(t.branchingFactor) {
        t.splitRoot()
    }

    inserted := t.insertNonFull(t.root, key)

    if inserted {
        t.size++
        t.updateBloomFilter(key)
    }

    return inserted
}

// Contains returns true if the tree contains the key
func (t *GenericBPlusTree[K]) Contains(key K) bool {
    // Check bloom filter first (if valid)
    if t.bloomFilter.IsValid() && !t.bloomFilterContains(key) {
        return false // Definitely not present
    }

    // Check the tree
    return t.findLeaf(t.root, key)
}

// Delete removes a key from the tree
func (t *GenericBPlusTree[K]) Delete(key K) bool {
    deleted, _ := t.deleteAndBalance(t.root, nil, -1, key)

    if deleted {
        t.size--
        // Handle root underflow and update bloom filter
    }

    return deleted
}

// RangeQuery returns all keys in the range [start, end]
func (t *GenericBPlusTree[K]) RangeQuery(start, end K) []K {
    result := make([]K, 0)

    // Find the leaf containing the start key
    leaf := t.findLeafNode(t.root, start)
    if leaf == nil {
        return result
    }

    // Traverse the linked list of leaves until we reach the end key
    for leaf != nil {
        for _, key := range leaf.Keys() {
            if (t.less(start, key) || t.equal(start, key)) &&
               (t.less(key, end) || t.equal(key, end)) {
                result = append(result, key)
            }

            if t.less(end, key) {
                return result // We've reached the end
            }
        }

        leaf = leaf.next
    }

    return result
}
```

## 5. Generic Set

```go
// GenericSet represents a set of values of type K
type GenericSet[K any] struct {
    tree *GenericBPlusTree[K]
}

// NewGenericSet creates a new set with the given branching factor
// and comparison functions
func NewGenericSet[K any](
    branchingFactor int,
    less func(a, b K) bool,
    equal func(a, b K) bool,
) *GenericSet[K] {
    return &GenericSet[K]{
        tree: NewGenericBPlusTree[K](branchingFactor, less, equal),
    }
}

// NewIntSet creates a new set for int values
func NewIntSet(branchingFactor int) *GenericSet[int] {
    return NewGenericSet[int](
        branchingFactor,
        func(a, b int) bool { return a < b },
        func(a, b int) bool { return a == b },
    )
}

// NewStringSet creates a new set for string values
func NewStringSet(branchingFactor int) *GenericSet[string] {
    return NewGenericSet[string](
        branchingFactor,
        func(a, b string) bool { return a < b },
        func(a, b string) bool { return a == b },
    )
}

// Add adds a value to the set
func (s *GenericSet[K]) Add(value K) bool {
    return s.tree.Insert(value)
}

// Contains returns true if the set contains the value
func (s *GenericSet[K]) Contains(value K) bool {
    return s.tree.Contains(value)
}

// Delete removes a value from the set
func (s *GenericSet[K]) Delete(value K) bool {
    return s.tree.Delete(value)
}

// GetAll returns all elements in the set
func (s *GenericSet[K]) GetAll() []K {
    // Implementation depends on the tree's traversal method
    return nil // Placeholder
}

// Range returns all elements in the range [start, end]
func (s *GenericSet[K]) Range(start, end K) []K {
    return s.tree.RangeQuery(start, end)
}
```

## 6. Bloom Filter Integration

```go
// bloomFilterContains checks if the bloom filter contains the key
func (t *GenericBPlusTree[K]) bloomFilterContains(key K) bool {
    // We need a way to hash the generic key to a uint64
    // This is a challenge with generic types

    // Option 1: Use reflection (not ideal for performance)
    hash := hashGenericKey(key)

    // Option 2: Require a hash function as part of the tree constructor
    // hash := t.hashFunc(key)

    return t.bloomFilter.Contains(hash)
}

// updateBloomFilter adds a key to the bloom filter
func (t *GenericBPlusTree[K]) updateBloomFilter(key K) {
    hash := hashGenericKey(key)
    t.bloomFilter.Add(hash)
}

// hashGenericKey generates a hash for a generic key
// This is a simple implementation and might not be suitable for all types
func hashGenericKey[K any](key K) uint64 {
    // This is a placeholder implementation
    // In practice, we would need a more robust hashing mechanism

    // Option 1: Use reflection (not ideal for performance)
    h := fnv.New64a()
    binary.Write(h, binary.LittleEndian, key)
    return h.Sum64()

    // Option 2: Type switch for common types
    // switch k := any(key).(type) {
    // case int:
    //     return uint64(k)
    // case string:
    //     h := fnv.New64a()
    //     h.Write([]byte(k))
    //     return h.Sum64()
    // default:
    //     // Fallback to reflection
    // }
}
```

## 7. Usage Examples

```go
func ExampleIntSet() {
    // Create a set of integers
    set := NewIntSet(4)

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
}

func ExampleStringSet() {
    // Create a set of strings
    set := NewStringSet(4)

    // Add some values
    set.Add("apple")
    set.Add("banana")
    set.Add("cherry")

    // Check if the set contains a value
    fmt.Println(set.Contains("banana")) // Output: true
    fmt.Println(set.Contains("date"))   // Output: false

    // Get a range of values
    fmt.Println(set.Range("apricot", "citrus")) // Output: [apple banana cherry]

    // Delete a value
    set.Delete("banana")
    fmt.Println(set.Contains("banana")) // Output: false
}
```

## 8. Challenges and Considerations

1. **Bloom Filter Integration**: The main challenge is hashing generic keys for the bloom filter. Options include:

   - Requiring a hash function as part of the tree constructor
   - Using reflection (with performance implications)
   - Using type switches for common types
   - Creating a separate GenericBloomFilter that works with the key type

2. **Performance**: Generic code may have some overhead compared to specialized code. Benchmarking will be important.

3. **Memory Usage**: Generic slices may use more memory than slices of primitive types.

4. **Type Constraints**: In a full implementation, we might want to add constraints to ensure the key type supports the operations we need.
