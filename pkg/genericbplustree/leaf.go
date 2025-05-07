package genericbplustree

import (
	"sort"
)

// LeafNode is a leaf node in the B+ tree
type LeafNode[K any] struct {
	keys []K
	next *LeafNode[K] // Pointer to the next leaf node for range queries
}

// NewLeafNode creates a new leaf node
func NewLeafNode[K any]() *LeafNode[K] {
	return &LeafNode[K]{
		keys: make([]K, 0),
		next: nil,
	}
}

// Type returns the type of the node
func (n *LeafNode[K]) Type() NodeType {
	return Leaf
}

// Keys returns the keys in the node
func (n *LeafNode[K]) Keys() []K {
	return n.keys
}

// Next returns the next leaf node
func (n *LeafNode[K]) Next() *LeafNode[K] {
	return n.next
}

// SetNext sets the next leaf node
func (n *LeafNode[K]) SetNext(next *LeafNode[K]) {
	n.next = next
}

// KeyCount returns the number of keys in the node
func (n *LeafNode[K]) KeyCount() int {
	return len(n.keys)
}

// IsFull returns true if the node is full
func (n *LeafNode[K]) IsFull(branchingFactor int) bool {
	return len(n.keys) >= branchingFactor
}

// IsUnderflow returns true if the node has too few keys
func (n *LeafNode[K]) IsUnderflow(branchingFactor int) bool {
	// For leaf nodes, minimum number of keys is ceil(m/2)
	return len(n.keys) < minLeafKeys(branchingFactor)
}

// InsertKey inserts a key into the node
func (n *LeafNode[K]) InsertKey(key K, less func(a, b K) bool) bool {
	// Find position to insert
	pos := n.findInsertPosition(key, less)

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

// findInsertPosition finds the position to insert a key
func (n *LeafNode[K]) findInsertPosition(key K, less func(a, b K) bool) int {
	// Find the position to insert using binary search
	return sort.Search(len(n.keys), func(i int) bool {
		return !less(n.keys[i], key) // equivalent to n.keys[i] >= key
	})
}

// DeleteKey deletes a key from the node
func (n *LeafNode[K]) DeleteKey(key K, equal func(a, b K) bool) bool {
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
func (n *LeafNode[K]) FindKey(key K, equal func(a, b K) bool) int {
	for i, k := range n.keys {
		if equal(k, key) {
			return i
		}
	}
	return -1
}

// Contains returns true if the node contains the key
func (n *LeafNode[K]) Contains(key K, equal func(a, b K) bool) bool {
	return n.FindKey(key, equal) != -1
}

// MergeWith merges this node with another leaf node
func (n *LeafNode[K]) MergeWith(other *LeafNode[K]) {
	// Add all keys from the other node
	n.keys = append(n.keys, other.keys...)

	// Update the next pointer
	n.next = other.next
}

// BorrowFromRight borrows a key from the right sibling
func (n *LeafNode[K]) BorrowFromRight(rightSibling *LeafNode[K], leafIndex int, parent *BranchNode[K]) {
	// Borrow the first key from the right sibling
	borrowedKey := rightSibling.keys[0]

	// Add the borrowed key to this node
	n.keys = append(n.keys, borrowedKey)

	// Remove the borrowed key from the right sibling
	rightSibling.keys = rightSibling.keys[1:]

	// Update the separator key in the parent
	if len(rightSibling.keys) > 0 {
		parent.keys[leafIndex] = rightSibling.keys[0]
	}
}

// BorrowFromLeft borrows a key from the left sibling
func (n *LeafNode[K]) BorrowFromLeft(leftSibling *LeafNode[K], leafIndex int, parent *BranchNode[K]) {
	// Borrow the last key from the left sibling
	lastKeyIndex := len(leftSibling.keys) - 1
	borrowedKey := leftSibling.keys[lastKeyIndex]

	// Insert the borrowed key at the beginning of this node's keys
	n.keys = append([]K{borrowedKey}, n.keys...)

	// Remove the borrowed key from the left sibling
	leftSibling.keys = leftSibling.keys[:lastKeyIndex]

	// Update the separator key in the parent
	parent.keys[leafIndex-1] = n.keys[0]
}

// Helper functions for minimum key calculations
func minLeafKeys(branchingFactor int) int {
	return (branchingFactor + 1) / 2
}

func minInternalKeys(branchingFactor int) int {
	return (branchingFactor+1)/2 - 1
}
