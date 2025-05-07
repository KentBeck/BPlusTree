package bplustree

import (
	"fmt"
)

// Helper functions for minimum key calculations
func minInternalKeys(branchingFactor int) int {
	return (branchingFactor+1)/2 - 1
}

func minLeafKeys(branchingFactor int) int {
	return (branchingFactor + 1) / 2
}

// BPlusTree represents a B+ tree data structure
type BPlusTree struct {
	root            Node
	branchingFactor int
	size            int
	bloomFilter     BloomFilterInterface
}

// NewBPlusTree creates a new B+ tree with the given branching factor
func NewBPlusTree(branchingFactor int) *BPlusTree {
	return NewBPlusTreeWithOptions(branchingFactor, true)
}

// NewBPlusTreeWithOptions creates a new B+ tree with the given branching factor and bloom filter option
func NewBPlusTreeWithOptions(branchingFactor int, useBloomFilter bool) *BPlusTree {
	if branchingFactor < 3 {
		branchingFactor = 3 // Minimum branching factor
	}

	var bloomFilter BloomFilterInterface
	if useBloomFilter {
		// Create a Bloom filter with reasonable default parameters
		// We'll use an expected capacity of 1000 elements and a false positive rate of 0.01
		bloomSize, hashFunctions := OptimalBloomFilterSize(1000, 0.01)
		bloomFilter = NewBloomFilter(bloomSize, hashFunctions)
	} else {
		// Use a NullBloomFilter that always returns "maybe"
		bloomFilter = NewNullBloomFilter()
	}

	return &BPlusTree{
		root:            NewLeaf(),
		branchingFactor: branchingFactor,
		size:            0,
		bloomFilter:     bloomFilter,
	}
}

// Size returns the number of keys in the tree
func (t *BPlusTree) Size() int {
	return t.size
}

// Height returns the height of the tree by calculating it on the fly
func (t *BPlusTree) Height() int {
	return t.calculateHeight(t.root)
}

// calculateHeight calculates the height of the subtree rooted at node
func (t *BPlusTree) calculateHeight(node Node) int {
	if node.Type() == Leaf {
		return 1
	}

	// For branch nodes, recursively calculate height of first child
	// All children should have the same height in a balanced B+ tree
	branchNode := node.(*BranchImpl)
	if len(branchNode.Children()) == 0 {
		return 1 // Empty branch node (shouldn't happen in a valid tree)
	}

	return 1 + t.calculateHeight(branchNode.Children()[0])
}

// Insert inserts a key into the tree
func (t *BPlusTree) Insert(key uint64) bool {
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

// splitRoot handles splitting the root when it's full
func (t *BPlusTree) splitRoot() {
	oldRoot := t.root
	t.root = NewBranch()
	newRoot := t.root.(*BranchImpl)
	newRoot.SetChild(0, oldRoot)
	t.splitChild(newRoot, 0)
	// Height increases automatically when the root is split
	// No need to update a height field since we calculate it on demand
}

// updateBloomFilter adds a key to the bloom filter if it's valid
func (t *BPlusTree) updateBloomFilter(key uint64) {
	if t.bloomFilter != nil && t.bloomFilter.IsValid() {
		t.bloomFilter.Add(key)
	}
}

// insertNonFull inserts a key into a non-full node
func (t *BPlusTree) insertNonFull(node Node, key uint64) bool {
	switch n := node.(type) {
	case *LeafImpl:
		return n.InsertKey(key)
	case *BranchImpl:
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
func (t *BPlusTree) splitChild(parent *BranchImpl, childIndex int) {
	child := parent.Children()[childIndex]

	switch c := child.(type) {
	case *BranchImpl:
		// Split internal node
		newChildImpl := NewBranch()
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

	case *LeafImpl:
		// Split leaf node
		newLeafImpl := NewLeaf()
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
	// Ensure bloom filter is valid
	if !t.bloomFilter.IsValid() {
		t.recomputeBloomFilter()
	}

	// Early return if bloom filter says key is definitely not present
	if !t.bloomFilter.Contains(key) {
		return false
	}

	// Check the tree since bloom filter says key might be present
	return t.findLeaf(t.root, key)
}

// recomputeBloomFilter recomputes the Bloom filter from all keys in the tree
func (t *BPlusTree) recomputeBloomFilter() {
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
	case *LeafImpl:
		// Add all keys in the leaf node to the Bloom filter
		for _, key := range n.Keys() {
			t.bloomFilter.Add(key)
		}
	case *BranchImpl:
		// Recursively add keys from all children
		for _, child := range n.Children() {
			t.addKeysToBloomFilter(child)
		}
	}
}

// findLeaf finds the leaf node that should contain the key
func (t *BPlusTree) findLeaf(node Node, key uint64) bool {
	switch n := node.(type) {
	case *LeafImpl:
		return n.Contains(key)
	case *BranchImpl:
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
	deleted, _ := t.deleteAndBalance(t.root, nil, -1, key)

	if deleted {
		t.decrementSize()
		t.handleRootUnderflow()
		t.invalidateBloomFilter()
	}

	return deleted
}

// decrementSize decrements the size counter if it's greater than 0
func (t *BPlusTree) decrementSize() {
	if t.size > 0 {
		t.size--
	}
}

// handleRootUnderflow handles the case where the root has no keys
func (t *BPlusTree) handleRootUnderflow() {
	if t.isEmptyInternalRoot() {
		t.promoteOnlyChild()
	}
}

// isEmptyInternalRoot returns true if the root is an internal node with no keys
func (t *BPlusTree) isEmptyInternalRoot() bool {
	return t.root.Type() == Branch &&
		len(t.root.Keys()) == 0 &&
		len(t.root.(*BranchImpl).Children()) > 0
}

// promoteOnlyChild makes the only child of the root the new root
func (t *BPlusTree) promoteOnlyChild() {
	t.root = t.root.(*BranchImpl).Children()[0]
	// Height decreases automatically when the root's only child becomes the new root
	// No need to update a height field since we calculate it on demand
}

// invalidateBloomFilter clears the bloom filter
func (t *BPlusTree) invalidateBloomFilter() {
	t.bloomFilter.Clear()
}

// deleteAndBalance removes a key from a node and balances the tree if necessary
// Returns:
// - deleted: true if the key was deleted
// - keyToReplaceInParent: a key that needs to be replaced in the parent (for internal nodes)
func (t *BPlusTree) deleteAndBalance(node Node, parent *BranchImpl, parentChildIndex int, key uint64) (bool, uint64) {
	switch n := node.(type) {
	case *LeafImpl:
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

	case *BranchImpl:
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
func (t *BPlusTree) balanceLeafNode(node *LeafImpl, parent *BranchImpl, nodeIndex int) uint64 {
	// Safety check
	if !t.isValidNodeIndex(parent, nodeIndex) {
		return 0
	}

	// Try borrowing from siblings
	if node.TryBorrowFromSibling(parent, nodeIndex, t.branchingFactor) {
		return 0
	}

	// If borrowing failed, merge with a sibling
	return t.mergeLeafWithSibling(node, parent, nodeIndex)
}

// isValidNodeIndex checks if the node index is valid
func (t *BPlusTree) isValidNodeIndex(parent *BranchImpl, nodeIndex int) bool {
	return nodeIndex >= 0 && nodeIndex < len(parent.Children())
}

// mergeLeafWithSibling merges the node with a sibling
func (t *BPlusTree) mergeLeafWithSibling(node *LeafImpl, parent *BranchImpl, nodeIndex int) uint64 {
	// Try to merge with left sibling first
	if t.mergeWithLeftLeaf(node, parent, nodeIndex) {
		return 0
	}

	// If that fails, try to merge with right sibling
	if t.mergeWithRightLeaf(node, parent, nodeIndex) {
		return 0
	}

	// This should never happen
	return 0
}

// mergeWithLeftLeaf attempts to merge with the left sibling
func (t *BPlusTree) mergeWithLeftLeaf(node *LeafImpl, parent *BranchImpl, nodeIndex int) bool {
	if nodeIndex <= 0 || nodeIndex-1 >= len(parent.Keys()) {
		return false
	}

	leftSibling, ok := parent.Children()[nodeIndex-1].(*LeafImpl)
	if !ok {
		return false
	}

	leftSibling.MergeWith(node)

	// Remove the separator key and the right child pointer from the parent
	if nodeIndex-1 < len(parent.Keys()) {
		parent.DeleteKey(parent.Keys()[nodeIndex-1])
		parent.RemoveChild(nodeIndex)
	}

	return true
}

// mergeWithRightLeaf attempts to merge with the right sibling
func (t *BPlusTree) mergeWithRightLeaf(node *LeafImpl, parent *BranchImpl, nodeIndex int) bool {
	if nodeIndex >= len(parent.Children())-1 || nodeIndex >= len(parent.Keys()) {
		return false
	}

	rightSibling, ok := parent.Children()[nodeIndex+1].(*LeafImpl)
	if !ok {
		return false
	}

	node.MergeWith(rightSibling)

	// Remove the separator key and the right child pointer from the parent
	if nodeIndex < len(parent.Keys()) {
		parent.DeleteKey(parent.Keys()[nodeIndex])
		parent.RemoveChild(nodeIndex + 1)
	}

	return true
}

// balanceInternalNode handles underflow in an internal node by borrowing from siblings or merging
// Returns a key that needs to be replaced in the parent (if any)
func (t *BPlusTree) balanceInternalNode(node *BranchImpl, parent *BranchImpl, nodeIndex int) uint64 {
	// Safety check
	if !t.isValidNodeIndex(parent, nodeIndex) {
		return 0
	}

	// Try borrowing from siblings
	if node.TryBorrowFromSibling(parent, nodeIndex, t.branchingFactor) {
		return 0
	}

	// If borrowing failed, merge with a sibling
	return t.mergeInternalWithSibling(node, parent, nodeIndex)
}

// mergeInternalWithSibling merges the node with a sibling
func (t *BPlusTree) mergeInternalWithSibling(node *BranchImpl, parent *BranchImpl, nodeIndex int) uint64 {
	// Try to merge with left sibling first
	if t.mergeWithLeftInternal(node, parent, nodeIndex) {
		return 0
	}

	// If that fails, try to merge with right sibling
	if t.mergeWithRightInternal(node, parent, nodeIndex) {
		return 0
	}

	// This should never happen
	return 0
}

// mergeWithLeftInternal attempts to merge with the left sibling
func (t *BPlusTree) mergeWithLeftInternal(node *BranchImpl, parent *BranchImpl, nodeIndex int) bool {
	if nodeIndex <= 0 || nodeIndex-1 >= len(parent.Keys()) {
		return false
	}

	leftSibling, ok := parent.Children()[nodeIndex-1].(*BranchImpl)
	if !ok {
		return false
	}

	separatorKey := parent.Keys()[nodeIndex-1]
	leftSibling.MergeWith(separatorKey, node)

	// Remove the separator key and the right child pointer from the parent
	parent.DeleteKey(separatorKey)
	parent.RemoveChild(nodeIndex)

	return true
}

// mergeWithRightInternal attempts to merge with the right sibling
func (t *BPlusTree) mergeWithRightInternal(node *BranchImpl, parent *BranchImpl, nodeIndex int) bool {
	if nodeIndex >= len(parent.Children())-1 || nodeIndex >= len(parent.Keys()) {
		return false
	}

	rightSibling, ok := parent.Children()[nodeIndex+1].(*BranchImpl)
	if !ok {
		return false
	}

	separatorKey := parent.Keys()[nodeIndex]
	node.MergeWith(separatorKey, rightSibling)

	// Remove the separator key and the right child pointer from the parent
	parent.DeleteKey(separatorKey)
	parent.RemoveChild(nodeIndex + 1)

	return true
}

// String returns a string representation of the tree
func (t *BPlusTree) String() string {
	return fmt.Sprintf("BPlusTree(size=%d, height=%d, branchingFactor=%d)",
		t.size, t.Height(), t.branchingFactor)
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

// CountKeys returns the actual number of keys in the tree by traversing it
func (t *BPlusTree) CountKeys() int {
	return t.countKeysInNode(t.root)
}

// countKeysInNode counts the number of keys in a subtree rooted at node
func (t *BPlusTree) countKeysInNode(node Node) int {
	switch n := node.(type) {
	case *LeafImpl:
		return len(n.Keys())
	case *BranchImpl:
		count := 0
		for _, child := range n.Children() {
			count += t.countKeysInNode(child)
		}
		return count
	}
	return 0
}

// ResetSize resets the size counter to the actual number of keys in the tree
func (t *BPlusTree) ResetSize() {
	t.size = t.CountKeys()
}

// GetAllKeys returns all keys in the tree
func (t *BPlusTree) GetAllKeys() []uint64 {
	keys := make([]uint64, 0, t.size)
	t.collectKeys(t.root, &keys)
	return keys
}

// collectKeys collects all keys in the subtree rooted at node
func (t *BPlusTree) collectKeys(node Node, keys *[]uint64) {
	switch n := node.(type) {
	case *LeafImpl:
		*keys = append(*keys, n.Keys()...)
	case *BranchImpl:
		for _, child := range n.Children() {
			t.collectKeys(child, keys)
		}
	}
}

// DeleteAll deletes all keys from the tree
func (t *BPlusTree) DeleteAll() {
	// Reset the tree to an empty leaf node
	t.root = NewLeaf()
	t.size = 0

	// Invalidate the Bloom filter
	t.bloomFilter.Clear()
}

// ForceDeleteKeys forcefully deletes all keys in the given slice
func (t *BPlusTree) ForceDeleteKeys(keys []uint64) int {
	// Get all keys currently in the tree
	keysInTree := t.GetAllKeys()

	// Create a map of keys to delete for O(1) lookup
	keysToDelete := make(map[uint64]bool)
	for _, key := range keys {
		keysToDelete[key] = true
	}

	// Filter out keys that are not in the tree
	keysToDeleteInTree := make([]uint64, 0)
	for _, key := range keysInTree {
		if keysToDelete[key] {
			keysToDeleteInTree = append(keysToDeleteInTree, key)
		}
	}

	// If there are no keys to delete, return 0
	if len(keysToDeleteInTree) == 0 {
		return 0
	}

	// If we need to delete all keys in the tree, use DeleteAll
	if len(keysToDeleteInTree) == len(keysInTree) {
		t.DeleteAll()
		return len(keysToDeleteInTree)
	}

	// Otherwise, delete each key individually
	deletedCount := 0
	for _, key := range keysToDeleteInTree {
		if t.Delete(key) {
			deletedCount++
		}
	}

	return deletedCount
}
