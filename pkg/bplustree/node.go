package bplustree

import (
	"sort"
)

// NodeType represents the type of a node in the B+ tree
type NodeType int

const (
	// Branch is a node that contains keys and pointers to other nodes
	Branch NodeType = iota
	// LeafNode is a node that contains keys and values
	LeafNode
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
}

// InternalNodeImpl represents an internal node in the B+ tree
type InternalNodeImpl struct {
	keys     []uint64
	children []Node
}

// NewInternalNode creates a new internal node
func NewInternalNode() *InternalNodeImpl {
	return &InternalNodeImpl{
		keys:     make([]uint64, 0),
		children: make([]Node, 0),
	}
}

// Type returns the type of the node
func (n *InternalNodeImpl) Type() NodeType {
	return Branch
}

// Keys returns the keys in the node
func (n *InternalNodeImpl) Keys() []uint64 {
	return n.keys
}

// Children returns the children of the node
func (n *InternalNodeImpl) Children() []Node {
	return n.children
}

// IsFull returns true if the node is full
func (n *InternalNodeImpl) IsFull(branchingFactor int) bool {
	return len(n.keys) >= branchingFactor-1
}

// IsUnderflow returns true if the node has too few keys
func (n *InternalNodeImpl) IsUnderflow(branchingFactor int) bool {
	// For internal nodes, minimum number of keys is ceil(m/2)-1
	// For branching factor 3, minimum is 1 key
	return len(n.keys) < minInternalKeys(branchingFactor)
}

// KeyCount returns the number of keys in the node
func (n *InternalNodeImpl) KeyCount() int {
	return len(n.keys)
}

// Contains returns true if the node contains the key
func (n *InternalNodeImpl) Contains(key uint64) bool {
	return n.FindKey(key) != -1
}

// InsertKey inserts a key and child into the node at the correct position
func (n *InternalNodeImpl) InsertKeyWithChild(key uint64, child Node) {
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
func (n *InternalNodeImpl) InsertKey(key uint64) bool {
	// This is a placeholder to satisfy the Node interface
	// Internal nodes should use InsertKeyWithChild instead
	return false
}

// DeleteKey deletes a key from the node
func (n *InternalNodeImpl) DeleteKey(key uint64) bool {
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
func (n *InternalNodeImpl) FindKey(key uint64) int {
	for i, k := range n.keys {
		if k == key {
			return i
		}
	}
	return -1
}

// FindChildIndex returns the index of the child that should contain the key
func (n *InternalNodeImpl) FindChildIndex(key uint64) int {
	pos := sort.Search(len(n.keys), func(i int) bool {
		return n.keys[i] > key
	})
	return pos
}

// SetChild sets the child at the given index
func (n *InternalNodeImpl) SetChild(index int, child Node) {
	if index < len(n.children) {
		n.children[index] = child
	} else if index == len(n.children) {
		n.children = append(n.children, child)
	}
}

// RemoveChild removes the child at the given index
func (n *InternalNodeImpl) RemoveChild(index int) {
	if index < len(n.children) {
		copy(n.children[index:], n.children[index+1:])
		n.children = n.children[:len(n.children)-1]
	}
}

// MergeWith merges this node with another internal node
// The keys and children from the other node are appended to this node
// The separator key from the parent is inserted between the two nodes
func (n *InternalNodeImpl) MergeWith(separatorKey uint64, other *InternalNodeImpl) {
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
func (n *InternalNodeImpl) BorrowFromRight(separatorKey uint64, rightSibling *InternalNodeImpl, parentIndex int, parent *InternalNodeImpl) {
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
func (n *InternalNodeImpl) BorrowFromLeft(separatorKey uint64, leftSibling *InternalNodeImpl, parentIndex int, parent *InternalNodeImpl) {
	// Insert the separator key at the beginning of this node's keys
	n.keys = append([]uint64{separatorKey}, n.keys...)

	// Insert the last child from the left sibling at the beginning of this node's children
	lastChildIndex := len(leftSibling.children) - 1
	n.children = append([]Node{leftSibling.children[lastChildIndex]}, n.children...)

	// Update the separator key in the parent
	parent.keys[parentIndex-1] = leftSibling.keys[len(leftSibling.keys)-1]

	// Remove the borrowed key and child from the left sibling
	leftSibling.keys = leftSibling.keys[:len(leftSibling.keys)-1]
	leftSibling.children = leftSibling.children[:len(leftSibling.children)-1]
}

// LeafNodeImpl represents a leaf node in the B+ tree
type LeafNodeImpl struct {
	keys []uint64
	next *LeafNodeImpl // Pointer to the next leaf node for range queries
}

// NewLeafNode creates a new leaf node
func NewLeafNode() *LeafNodeImpl {
	return &LeafNodeImpl{
		keys: make([]uint64, 0),
		next: nil,
	}
}

// Type returns the type of the node
func (n *LeafNodeImpl) Type() NodeType {
	return LeafNode
}

// Keys returns the keys in the node
func (n *LeafNodeImpl) Keys() []uint64 {
	return n.keys
}

// Next returns the next leaf node
func (n *LeafNodeImpl) Next() *LeafNodeImpl {
	return n.next
}

// SetNext sets the next leaf node
func (n *LeafNodeImpl) SetNext(next *LeafNodeImpl) {
	n.next = next
}

// IsFull returns true if the node is full
func (n *LeafNodeImpl) IsFull(branchingFactor int) bool {
	return len(n.keys) >= branchingFactor
}

// IsUnderflow returns true if the node has too few keys
func (n *LeafNodeImpl) IsUnderflow(branchingFactor int) bool {
	// For leaf nodes, minimum number of keys is ceil(m/2)
	// For branching factor 3, minimum is 2 keys
	return len(n.keys) < minLeafKeys(branchingFactor)
}

// KeyCount returns the number of keys in the node
func (n *LeafNodeImpl) KeyCount() int {
	return len(n.keys)
}

// InsertKey inserts a key into the node
func (n *LeafNodeImpl) InsertKey(key uint64) bool {
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
func (n *LeafNodeImpl) DeleteKey(key uint64) bool {
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
func (n *LeafNodeImpl) FindKey(key uint64) int {
	for i, k := range n.keys {
		if k == key {
			return i
		}
	}
	return -1
}

// Contains returns true if the node contains the key
func (n *LeafNodeImpl) Contains(key uint64) bool {
	return n.FindKey(key) != -1
}

// MergeWith merges this node with another leaf node
// The keys from the other node are appended to this node
func (n *LeafNodeImpl) MergeWith(other *LeafNodeImpl) {
	// Add all keys from the other node
	n.keys = append(n.keys, other.keys...)

	// Update the next pointer
	n.next = other.next
}

// BorrowFromRight borrows a key from the right sibling
// The borrowed key is removed from the right sibling
// The parent's separator key is updated to the new minimum key in the right sibling
func (n *LeafNodeImpl) BorrowFromRight(rightSibling *LeafNodeImpl, parentIndex int, parent *InternalNodeImpl) {
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
func (n *LeafNodeImpl) BorrowFromLeft(leftSibling *LeafNodeImpl) {
	// Borrow the last key from the left sibling
	lastKeyIndex := len(leftSibling.keys) - 1
	borrowedKey := leftSibling.keys[lastKeyIndex]

	// Insert the borrowed key at the beginning of this node's keys
	n.keys = append([]uint64{borrowedKey}, n.keys...)

	// Remove the borrowed key from the left sibling
	leftSibling.keys = leftSibling.keys[:lastKeyIndex]
}
