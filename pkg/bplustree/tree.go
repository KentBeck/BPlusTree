package bplustree

import (
	"fmt"
)

// BPlusTree represents a B+ tree data structure
type BPlusTree struct {
	root            Node
	branchingFactor int
	height          int
	size            int
	bloomFilter     BloomFilterInterface
}

// NewBPlusTree creates a new B+ tree with the given branching factor
func NewBPlusTree(branchingFactor int) *BPlusTree {
	if branchingFactor < 3 {
		branchingFactor = 3 // Minimum branching factor
	}

	// Create a Bloom filter with reasonable default parameters
	// We'll use an expected capacity of 1000 elements and a false positive rate of 0.01
	// These parameters can be tuned based on the expected usage
	bloomSize, hashFunctions := OptimalBloomFilterSize(1000, 0.01)

	return &BPlusTree{
		root:            NewLeafNode(),
		branchingFactor: branchingFactor,
		height:          1,
		size:            0,
		bloomFilter:     NewBloomFilter(bloomSize, hashFunctions),
	}
}

// Size returns the number of keys in the tree
func (t *BPlusTree) Size() int {
	return t.size
}

// Height returns the height of the tree
func (t *BPlusTree) Height() int {
	return t.height
}

// Insert inserts a key into the tree
func (t *BPlusTree) Insert(key uint64) bool {
	// Handle the case where the root is full
	if t.root.IsFull(t.branchingFactor) {
		oldRoot := t.root
		t.root = NewInternalNode()
		newRoot := t.root.(*InternalNodeImpl)
		newRoot.SetChild(0, oldRoot)
		t.splitChild(newRoot, 0)
		t.height++
	}

	// Insert the key
	inserted := t.insertNonFull(t.root, key)
	if inserted {
		t.size++
		// Add the key to the Bloom filter instead of invalidating it
		// If the filter is not valid, we'll recompute it on the next Contains call
		if t.bloomFilter != nil && t.bloomFilter.IsValid() {
			t.bloomFilter.Add(key)
		}
	}
	return inserted
}

// insertNonFull inserts a key into a non-full node
func (t *BPlusTree) insertNonFull(node Node, key uint64) bool {
	switch n := node.(type) {
	case *LeafNodeImpl:
		return n.InsertKey(key)
	case *InternalNodeImpl:
		// Find the child that should contain the key
		childIndex := n.FindChildIndex(key)
		child := n.Children()[childIndex]

		// If the child is full, split it
		if child.IsFull(t.branchingFactor) {
			t.splitChild(n, childIndex)
			// After splitting, determine which child to go to
			if key > n.Keys()[childIndex] {
				childIndex++
			}
			child = n.Children()[childIndex]
		}

		// Recursively insert into the child
		return t.insertNonFull(child, key)
	}
	return false
}

// splitChild splits a full child of an internal node
func (t *BPlusTree) splitChild(parent *InternalNodeImpl, childIndex int) {
	child := parent.Children()[childIndex]

	switch c := child.(type) {
	case *InternalNodeImpl:
		// Split internal node
		newChildImpl := NewInternalNode()
		midIndex := t.branchingFactor/2 - 1
		midKey := c.Keys()[midIndex]

		// Move keys and children to the new node
		newChildImpl.keys = append(newChildImpl.keys, c.Keys()[midIndex+1:]...)
		newChildImpl.children = append(newChildImpl.children, c.Children()[midIndex+1:]...)

		// Update the original child
		c.keys = c.Keys()[:midIndex]
		c.children = c.Children()[:midIndex+1]

		// Insert the new child into the parent
		parent.InsertKeyWithChild(midKey, newChildImpl)

	case *LeafNodeImpl:
		// Split leaf node
		newLeafImpl := NewLeafNode()
		midIndex := t.branchingFactor / 2

		// Move keys to the new leaf
		newLeafImpl.keys = append(newLeafImpl.keys, c.Keys()[midIndex:]...)

		// Update the original leaf
		c.keys = c.Keys()[:midIndex]

		// Update the linked list of leaves
		newLeafImpl.next = c.next
		c.next = newLeafImpl

		// Insert the new leaf into the parent
		// Use the first key of the new leaf as the separator key
		if len(newLeafImpl.Keys()) > 0 {
			parent.InsertKeyWithChild(newLeafImpl.Keys()[0], newLeafImpl)
		}
	}
}

// Contains returns true if the tree contains the key
func (t *BPlusTree) Contains(key uint64) bool {
	// If Bloom filter is disabled, go directly to tree traversal
	if t.bloomFilter == nil {
		return t.findLeaf(t.root, key)
	}

	// Check if the Bloom filter is valid
	if !t.bloomFilter.IsValid() {
		// Recompute the Bloom filter from all keys
		t.recomputeBloomFilter()
	}

	// Check if the key might be in the set using the Bloom filter
	if !t.bloomFilter.Contains(key) {
		// If the Bloom filter says the key is definitely not in the set, return false
		return false
	}

	// If the Bloom filter says the key might be in the set, check the tree
	return t.findLeaf(t.root, key)
}

// recomputeBloomFilter recomputes the Bloom filter from all keys in the tree
func (t *BPlusTree) recomputeBloomFilter() {
	// If Bloom filter is disabled, do nothing
	if t.bloomFilter == nil {
		return
	}

	// Clear the Bloom filter
	t.bloomFilter.Clear()

	// Add all keys to the Bloom filter
	t.addKeysToBloomFilter(t.root)

	// Mark the Bloom filter as valid
	t.bloomFilter.SetValid()
}

// addKeysToBloomFilter adds all keys in the subtree rooted at node to the Bloom filter
func (t *BPlusTree) addKeysToBloomFilter(node Node) {
	switch n := node.(type) {
	case *LeafNodeImpl:
		// Add all keys in the leaf node to the Bloom filter
		for _, key := range n.Keys() {
			t.bloomFilter.Add(key)
		}
	case *InternalNodeImpl:
		// Recursively add keys from all children
		for _, child := range n.Children() {
			t.addKeysToBloomFilter(child)
		}
	}
}

// findLeaf finds the leaf node that should contain the key
func (t *BPlusTree) findLeaf(node Node, key uint64) bool {
	switch n := node.(type) {
	case *LeafNodeImpl:
		return n.Contains(key)
	case *InternalNodeImpl:
		childIndex := n.FindChildIndex(key)
		// Safety check to avoid index out of range
		if childIndex >= len(n.Children()) {
			return false
		}
		return t.findLeaf(n.Children()[childIndex], key)
	}
	return false
}

// Delete removes a key from the tree
func (t *BPlusTree) Delete(key uint64) bool {
	// Try to delete the key
	deleted, _ := t.deleteAndBalance(t.root, nil, -1, key)
	if deleted {
		t.size--

		// If the root is an internal node with no keys, make its only child the new root
		if t.root.Type() == InternalNode && len(t.root.Keys()) == 0 {
			internalRoot := t.root.(*InternalNodeImpl)
			if len(internalRoot.Children()) > 0 {
				t.root = internalRoot.Children()[0]
				t.height--
			}
		}

		// Invalidate the Bloom filter since we've modified the tree
		t.bloomFilter.Clear()
	}
	return deleted
}

// deleteAndBalance removes a key from a node and balances the tree if necessary
// Returns:
// - deleted: true if the key was deleted
// - keyToReplaceInParent: a key that needs to be replaced in the parent (for internal nodes)
func (t *BPlusTree) deleteAndBalance(node Node, parent *InternalNodeImpl, parentChildIndex int, key uint64) (bool, uint64) {
	switch n := node.(type) {
	case *LeafNodeImpl:
		// Try to delete the key
		if !n.DeleteKey(key) {
			return false, 0
		}

		// If this is the root or it doesn't underflow, we're done
		if parent == nil || !n.IsUnderflow(t.branchingFactor) {
			return true, 0
		}

		// Handle underflow by borrowing or merging
		return true, t.balanceLeafNode(n, parent, parentChildIndex)

	case *InternalNodeImpl:
		// Find the child that should contain the key
		childIndex := n.FindChildIndex(key)
		// Check if childIndex is valid
		if childIndex >= len(n.Children()) {
			return false, 0
		}

		child := n.Children()[childIndex]

		// Recursively delete from the child
		deleted, keyToReplace := t.deleteAndBalance(child, n, childIndex, key)
		if !deleted {
			return false, 0
		}

		// If we need to replace a key in this node
		if keyToReplace != 0 && len(n.keys) > 0 {
			// Find the key to replace (it's the key at childIndex-1 unless childIndex is 0)
			if childIndex > 0 && childIndex-1 < len(n.keys) {
				n.keys[childIndex-1] = keyToReplace
			}
		}

		// If this is the root or it doesn't underflow, we're done
		if parent == nil || !n.IsUnderflow(t.branchingFactor) {
			return true, 0
		}

		// Handle underflow by borrowing or merging
		return true, t.balanceInternalNode(n, parent, parentChildIndex)
	}

	return false, 0
}

// balanceLeafNode handles underflow in a leaf node by borrowing from siblings or merging
// Returns a key that needs to be replaced in the parent (if any)
func (t *BPlusTree) balanceLeafNode(node *LeafNodeImpl, parent *InternalNodeImpl, nodeIndex int) uint64 {
	// Safety check
	if nodeIndex < 0 || nodeIndex >= len(parent.Children()) {
		return 0
	}

	// Try to borrow from right sibling
	if nodeIndex < len(parent.Children())-1 && nodeIndex < len(parent.Keys()) {
		rightSibling, ok := parent.Children()[nodeIndex+1].(*LeafNodeImpl)
		if !ok {
			return 0
		}

		// Check if right sibling can spare a key
		minKeys := (t.branchingFactor + 1) / 2
		if len(rightSibling.Keys()) > minKeys {
			node.BorrowFromRight(rightSibling, nodeIndex, parent)
			return 0
		}
	}

	// Try to borrow from left sibling
	if nodeIndex > 0 && nodeIndex-1 < len(parent.Keys()) {
		leftSibling, ok := parent.Children()[nodeIndex-1].(*LeafNodeImpl)
		if !ok {
			return 0
		}

		// Check if left sibling can spare a key
		minKeys := (t.branchingFactor + 1) / 2
		if len(leftSibling.Keys()) > minKeys {
			node.BorrowFromLeft(leftSibling)
			// Update the separator key in the parent
			if len(node.Keys()) > 0 {
				parent.keys[nodeIndex-1] = node.Keys()[0]
			}
			return 0
		}
	}

	// Merge with a sibling
	// Prefer merging with left sibling
	if nodeIndex > 0 && nodeIndex-1 < len(parent.Keys()) {
		leftSibling, ok := parent.Children()[nodeIndex-1].(*LeafNodeImpl)
		if !ok {
			return 0
		}

		leftSibling.MergeWith(node)

		// Remove the separator key and the right child pointer from the parent
		if nodeIndex-1 < len(parent.Keys()) {
			parent.DeleteKey(parent.Keys()[nodeIndex-1])
			parent.RemoveChild(nodeIndex)
		}

		return 0
	}

	// Merge with right sibling
	if nodeIndex < len(parent.Children())-1 && nodeIndex < len(parent.Keys()) {
		rightSibling, ok := parent.Children()[nodeIndex+1].(*LeafNodeImpl)
		if !ok {
			return 0
		}

		node.MergeWith(rightSibling)

		// Remove the separator key and the right child pointer from the parent
		if nodeIndex < len(parent.Keys()) {
			parent.DeleteKey(parent.Keys()[nodeIndex])
			parent.RemoveChild(nodeIndex + 1)
		}

		return 0
	}

	// This should never happen
	return 0
}

// balanceInternalNode handles underflow in an internal node by borrowing from siblings or merging
// Returns a key that needs to be replaced in the parent (if any)
func (t *BPlusTree) balanceInternalNode(node *InternalNodeImpl, parent *InternalNodeImpl, nodeIndex int) uint64 {
	// Safety check
	if nodeIndex < 0 || nodeIndex >= len(parent.Children()) {
		return 0
	}

	// Try to borrow from right sibling
	if nodeIndex < len(parent.Children())-1 && nodeIndex < len(parent.Keys()) {
		rightSibling, ok := parent.Children()[nodeIndex+1].(*InternalNodeImpl)
		if !ok {
			return 0
		}

		// Check if right sibling can spare a key
		minKeys := (t.branchingFactor+1)/2 - 1
		if len(rightSibling.Keys()) > minKeys {
			separatorKey := parent.Keys()[nodeIndex]
			node.BorrowFromRight(separatorKey, rightSibling, nodeIndex, parent)
			return 0
		}
	}

	// Try to borrow from left sibling
	if nodeIndex > 0 && nodeIndex-1 < len(parent.Keys()) {
		leftSibling, ok := parent.Children()[nodeIndex-1].(*InternalNodeImpl)
		if !ok {
			return 0
		}

		// Check if left sibling can spare a key
		minKeys := (t.branchingFactor+1)/2 - 1
		if len(leftSibling.Keys()) > minKeys {
			separatorKey := parent.Keys()[nodeIndex-1]
			node.BorrowFromLeft(separatorKey, leftSibling, nodeIndex, parent)
			return 0
		}
	}

	// Merge with a sibling
	// Prefer merging with left sibling
	if nodeIndex > 0 && nodeIndex-1 < len(parent.Keys()) {
		leftSibling, ok := parent.Children()[nodeIndex-1].(*InternalNodeImpl)
		if !ok {
			return 0
		}

		separatorKey := parent.Keys()[nodeIndex-1]
		leftSibling.MergeWith(separatorKey, node)

		// Remove the separator key and the right child pointer from the parent
		parent.DeleteKey(separatorKey)
		parent.RemoveChild(nodeIndex)

		return 0
	}

	// Merge with right sibling
	if nodeIndex < len(parent.Children())-1 && nodeIndex < len(parent.Keys()) {
		rightSibling, ok := parent.Children()[nodeIndex+1].(*InternalNodeImpl)
		if !ok {
			return 0
		}

		separatorKey := parent.Keys()[nodeIndex]
		node.MergeWith(separatorKey, rightSibling)

		// Remove the separator key and the right child pointer from the parent
		parent.DeleteKey(separatorKey)
		parent.RemoveChild(nodeIndex + 1)

		return 0
	}

	// This should never happen
	return 0
}

// String returns a string representation of the tree
func (t *BPlusTree) String() string {
	return fmt.Sprintf("BPlusTree(size=%d, height=%d, branchingFactor=%d)",
		t.size, t.height, t.branchingFactor)
}

// SetBloomFilterParams sets the parameters for the Bloom filter
func (t *BPlusTree) SetBloomFilterParams(size int, hashFunctions int) {
	t.bloomFilter = NewBloomFilter(size, hashFunctions)
}

// SetCustomBloomFilter sets a custom Bloom filter implementation
func (t *BPlusTree) SetCustomBloomFilter(filter interface{}) {
	if bf, ok := filter.(BloomFilterInterface); ok {
		t.bloomFilter = bf
	}
}

// DisableBloomFilter disables the Bloom filter by making Contains bypass it
func (t *BPlusTree) DisableBloomFilter() {
	// Set a nil Bloom filter to disable it
	t.bloomFilter = nil
}
