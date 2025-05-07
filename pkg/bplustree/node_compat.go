package bplustree

// This file provides backward compatibility with the old Node interface

// NodeType represents the type of node (leaf or branch)
type NodeType int

const (
	Leaf NodeType = iota
	Branch
)

// Node is an interface for nodes in the B+ tree
// This is kept for backward compatibility
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

// LeafImpl is a leaf node in the B+ tree
// This is kept for backward compatibility
type LeafImpl struct {
	*GenericLeafNode[uint64]
}

// BranchImpl is an internal node in the B+ tree
// This is kept for backward compatibility
type BranchImpl struct {
	*GenericBranchNode[uint64]
}

// NewLeaf creates a new leaf node
// This is kept for backward compatibility
func NewLeaf() *LeafImpl {
	return &LeafImpl{
		GenericLeafNode: NewGenericLeafNode[uint64](),
	}
}

// NewBranch creates a new branch node
// This is kept for backward compatibility
func NewBranch() *BranchImpl {
	return &BranchImpl{
		GenericBranchNode: NewGenericBranchNode[uint64](),
	}
}

// Implement the Node interface for LeafImpl

func (n *LeafImpl) Type() NodeType {
	return Leaf
}

func (n *LeafImpl) Keys() []uint64 {
	return n.GenericLeafNode.Keys()
}

func (n *LeafImpl) KeyCount() int {
	return n.GenericLeafNode.KeyCount()
}

func (n *LeafImpl) IsFull(branchingFactor int) bool {
	return n.GenericLeafNode.IsFull(branchingFactor)
}

func (n *LeafImpl) IsUnderflow(branchingFactor int) bool {
	return n.GenericLeafNode.IsUnderflow(branchingFactor)
}

func (n *LeafImpl) InsertKey(key uint64) bool {
	return n.GenericLeafNode.InsertKey(key, func(a, b uint64) bool { return a < b })
}

func (n *LeafImpl) DeleteKey(key uint64) bool {
	return n.GenericLeafNode.DeleteKey(key, func(a, b uint64) bool { return a == b })
}

func (n *LeafImpl) FindKey(key uint64) int {
	return n.GenericLeafNode.FindKey(key, func(a, b uint64) bool { return a == b })
}

func (n *LeafImpl) Contains(key uint64) bool {
	return n.GenericLeafNode.Contains(key, func(a, b uint64) bool { return a == b })
}

// Add methods needed by the tests
func (n *LeafImpl) BorrowFromRight(rightSibling *LeafImpl, leafIndex int, parent *BranchImpl) {
	n.GenericLeafNode.BorrowFromRight(rightSibling.GenericLeafNode, leafIndex, parent.GenericBranchNode)
}

func (n *LeafImpl) BorrowFromLeft(leftSibling *LeafImpl, leafIndex int, parent *BranchImpl) {
	n.GenericLeafNode.BorrowFromLeft(leftSibling.GenericLeafNode, leafIndex, parent.GenericBranchNode)
}

func (n *LeafImpl) MergeWith(other *LeafImpl) {
	n.GenericLeafNode.MergeWith(other.GenericLeafNode)
}

func (n *LeafImpl) TryBorrowFromSibling(parent *BranchImpl, nodeIndex int, branchingFactor int) bool {
	// Try to borrow from right sibling
	if nodeIndex < len(parent.Children())-1 {
		rightSibling, ok := parent.Children()[nodeIndex+1].(*LeafImpl)
		if ok && len(rightSibling.Keys()) > minLeafKeys(branchingFactor) {
			n.BorrowFromRight(rightSibling, nodeIndex, parent)
			return true
		}
	}

	// Try to borrow from left sibling
	if nodeIndex > 0 {
		leftSibling, ok := parent.Children()[nodeIndex-1].(*LeafImpl)
		if ok && len(leftSibling.Keys()) > minLeafKeys(branchingFactor) {
			n.BorrowFromLeft(leftSibling, nodeIndex, parent)
			return true
		}
	}

	return false
}

// Implement the Node interface for BranchImpl

func (n *BranchImpl) Type() NodeType {
	return Branch
}

func (n *BranchImpl) Keys() []uint64 {
	return n.GenericBranchNode.Keys()
}

func (n *BranchImpl) KeyCount() int {
	return n.GenericBranchNode.KeyCount()
}

func (n *BranchImpl) IsFull(branchingFactor int) bool {
	return n.GenericBranchNode.IsFull(branchingFactor)
}

func (n *BranchImpl) IsUnderflow(branchingFactor int) bool {
	return n.GenericBranchNode.IsUnderflow(branchingFactor)
}

func (n *BranchImpl) InsertKey(key uint64) bool {
	return n.GenericBranchNode.InsertKey(key, func(a, b uint64) bool { return a < b })
}

func (n *BranchImpl) DeleteKey(key uint64) bool {
	return n.GenericBranchNode.DeleteKey(key, func(a, b uint64) bool { return a == b })
}

func (n *BranchImpl) FindKey(key uint64) int {
	return n.GenericBranchNode.FindKey(key, func(a, b uint64) bool { return a == b })
}

func (n *BranchImpl) Contains(key uint64) bool {
	return n.GenericBranchNode.Contains(key, func(a, b uint64) bool { return a == b })
}

// Add methods needed by the tests
func (n *BranchImpl) Children() []Node {
	genericChildren := n.GenericBranchNode.Children()
	children := make([]Node, len(genericChildren))
	for i, child := range genericChildren {
		switch c := child.(type) {
		case *GenericLeafNode[uint64]:
			children[i] = &LeafImpl{GenericLeafNode: c}
		case *GenericBranchNode[uint64]:
			children[i] = &BranchImpl{GenericBranchNode: c}
		}
	}
	return children
}

func (n *BranchImpl) SetChild(index int, child Node) {
	switch c := child.(type) {
	case *LeafImpl:
		n.GenericBranchNode.SetChild(index, c.GenericLeafNode)
	case *BranchImpl:
		n.GenericBranchNode.SetChild(index, c.GenericBranchNode)
	}
}

func (n *BranchImpl) RemoveChild(index int) {
	n.GenericBranchNode.RemoveChild(index)
}

func (n *BranchImpl) FindChildIndex(key uint64) int {
	return n.GenericBranchNode.FindChildIndex(key, func(a, b uint64) bool { return a < b })
}

func (n *BranchImpl) InsertKeyWithChild(key uint64, child Node) {
	switch c := child.(type) {
	case *LeafImpl:
		n.GenericBranchNode.InsertKeyWithChild(key, c.GenericLeafNode, func(a, b uint64) bool { return a < b })
	case *BranchImpl:
		n.GenericBranchNode.InsertKeyWithChild(key, c.GenericBranchNode, func(a, b uint64) bool { return a < b })
	}
}

func (n *BranchImpl) BorrowFromRight(separatorKey uint64, rightSibling *BranchImpl, branchIndex int, parent *BranchImpl) {
	n.GenericBranchNode.BorrowFromRight(separatorKey, rightSibling.GenericBranchNode, branchIndex, parent.GenericBranchNode)
}

func (n *BranchImpl) BorrowFromLeft(separatorKey uint64, leftSibling *BranchImpl, branchIndex int, parent *BranchImpl) {
	n.GenericBranchNode.BorrowFromLeft(separatorKey, leftSibling.GenericBranchNode, branchIndex, parent.GenericBranchNode)
}

func (n *BranchImpl) MergeWith(separatorKey uint64, other *BranchImpl) {
	n.GenericBranchNode.MergeWith(separatorKey, other.GenericBranchNode)
}

func (n *BranchImpl) TryBorrowFromSibling(parent *BranchImpl, branchIndex int, branchingFactor int) bool {
	// Try to borrow from right sibling
	if branchIndex < len(parent.Children())-1 {
		rightSibling, ok := parent.Children()[branchIndex+1].(*BranchImpl)
		if ok && len(rightSibling.Keys()) > minInternalKeys(branchingFactor) {
			separatorKey := parent.Keys()[branchIndex]
			n.BorrowFromRight(separatorKey, rightSibling, branchIndex, parent)
			return true
		}
	}

	// Try to borrow from left sibling
	if branchIndex > 0 {
		leftSibling, ok := parent.Children()[branchIndex-1].(*BranchImpl)
		if ok && len(leftSibling.Keys()) > minInternalKeys(branchingFactor) {
			separatorKey := parent.Keys()[branchIndex-1]
			n.BorrowFromLeft(separatorKey, leftSibling, branchIndex, parent)
			return true
		}
	}

	return false
}

// Add methods needed by the edge_cases_test.go file
func (n *BranchImpl) tryBorrowFromLeft(parent *BranchImpl, index int, branchingFactor int) bool {
	if index <= 0 || index-1 >= len(parent.Keys()) {
		return false
	}

	leftSibling, ok := parent.Children()[index-1].(*BranchImpl)
	if !ok || len(leftSibling.Keys()) <= minInternalKeys(branchingFactor) {
		return false
	}

	separatorKey := parent.Keys()[index-1]
	n.BorrowFromLeft(separatorKey, leftSibling, index, parent)
	return true
}

func (n *BranchImpl) tryBorrowFromRight(parent *BranchImpl, index int, branchingFactor int) bool {
	if index >= len(parent.Children())-1 || index >= len(parent.Keys()) {
		return false
	}

	rightSibling, ok := parent.Children()[index+1].(*BranchImpl)
	if !ok || len(rightSibling.Keys()) <= minInternalKeys(branchingFactor) {
		return false
	}

	separatorKey := parent.Keys()[index]
	n.BorrowFromRight(separatorKey, rightSibling, index, parent)
	return true
}
