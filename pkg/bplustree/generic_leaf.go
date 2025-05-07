package bplustree

import (
	"sort"
)

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

// Next returns the next leaf node
func (n *GenericLeafNode[K]) Next() *GenericLeafNode[K] {
	return n.next
}

// SetNext sets the next leaf node
func (n *GenericLeafNode[K]) SetNext(next *GenericLeafNode[K]) {
	n.next = next
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
	return len(n.keys) < minLeafKeys(branchingFactor)
}

// InsertKey inserts a key into the node
func (n *GenericLeafNode[K]) InsertKey(key K, less func(a, b K) bool) bool {
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
func (n *GenericLeafNode[K]) findInsertPosition(key K, less func(a, b K) bool) int {
	// Find the position to insert using binary search
	return sort.Search(len(n.keys), func(i int) bool {
		return !less(n.keys[i], key) // equivalent to n.keys[i] >= key
	})
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

// MergeWith merges this node with another leaf node
func (n *GenericLeafNode[K]) MergeWith(other *GenericLeafNode[K]) {
	// Add all keys from the other node
	n.keys = append(n.keys, other.keys...)

	// Update the next pointer
	n.next = other.next
}

// BorrowFromRight borrows a key from the right sibling
func (n *GenericLeafNode[K]) BorrowFromRight(rightSibling *GenericLeafNode[K], parentIndex int, parent *GenericBranchNode[K]) {
	// Borrow the first key from the right sibling
	borrowedKey := rightSibling.keys[0]

	// Add the borrowed key to this node
	n.keys = append(n.keys, borrowedKey)

	// Remove the borrowed key from the right sibling
	rightSibling.keys = rightSibling.keys[1:]

	// Update the separator key in the parent
	if len(rightSibling.keys) > 0 {
		parent.keys[parentIndex] = rightSibling.keys[0]
	}
}

// BorrowFromLeft borrows a key from the left sibling
func (n *GenericLeafNode[K]) BorrowFromLeft(leftSibling *GenericLeafNode[K], parentIndex int, parent *GenericBranchNode[K]) {
	// Borrow the last key from the left sibling
	lastKeyIndex := len(leftSibling.keys) - 1
	borrowedKey := leftSibling.keys[lastKeyIndex]

	// Insert the borrowed key at the beginning of this node's keys
	n.keys = append([]K{borrowedKey}, n.keys...)

	// Remove the borrowed key from the left sibling
	leftSibling.keys = leftSibling.keys[:lastKeyIndex]

	// Update the separator key in the parent
	parent.keys[parentIndex-1] = n.keys[0]
}
