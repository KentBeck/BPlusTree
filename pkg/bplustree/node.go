package bplustree

import (
	"sort"
)

// NodeType represents the type of a node in the B+ tree
type NodeType int

const (
	// Branch is a node that contains keys and pointers to other nodes
	Branch NodeType = iota
	// Leaf is a node that contains keys and values
	Leaf
)

// Node interface represents a node in the B+ tree
type Node interface {
	// Type returns the type of the node
	Type() NodeType
	// Keys returns the keys in the node
	Keys() []uint64
	// KeyCount returns the number of keys in the node
	KeyCount() int
	// IsFull returns true if the node is full
	IsFull(branchingFactor int) bool
	// IsUnderflow returns true if the node has too few keys
	IsUnderflow(branchingFactor int) bool
	// InsertKey inserts a key into the node
	InsertKey(key uint64) bool
	// DeleteKey deletes a key from the node
	DeleteKey(key uint64) bool
	// FindKey returns the index of the key in the node, or -1 if not found
	FindKey(key uint64) int
	// Contains returns true if the node contains the key
	Contains(key uint64) bool
	// TryBorrowFromSibling attempts to borrow a key from a sibling
	// Returns true if borrowing was successful
	TryBorrowFromSibling(parent *BranchImpl, nodeIndex int, branchingFactor int) bool
}

// BranchImpl represents an internal node in the B+ tree
type BranchImpl struct {
	keys     []uint64
	children []Node
}

// NewBranch creates a new internal node
func NewBranch() *BranchImpl {
	return &BranchImpl{
		keys:     make([]uint64, 0),
		children: make([]Node, 0),
	}
}

// Type returns the type of the node
func (n *BranchImpl) Type() NodeType {
	return Branch
}

// Keys returns the keys in the node
func (n *BranchImpl) Keys() []uint64 {
	return n.keys
}

// Children returns the children of the node
func (n *BranchImpl) Children() []Node {
	return n.children
}

// IsFull returns true if the node is full
func (n *BranchImpl) IsFull(branchingFactor int) bool {
	return len(n.keys) >= branchingFactor-1
}

// IsUnderflow returns true if the node has too few keys
func (n *BranchImpl) IsUnderflow(branchingFactor int) bool {
	// For internal nodes, minimum number of keys is ceil(m/2)-1
	// For branching factor 3, minimum is 1 key
	return len(n.keys) < minInternalKeys(branchingFactor)
}

// KeyCount returns the number of keys in the node
func (n *BranchImpl) KeyCount() int {
	return len(n.keys)
}

// Contains returns true if the node contains the key
func (n *BranchImpl) Contains(key uint64) bool {
	return n.FindKey(key) != -1
}

// InsertKey inserts a key and child into the node at the correct position
func (n *BranchImpl) InsertKeyWithChild(key uint64, child Node) {
	pos := sort.Search(len(n.keys), func(i int) bool {
		return n.keys[i] >= key
	})

	// Insert key
	n.keys = append(n.keys, 0)
	copy(n.keys[pos+1:], n.keys[pos:])
	n.keys[pos] = key

	// Insert child (goes to the right of the key)
	n.children = append(n.children, nil)
	copy(n.children[pos+2:], n.children[pos+1:])
	n.children[pos+1] = child
}

// InsertKey inserts a key into the node
func (n *BranchImpl) InsertKey(key uint64) bool {
	// This is a placeholder to satisfy the Node interface
	// Internal nodes should use InsertKeyWithChild instead
	return false
}

// DeleteKey deletes a key from the node
func (n *BranchImpl) DeleteKey(key uint64) bool {
	pos := n.FindKey(key)
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
func (n *BranchImpl) FindKey(key uint64) int {
	for i, k := range n.keys {
		if k == key {
			return i
		}
	}
	return -1
}

// FindChildIndex returns the index of the child that should contain the key
func (n *BranchImpl) FindChildIndex(key uint64) int {
	pos := sort.Search(len(n.keys), func(i int) bool {
		return n.keys[i] > key
	})
	return pos
}

// SetChild sets the child at the given index
func (n *BranchImpl) SetChild(index int, child Node) {
	if index < len(n.children) {
		n.children[index] = child
	} else if index == len(n.children) {
		n.children = append(n.children, child)
	}
}

// RemoveChild removes the child at the given index
func (n *BranchImpl) RemoveChild(index int) {
	if index < len(n.children) {
		copy(n.children[index:], n.children[index+1:])
		n.children = n.children[:len(n.children)-1]
	}
}

// MergeWith merges this node with another internal node
// The keys and children from the other node are appended to this node
// The separator key from the parent is inserted between the two nodes
func (n *BranchImpl) MergeWith(separatorKey uint64, other *BranchImpl) {
	// Add the separator key
	n.keys = append(n.keys, separatorKey)

	// Add all keys from the other node
	n.keys = append(n.keys, other.keys...)

	// Add all children from the other node
	n.children = append(n.children, other.children...)
}

// BorrowFromRight borrows a key and child from the right sibling
// The separator key from the parent is moved down to this node
// The leftmost key from the right sibling becomes the new separator in the parent
func (n *BranchImpl) BorrowFromRight(separatorKey uint64, rightSibling *BranchImpl, parentIndex int, parent *BranchImpl) {
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
// The separator key from the parent is moved down to this node
// The rightmost key from the left sibling becomes the new separator in the parent
func (n *BranchImpl) BorrowFromLeft(separatorKey uint64, leftSibling *BranchImpl, nodeIndex int, parent *BranchImpl) {
	// Insert the separator key at the beginning of this node's keys
	n.keys = append([]uint64{separatorKey}, n.keys...)

	// Insert the last child from the left sibling at the beginning of this node's children
	lastChildIndex := len(leftSibling.children) - 1
	n.children = append([]Node{leftSibling.children[lastChildIndex]}, n.children...)

	// Update the separator key in the parent
	parent.keys[nodeIndex-1] = leftSibling.keys[len(leftSibling.keys)-1]

	// Remove the borrowed key and child from the left sibling
	leftSibling.keys = leftSibling.keys[:len(leftSibling.keys)-1]
	leftSibling.children = leftSibling.children[:len(leftSibling.children)-1]
}

// TryBorrowFromSibling attempts to borrow a key from a sibling
// Returns true if borrowing was successful
func (n *BranchImpl) TryBorrowFromSibling(parent *BranchImpl, nodeIndex int, branchingFactor int) bool {
	// Try to borrow from right sibling first
	if n.tryBorrowFromRight(parent, nodeIndex, branchingFactor) {
		return true
	}

	// If that fails, try to borrow from left sibling
	return n.tryBorrowFromLeft(parent, nodeIndex, branchingFactor)
}

// tryBorrowFromRight attempts to borrow a key from the right sibling
func (n *BranchImpl) tryBorrowFromRight(parent *BranchImpl, nodeIndex int, branchingFactor int) bool {
	// Check if right sibling exists and has enough keys
	if nodeIndex >= len(parent.Children())-1 || nodeIndex >= len(parent.Keys()) {
		return false
	}

	rightSibling, ok := parent.Children()[nodeIndex+1].(*BranchImpl)
	if !ok {
		return false
	}

	minKeys := minInternalKeys(branchingFactor)
	if len(rightSibling.Keys()) <= minKeys {
		return false
	}

	// Borrow from right sibling
	separatorKey := parent.Keys()[nodeIndex]
	n.BorrowFromRight(separatorKey, rightSibling, nodeIndex, parent)
	return true
}

// tryBorrowFromLeft attempts to borrow a key from the left sibling
func (n *BranchImpl) tryBorrowFromLeft(parent *BranchImpl, nodeIndex int, branchingFactor int) bool {
	// Check if left sibling exists and has enough keys
	if nodeIndex <= 0 || nodeIndex-1 >= len(parent.Keys()) {
		return false
	}

	leftSibling, ok := parent.Children()[nodeIndex-1].(*BranchImpl)
	if !ok {
		return false
	}

	minKeys := minInternalKeys(branchingFactor)
	if len(leftSibling.Keys()) <= minKeys {
		return false
	}

	// Borrow from left sibling
	separatorKey := parent.Keys()[nodeIndex-1]
	n.BorrowFromLeft(separatorKey, leftSibling, nodeIndex, parent)
	return true
}

// LeafImpl represents a leaf node in the B+ tree
type LeafImpl struct {
	keys []uint64
	next *LeafImpl // Pointer to the next leaf node for range queries
}

// NewLeaf creates a new leaf node
func NewLeaf() *LeafImpl {
	return &LeafImpl{
		keys: make([]uint64, 0),
		next: nil,
	}
}

// Type returns the type of the node
func (n *LeafImpl) Type() NodeType {
	return Leaf
}

// Keys returns the keys in the node
func (n *LeafImpl) Keys() []uint64 {
	return n.keys
}

// Next returns the next leaf node
func (n *LeafImpl) Next() *LeafImpl {
	return n.next
}

// SetNext sets the next leaf node
func (n *LeafImpl) SetNext(next *LeafImpl) {
	n.next = next
}

// IsFull returns true if the node is full
func (n *LeafImpl) IsFull(branchingFactor int) bool {
	return len(n.keys) >= branchingFactor
}

// IsUnderflow returns true if the node has too few keys
func (n *LeafImpl) IsUnderflow(branchingFactor int) bool {
	// For leaf nodes, minimum number of keys is ceil(m/2)
	// For branching factor 3, minimum is 2 keys
	return len(n.keys) < minLeafKeys(branchingFactor)
}

// KeyCount returns the number of keys in the node
func (n *LeafImpl) KeyCount() int {
	return len(n.keys)
}

// InsertKey inserts a key into the node
func (n *LeafImpl) InsertKey(key uint64) bool {
	// Find position to insert
	pos := sort.Search(len(n.keys), func(i int) bool {
		return n.keys[i] >= key
	})

	// Check if key already exists
	if pos < len(n.keys) && n.keys[pos] == key {
		return false // Key already exists
	}

	// Insert key
	n.keys = append(n.keys, 0)
	copy(n.keys[pos+1:], n.keys[pos:])
	n.keys[pos] = key
	return true
}

// DeleteKey deletes a key from the node
func (n *LeafImpl) DeleteKey(key uint64) bool {
	pos := n.FindKey(key)
	if pos == -1 {
		return false
	}

	// Remove key
	copy(n.keys[pos:], n.keys[pos+1:])
	n.keys = n.keys[:len(n.keys)-1]
	return true
}

// FindKey returns the index of the key in the node, or -1 if not found
func (n *LeafImpl) FindKey(key uint64) int {
	for i, k := range n.keys {
		if k == key {
			return i
		}
	}
	return -1
}

// Contains returns true if the node contains the key
func (n *LeafImpl) Contains(key uint64) bool {
	return n.FindKey(key) != -1
}

// MergeWith merges this node with another leaf node
// The keys from the other node are appended to this node
func (n *LeafImpl) MergeWith(other *LeafImpl) {
	// Add all keys from the other node
	n.keys = append(n.keys, other.keys...)

	// Update the next pointer
	n.next = other.next
}

// BorrowFromRight borrows a key from the right sibling
// The borrowed key is removed from the right sibling
// The parent's separator key is updated to the new minimum key in the right sibling
func (n *LeafImpl) BorrowFromRight(rightSibling *LeafImpl, parentIndex int, parent *BranchImpl) {
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
// The borrowed key is removed from the left sibling
func (n *LeafImpl) BorrowFromLeft(leftSibling *LeafImpl) {
	// Borrow the last key from the left sibling
	lastKeyIndex := len(leftSibling.keys) - 1
	borrowedKey := leftSibling.keys[lastKeyIndex]

	// Insert the borrowed key at the beginning of this node's keys
	n.keys = append([]uint64{borrowedKey}, n.keys...)

	// Remove the borrowed key from the left sibling
	leftSibling.keys = leftSibling.keys[:lastKeyIndex]
}

// TryBorrowFromSibling attempts to borrow a key from a sibling
// Returns true if borrowing was successful
func (n *LeafImpl) TryBorrowFromSibling(parent *BranchImpl, nodeIndex int, branchingFactor int) bool {
	// Try to borrow from right sibling first
	if n.tryBorrowFromRight(parent, nodeIndex, branchingFactor) {
		return true
	}

	// If that fails, try to borrow from left sibling
	return n.tryBorrowFromLeft(parent, nodeIndex, branchingFactor)
}

// tryBorrowFromRight attempts to borrow a key from the right sibling
func (n *LeafImpl) tryBorrowFromRight(parent *BranchImpl, nodeIndex int, branchingFactor int) bool {
	// Check if right sibling exists and has enough keys
	if nodeIndex >= len(parent.Children())-1 || nodeIndex >= len(parent.Keys()) {
		return false
	}

	rightSibling, ok := parent.Children()[nodeIndex+1].(*LeafImpl)
	if !ok {
		return false
	}

	minKeys := minLeafKeys(branchingFactor)
	if len(rightSibling.Keys()) <= minKeys {
		return false
	}

	// Borrow from right sibling
	n.BorrowFromRight(rightSibling, nodeIndex, parent)
	return true
}

// tryBorrowFromLeft attempts to borrow a key from the left sibling
func (n *LeafImpl) tryBorrowFromLeft(parent *BranchImpl, nodeIndex int, branchingFactor int) bool {
	// Check if left sibling exists and has enough keys
	if nodeIndex <= 0 || nodeIndex-1 >= len(parent.Keys()) {
		return false
	}

	leftSibling, ok := parent.Children()[nodeIndex-1].(*LeafImpl)
	if !ok {
		return false
	}

	minKeys := minLeafKeys(branchingFactor)
	if len(leftSibling.Keys()) <= minKeys {
		return false
	}

	// Borrow from left sibling
	n.BorrowFromLeft(leftSibling)
	// Update the separator key in the parent
	if len(n.Keys()) > 0 {
		parent.keys[nodeIndex-1] = n.Keys()[0]
	}
	return true
}
