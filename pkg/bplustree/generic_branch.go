package bplustree

import (
	"sort"
)

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
	return len(n.keys) < minInternalKeys(branchingFactor)
}

// InsertKeyWithChild inserts a key and child into the node at the correct position
func (n *GenericBranchNode[K]) InsertKeyWithChild(key K, child GenericNode[K], less func(a, b K) bool) {
	pos := n.findInsertPosition(key, less)

	// Insert key
	n.keys = append(n.keys, *new(K)) // Add zero value of K
	copy(n.keys[pos+1:], n.keys[pos:])
	n.keys[pos] = key

	// Insert child (goes to the right of the key)
	n.children = append(n.children, nil)
	copy(n.children[pos+2:], n.children[pos+1:])
	n.children[pos+1] = child
}

// findInsertPosition finds the position to insert a key
func (n *GenericBranchNode[K]) findInsertPosition(key K, less func(a, b K) bool) int {
	// Find the position to insert using binary search
	return sort.Search(len(n.keys), func(i int) bool {
		return !less(n.keys[i], key) // equivalent to n.keys[i] >= key
	})
}

// InsertKey inserts a key into the node
func (n *GenericBranchNode[K]) InsertKey(key K, less func(a, b K) bool) bool {
	// This is a placeholder to satisfy the Node interface
	// Branch nodes should use InsertKeyWithChild instead
	return false
}

// DeleteKey deletes a key from the node
func (n *GenericBranchNode[K]) DeleteKey(key K, equal func(a, b K) bool) bool {
	pos := n.FindKey(key, equal)
	if pos == -1 {
		return false
	}

	// Remove key
	copy(n.keys[pos:], n.keys[pos+1:])
	n.keys = n.keys[:len(n.keys)-1]

	// Remove child to the right of the key
	copy(n.children[pos+1:], n.children[pos+2:])
	n.children = n.children[:len(n.children)-1]

	return true
}

// FindKey returns the index of the key in the node, or -1 if not found
func (n *GenericBranchNode[K]) FindKey(key K, equal func(a, b K) bool) int {
	for i, k := range n.keys {
		if equal(k, key) {
			return i
		}
	}
	return -1
}

// Contains returns true if the node contains the key
func (n *GenericBranchNode[K]) Contains(key K, equal func(a, b K) bool) bool {
	return n.FindKey(key, equal) != -1
}

// FindChildIndex returns the index of the child that should contain the key
func (n *GenericBranchNode[K]) FindChildIndex(key K, less func(a, b K) bool) int {
	// Find the position using binary search
	pos := sort.Search(len(n.keys), func(i int) bool {
		return !less(n.keys[i], key) // equivalent to n.keys[i] >= key
	})

	// If all keys are less than the search key, return the last child
	if pos == len(n.keys) {
		return pos
	}

	// If the key at pos is equal to the search key, return the child to the right
	// Otherwise, return the child at pos
	return pos
}

// SetChild sets the child at the given index
func (n *GenericBranchNode[K]) SetChild(index int, child GenericNode[K]) {
	if index < len(n.children) {
		n.children[index] = child
	} else if index == len(n.children) {
		n.children = append(n.children, child)
	}
}

// RemoveChild removes the child at the given index
func (n *GenericBranchNode[K]) RemoveChild(index int) {
	if index < len(n.children) {
		copy(n.children[index:], n.children[index+1:])
		n.children = n.children[:len(n.children)-1]
	}
}

// MergeWith merges this node with another branch node
func (n *GenericBranchNode[K]) MergeWith(separatorKey K, other *GenericBranchNode[K]) {
	// Add the separator key
	n.keys = append(n.keys, separatorKey)

	// Add all keys from the other node
	n.keys = append(n.keys, other.keys...)

	// Add all children from the other node
	n.children = append(n.children, other.children...)
}

// BorrowFromRight borrows a key and child from the right sibling
func (n *GenericBranchNode[K]) BorrowFromRight(separatorKey K, rightSibling *GenericBranchNode[K], parentIndex int, parent *GenericBranchNode[K]) {
	// Add the separator key from parent to this node
	n.keys = append(n.keys, separatorKey)

	// Add the first child from the right sibling to this node
	n.children = append(n.children, rightSibling.children[0])

	// Update the separator key in the parent
	parent.keys[parentIndex] = rightSibling.keys[0]

	// Remove the borrowed key and child from the right sibling
	rightSibling.keys = rightSibling.keys[1:]
	rightSibling.children = rightSibling.children[1:]
}

// BorrowFromLeft borrows a key and child from the left sibling
func (n *GenericBranchNode[K]) BorrowFromLeft(separatorKey K, leftSibling *GenericBranchNode[K], parentIndex int, parent *GenericBranchNode[K]) {
	// Insert the separator key at the beginning of this node's keys
	n.keys = append([]K{separatorKey}, n.keys...)

	// Insert the last child from the left sibling at the beginning of this node's children
	lastChildIndex := len(leftSibling.children) - 1
	n.children = append([]GenericNode[K]{leftSibling.children[lastChildIndex]}, n.children...)

	// Update the separator key in the parent
	parent.keys[parentIndex-1] = leftSibling.keys[len(leftSibling.keys)-1]

	// Remove the borrowed key and child from the left sibling
	leftSibling.keys = leftSibling.keys[:len(leftSibling.keys)-1]
	leftSibling.children = leftSibling.children[:len(leftSibling.children)-1]
}
