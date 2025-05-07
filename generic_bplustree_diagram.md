# Generic B+ Tree Component Diagram

```mermaid
classDiagram
    class GenericNode~K~ {
        <<interface>>
        +Type() NodeType
        +Keys() []K
        +KeyCount() int
        +IsFull(branchingFactor int) bool
        +IsUnderflow(branchingFactor int) bool
        +InsertKey(key K, less func(a, b K) bool) bool
        +DeleteKey(key K, equal func(a, b K) bool) bool
        +FindKey(key K, equal func(a, b K) bool) int
        +Contains(key K, equal func(a, b K) bool) bool
    }

    class GenericLeafNode~K~ {
        -keys []K
        -next *GenericLeafNode~K~
        +Type() NodeType
        +Keys() []K
        +KeyCount() int
        +IsFull(branchingFactor int) bool
        +IsUnderflow(branchingFactor int) bool
        +InsertKey(key K, less func(a, b K) bool) bool
        +DeleteKey(key K, equal func(a, b K) bool) bool
        +FindKey(key K, equal func(a, b K) bool) int
        +Contains(key K, equal func(a, b K) bool) bool
        +MergeWith(other *GenericLeafNode~K~)
        +BorrowFromRight(rightSibling *GenericLeafNode~K~, parentIndex int, parent *GenericBranchNode~K~)
        +BorrowFromLeft(leftSibling *GenericLeafNode~K~, parentIndex int, parent *GenericBranchNode~K~)
    }

    class GenericBranchNode~K~ {
        -keys []K
        -children []GenericNode~K~
        +Type() NodeType
        +Keys() []K
        +KeyCount() int
        +IsFull(branchingFactor int) bool
        +IsUnderflow(branchingFactor int) bool
        +InsertKey(key K, less func(a, b K) bool) bool
        +DeleteKey(key K, equal func(a, b K) bool) bool
        +FindKey(key K, equal func(a, b K) bool) int
        +Contains(key K, equal func(a, b K) bool) bool
        +InsertKeyWithChild(key K, child GenericNode~K~, less func(a, b K) bool)
        +FindChildIndex(key K, less func(a, b K) bool) int
        +MergeWith(separatorKey K, other *GenericBranchNode~K~)
        +BorrowFromRight(separatorKey K, rightSibling *GenericBranchNode~K~, parentIndex int, parent *GenericBranchNode~K~)
        +BorrowFromLeft(separatorKey K, leftSibling *GenericBranchNode~K~, parentIndex int, parent *GenericBranchNode~K~)
    }

    class GenericBPlusTree~K~ {
        -root GenericNode~K~
        -branchingFactor int
        -height int
        -size int
        -less func(a, b K) bool
        -equal func(a, b K) bool
        -bloomFilter BloomFilterInterface
        +Insert(key K) bool
        +Contains(key K) bool
        +Delete(key K) bool
        +GetAllKeys() []K
        +RangeQuery(start K, end K) []K
        -splitRoot()
        -insertNonFull(node GenericNode~K~, key K) bool
        -splitChild(parent *GenericBranchNode~K~, childIndex int)
        -findLeaf(node GenericNode~K~, key K) bool
        -deleteAndBalance(node GenericNode~K~, parent *GenericBranchNode~K~, parentChildIndex int, key K) (bool, K)
    }

    class GenericSet~K~ {
        -tree *GenericBPlusTree~K~
        +Add(value K) bool
        +Contains(value K) bool
        +Delete(value K) bool
        +GetAll() []K
        +Range(start K, end K) []K
        +Size() int
        +IsEmpty() bool
        +Clear()
    }

    class BloomFilterInterface {
        <<interface>>
        +Add(key uint64)
        +Contains(key uint64) bool
        +Clear()
        +IsValid() bool
        +SetValid()
    }

    class GenericBloomFilter~K~ {
        -filter BloomFilterInterface
        -hashFunc func(K) uint64
        +Add(key K)
        +Contains(key K) bool
        +Clear()
        +IsValid() bool
        +SetValid()
    }

    GenericNode~K~ <|.. GenericLeafNode~K~ : implements
    GenericNode~K~ <|.. GenericBranchNode~K~ : implements
    GenericBPlusTree~K~ o-- GenericNode~K~ : has root
    GenericBPlusTree~K~ o-- BloomFilterInterface : has filter
    GenericSet~K~ o-- GenericBPlusTree~K~ : has tree
    GenericBloomFilter~K~ ..|> BloomFilterInterface : implements
    GenericBPlusTree~K~ o-- GenericBloomFilter~K~ : may use
```

## Key Relationships

1. **GenericNode Interface**: The core interface that both leaf and branch nodes implement.

2. **Node Implementations**:

   - `GenericLeafNode`: Stores the actual keys and implements the GenericNode interface.
   - `GenericBranchNode`: Stores keys and pointers to child nodes and implements the GenericNode interface.

3. **Tree Structure**:

   - `GenericBPlusTree`: Contains the root node and operations for manipulating the tree.
   - The root can be either a leaf or branch node, depending on the tree's height.

4. **Bloom Filter Integration**:

   - `GenericBloomFilter`: A generic wrapper around the BloomFilterInterface that can work with any key type.
   - The B+ tree uses the bloom filter to optimize Contains operations.

5. **Set Abstraction**:
   - `GenericSet`: A high-level abstraction that uses the B+ tree to implement a set data structure.
   - Provides a simpler API for common set operations.

## Data Flow

1. **Insertion**:

   ```
   GenericSet.Add(key) → GenericBPlusTree.Insert(key) →
   [find leaf node] → GenericLeafNode.InsertKey(key) →
   [if node is full] → GenericBPlusTree.splitChild() →
   [update bloom filter]
   ```

2. **Lookup**:

   ```
   GenericSet.Contains(key) → GenericBPlusTree.Contains(key) →
   [check bloom filter] → [if might contain] →
   [find leaf node] → GenericLeafNode.Contains(key)
   ```

3. **Deletion**:

   ```
   GenericSet.Delete(key) → GenericBPlusTree.Delete(key) →
   [find leaf node] → GenericLeafNode.DeleteKey(key) →
   [if underflow] → [rebalance tree] →
   [update bloom filter]
   ```

4. **Range Query** (new functionality):
   ```
   GenericSet.Range(start, end) → GenericBPlusTree.RangeQuery(start, end) →
   [find start leaf] → [traverse linked list of leaves until end] →
   [collect keys in range]
   ```
