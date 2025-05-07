package genericbplustree

// NodeType represents the type of node (leaf or branch)
type NodeType int

const (
	Leaf NodeType = iota
	Branch
)

// Node is a generic interface for nodes in the B+ tree
type Node[K any] interface {
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
