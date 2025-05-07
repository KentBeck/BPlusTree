package bplustree

import (
	"fmt"
)

// GenericBPlusTree is a B+ tree that works with any comparable type
type GenericBPlusTree[K any] struct {
	root            GenericNode[K]
	branchingFactor int
	height          int
	size            int
	less            func(a, b K) bool
	equal           func(a, b K) bool
	hashFunc        func(K) uint64
	bloomFilter     BloomFilterInterface
}

// NewGenericBPlusTree creates a new generic B+ tree
func NewGenericBPlusTree[K any](
	branchingFactor int,
	less func(a, b K) bool,
	equal func(a, b K) bool,
	hashFunc func(K) uint64,
) *GenericBPlusTree[K] {
	if branchingFactor < 3 {
		branchingFactor = 3 // Minimum branching factor
	}

	// Create a Bloom filter with reasonable default parameters
	bloomSize, hashFunctions := OptimalBloomFilterSize(1000, 0.01)

	return &GenericBPlusTree[K]{
		root:            NewGenericLeafNode[K](),
		branchingFactor: branchingFactor,
		height:          1,
		size:            0,
		less:            less,
		equal:           equal,
		hashFunc:        hashFunc,
		bloomFilter:     NewBloomFilter(bloomSize, hashFunctions),
	}
}

// Size returns the number of keys in the tree
func (t *GenericBPlusTree[K]) Size() int {
	return t.size
}

// Height returns the height of the tree
func (t *GenericBPlusTree[K]) Height() int {
	return t.height
}

// Insert inserts a key into the tree
func (t *GenericBPlusTree[K]) Insert(key K) bool {
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
func (t *GenericBPlusTree[K]) splitRoot() {
	oldRoot := t.root
	t.root = NewGenericBranchNode[K]()
	newRoot := t.root.(*GenericBranchNode[K])
	newRoot.SetChild(0, oldRoot)
	t.splitChild(newRoot, 0)
	t.height++
}

// updateBloomFilter adds a key to the bloom filter if it's valid
func (t *GenericBPlusTree[K]) updateBloomFilter(key K) {
	hash := t.hashFunc(key)
	if t.bloomFilter.IsValid() {
		t.bloomFilter.Add(hash)
	}
}

// insertNonFull inserts a key into a non-full node
func (t *GenericBPlusTree[K]) insertNonFull(node GenericNode[K], key K) bool {
	switch n := node.(type) {
	case *GenericLeafNode[K]:
		return n.InsertKey(key, t.less)
	case *GenericBranchNode[K]:
		// Find the child that should contain the key
		childIndex := n.FindChildIndex(key, t.less)
		child := n.Children()[childIndex]

		// If the child is full, split it
		if child.IsFull(t.branchingFactor) {
			t.splitChild(n, childIndex)
			// After splitting, determine which child to go to
			if childIndex < len(n.Keys()) && !t.less(key, n.Keys()[childIndex]) {
				childIndex++
			}
			child = n.Children()[childIndex]
		}

		// Recursively insert into the child
		return t.insertNonFull(child, key)
	}
	return false
}

// splitChild splits a full child of a branch node
func (t *GenericBPlusTree[K]) splitChild(parent *GenericBranchNode[K], childIndex int) {
	child := parent.Children()[childIndex]

	switch c := child.(type) {
	case *GenericBranchNode[K]:
		// Split branch node
		newChildImpl := NewGenericBranchNode[K]()
		midIndex := t.branchingFactor/2 - 1
		midKey := c.Keys()[midIndex]

		// Move keys and children to the new node
		newChildImpl.keys = append(newChildImpl.keys, c.Keys()[midIndex+1:]...)
		newChildImpl.children = append(newChildImpl.children, c.Children()[midIndex+1:]...)

		// Update the original child
		c.keys = c.Keys()[:midIndex]
		c.children = c.Children()[:midIndex+1]

		// Insert the new child into the parent
		parent.InsertKeyWithChild(midKey, newChildImpl, t.less)

	case *GenericLeafNode[K]:
		// Split leaf node
		newLeafImpl := NewGenericLeafNode[K]()
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
			parent.InsertKeyWithChild(newLeafImpl.Keys()[0], newLeafImpl, t.less)
		}
	}
}

// Contains returns true if the tree contains the key
func (t *GenericBPlusTree[K]) Contains(key K) bool {
	// Ensure bloom filter is valid
	if !t.bloomFilter.IsValid() {
		t.recomputeBloomFilter()
	}

	// Early return if bloom filter says key is definitely not present
	hash := t.hashFunc(key)
	if !t.bloomFilter.Contains(hash) {
		return false
	}

	// Check the tree since bloom filter says key might be present
	return t.findLeaf(t.root, key)
}

// recomputeBloomFilter recomputes the Bloom filter from all keys in the tree
func (t *GenericBPlusTree[K]) recomputeBloomFilter() {
	// Clear the Bloom filter
	t.bloomFilter.Clear()

	// Add all keys to the Bloom filter
	t.addKeysToBloomFilter(t.root)

	// Mark the Bloom filter as valid
	t.bloomFilter.SetValid()
}

// addKeysToBloomFilter adds all keys in the subtree rooted at node to the Bloom filter
func (t *GenericBPlusTree[K]) addKeysToBloomFilter(node GenericNode[K]) {
	switch n := node.(type) {
	case *GenericLeafNode[K]:
		// Add all keys in the leaf node to the Bloom filter
		for _, key := range n.Keys() {
			hash := t.hashFunc(key)
			t.bloomFilter.Add(hash)
		}
	case *GenericBranchNode[K]:
		// Recursively add keys from all children
		for _, child := range n.Children() {
			t.addKeysToBloomFilter(child)
		}
	}
}

// findLeaf finds the leaf node that should contain the key
func (t *GenericBPlusTree[K]) findLeaf(node GenericNode[K], key K) bool {
	switch n := node.(type) {
	case *GenericLeafNode[K]:
		return n.Contains(key, t.equal)
	case *GenericBranchNode[K]:
		childIndex := n.FindChildIndex(key, t.less)
		// Safety check to avoid index out of range
		if childIndex >= len(n.Children()) {
			return false
		}
		return t.findLeaf(n.Children()[childIndex], key)
	}
	return false
}

// findLeafNode finds and returns the leaf node that should contain the key
func (t *GenericBPlusTree[K]) findLeafNode(node GenericNode[K], key K) *GenericLeafNode[K] {
	switch n := node.(type) {
	case *GenericLeafNode[K]:
		return n
	case *GenericBranchNode[K]:
		childIndex := n.FindChildIndex(key, t.less)
		// Safety check to avoid index out of range
		if childIndex >= len(n.Children()) {
			return nil
		}
		return t.findLeafNode(n.Children()[childIndex], key)
	}
	return nil
}

// Delete removes a key from the tree
func (t *GenericBPlusTree[K]) Delete(key K) bool {
	deleted, _ := t.deleteAndBalance(t.root, nil, -1, key)

	if deleted {
		t.decrementSize()
		t.handleRootUnderflow()
		t.invalidateBloomFilter()
	}

	return deleted
}

// decrementSize decrements the size counter if it's greater than 0
func (t *GenericBPlusTree[K]) decrementSize() {
	if t.size > 0 {
		t.size--
	}
}

// handleRootUnderflow handles the case where the root has no keys
func (t *GenericBPlusTree[K]) handleRootUnderflow() {
	if t.isEmptyInternalRoot() {
		t.promoteOnlyChild()
	}
}

// isEmptyInternalRoot returns true if the root is an internal node with no keys
func (t *GenericBPlusTree[K]) isEmptyInternalRoot() bool {
	if branch, ok := t.root.(*GenericBranchNode[K]); ok {
		return len(branch.Keys()) == 0 && len(branch.Children()) > 0
	}
	return false
}

// promoteOnlyChild makes the only child of the root the new root
func (t *GenericBPlusTree[K]) promoteOnlyChild() {
	if branch, ok := t.root.(*GenericBranchNode[K]); ok {
		t.root = branch.Children()[0]
		t.height--
	}
}

// invalidateBloomFilter clears the bloom filter
func (t *GenericBPlusTree[K]) invalidateBloomFilter() {
	t.bloomFilter.Clear()
}

// deleteAndBalance removes a key from a node and balances the tree if necessary
// Returns:
// - deleted: true if the key was deleted
// - keyToReplaceInParent: a key that needs to be replaced in the parent (for internal nodes)
func (t *GenericBPlusTree[K]) deleteAndBalance(node GenericNode[K], parent *GenericBranchNode[K], parentChildIndex int, key K) (bool, K) {
	var zeroKey K // Zero value of K

	switch n := node.(type) {
	case *GenericLeafNode[K]:
		// Try to delete the key
		if !n.DeleteKey(key, t.equal) {
			return false, zeroKey
		}

		// If this is the root or it doesn't underflow, we're done
		if parent == nil || !n.IsUnderflow(t.branchingFactor) {
			return true, zeroKey
		}

		// Handle underflow by borrowing or merging
		return t.handleLeafUnderflow(n, parent, parentChildIndex), zeroKey

	case *GenericBranchNode[K]:
		// Find the child that should contain the key
		keyIndex := n.FindKey(key, t.equal)
		if keyIndex != -1 {
			// The key is in this internal node, so we need to find a replacement
			// Get the rightmost key from the left subtree
			leftChild := n.Children()[keyIndex]
			replacementKey, success := t.findAndRemoveMax(leftChild, n, keyIndex)
			if success {
				// Replace the key in this node
				n.keys[keyIndex] = replacementKey
				return true, zeroKey
			}
			return false, zeroKey
		}

		// The key is not in this node, so we need to find the child that should contain it
		childIndex := n.FindChildIndex(key, t.less)
		if childIndex >= len(n.Children()) {
			return false, zeroKey
		}

		// Recursively delete from the child
		deleted, keyToReplace := t.deleteAndBalance(n.Children()[childIndex], n, childIndex, key)
		if !deleted {
			return false, zeroKey
		}

		// If we need to replace a key in this node
		if !t.equal(keyToReplace, zeroKey) {
			if childIndex > 0 {
				n.keys[childIndex-1] = keyToReplace
			}
		}

		// Check if the child underflowed and needs rebalancing
		child := n.Children()[childIndex]
		if child.IsUnderflow(t.branchingFactor) {
			return true, t.handleBranchUnderflow(n, childIndex)
		}

		return true, zeroKey
	}

	return false, zeroKey
}

// findAndRemoveMax finds and removes the maximum key in the subtree rooted at node
func (t *GenericBPlusTree[K]) findAndRemoveMax(node GenericNode[K], parent *GenericBranchNode[K], parentChildIndex int) (K, bool) {
	var zeroKey K // Zero value of K

	switch n := node.(type) {
	case *GenericLeafNode[K]:
		if len(n.Keys()) == 0 {
			return zeroKey, false
		}
		// Get the maximum key
		maxKey := n.Keys()[len(n.Keys())-1]
		// Remove it
		n.keys = n.Keys()[:len(n.Keys())-1]
		// Handle underflow if necessary
		if parent != nil && n.IsUnderflow(t.branchingFactor) {
			t.handleLeafUnderflow(n, parent, parentChildIndex)
		}
		return maxKey, true

	case *GenericBranchNode[K]:
		// Recursively find and remove the maximum key from the rightmost child
		childIndex := len(n.Children()) - 1
		maxKey, success := t.findAndRemoveMax(n.Children()[childIndex], n, childIndex)
		if !success {
			return zeroKey, false
		}
		// Handle underflow if necessary
		child := n.Children()[childIndex]
		if child.IsUnderflow(t.branchingFactor) {
			t.handleBranchUnderflow(n, childIndex)
		}
		return maxKey, true
	}

	return zeroKey, false
}

// handleLeafUnderflow handles the case where a leaf node has too few keys
func (t *GenericBPlusTree[K]) handleLeafUnderflow(leaf *GenericLeafNode[K], parent *GenericBranchNode[K], leafIndex int) bool {
	// Try to borrow from siblings
	if t.tryBorrowFromSiblingLeaf(leaf, parent, leafIndex) {
		return true
	}

	// If borrowing fails, merge with a sibling
	return t.mergeLeafWithSibling(leaf, parent, leafIndex)
}

// tryBorrowFromSiblingLeaf tries to borrow a key from a sibling leaf
func (t *GenericBPlusTree[K]) tryBorrowFromSiblingLeaf(leaf *GenericLeafNode[K], parent *GenericBranchNode[K], leafIndex int) bool {
	// Try to borrow from right sibling
	if leafIndex < len(parent.Children())-1 {
		rightSibling, ok := parent.Children()[leafIndex+1].(*GenericLeafNode[K])
		if ok && len(rightSibling.Keys()) > minLeafKeys(t.branchingFactor) {
			leaf.BorrowFromRight(rightSibling, leafIndex, parent)
			return true
		}
	}

	// Try to borrow from left sibling
	if leafIndex > 0 {
		leftSibling, ok := parent.Children()[leafIndex-1].(*GenericLeafNode[K])
		if ok && len(leftSibling.Keys()) > minLeafKeys(t.branchingFactor) {
			leaf.BorrowFromLeft(leftSibling, leafIndex, parent)
			return true
		}
	}

	return false
}

// mergeLeafWithSibling merges a leaf node with one of its siblings
func (t *GenericBPlusTree[K]) mergeLeafWithSibling(leaf *GenericLeafNode[K], parent *GenericBranchNode[K], leafIndex int) bool {
	// Try to merge with left sibling
	if leafIndex > 0 {
		leftSibling, ok := parent.Children()[leafIndex-1].(*GenericLeafNode[K])
		if ok {
			// Merge leaf into left sibling
			leftSibling.MergeWith(leaf)
			// Remove the separator key and the leaf from the parent
			parent.DeleteKey(parent.Keys()[leafIndex-1], t.equal)
			parent.RemoveChild(leafIndex)
			return true
		}
	}

	// Try to merge with right sibling
	if leafIndex < len(parent.Children())-1 {
		rightSibling, ok := parent.Children()[leafIndex+1].(*GenericLeafNode[K])
		if ok {
			// Merge right sibling into leaf
			leaf.MergeWith(rightSibling)
			// Remove the separator key and the right sibling from the parent
			parent.DeleteKey(parent.Keys()[leafIndex], t.equal)
			parent.RemoveChild(leafIndex + 1)
			return true
		}
	}

	return false
}

// handleBranchUnderflow handles the case where a branch node has too few keys
func (t *GenericBPlusTree[K]) handleBranchUnderflow(parent *GenericBranchNode[K], childIndex int) K {
	var zeroKey K // Zero value of K

	child, ok := parent.Children()[childIndex].(*GenericBranchNode[K])
	if !ok {
		return zeroKey
	}

	// Try to borrow from siblings
	if t.tryBorrowFromSiblingBranch(child, parent, childIndex) {
		return zeroKey
	}

	// If borrowing fails, merge with a sibling
	return t.mergeBranchWithSibling(child, parent, childIndex)
}

// tryBorrowFromSiblingBranch tries to borrow a key from a sibling branch
func (t *GenericBPlusTree[K]) tryBorrowFromSiblingBranch(branch *GenericBranchNode[K], parent *GenericBranchNode[K], branchIndex int) bool {
	// Try to borrow from right sibling
	if branchIndex < len(parent.Children())-1 {
		rightSibling, ok := parent.Children()[branchIndex+1].(*GenericBranchNode[K])
		if ok && len(rightSibling.Keys()) > minInternalKeys(t.branchingFactor) {
			separatorKey := parent.Keys()[branchIndex]
			branch.BorrowFromRight(separatorKey, rightSibling, branchIndex, parent)
			return true
		}
	}

	// Try to borrow from left sibling
	if branchIndex > 0 {
		leftSibling, ok := parent.Children()[branchIndex-1].(*GenericBranchNode[K])
		if ok && len(leftSibling.Keys()) > minInternalKeys(t.branchingFactor) {
			separatorKey := parent.Keys()[branchIndex-1]
			branch.BorrowFromLeft(separatorKey, leftSibling, branchIndex, parent)
			return true
		}
	}

	return false
}

// mergeBranchWithSibling merges a branch node with one of its siblings
func (t *GenericBPlusTree[K]) mergeBranchWithSibling(branch *GenericBranchNode[K], parent *GenericBranchNode[K], branchIndex int) K {
	var keyToReturn K // Zero value of K

	// Try to merge with left sibling
	if branchIndex > 0 {
		leftSibling, ok := parent.Children()[branchIndex-1].(*GenericBranchNode[K])
		if ok {
			// Get the separator key from the parent
			separatorKey := parent.Keys()[branchIndex-1]
			// Merge branch into left sibling
			leftSibling.MergeWith(separatorKey, branch)
			// Remove the separator key and the branch from the parent
			parent.DeleteKey(separatorKey, t.equal)
			parent.RemoveChild(branchIndex)
			return keyToReturn
		}
	}

	// Try to merge with right sibling
	if branchIndex < len(parent.Children())-1 {
		rightSibling, ok := parent.Children()[branchIndex+1].(*GenericBranchNode[K])
		if ok {
			// Get the separator key from the parent
			separatorKey := parent.Keys()[branchIndex]
			// Merge right sibling into branch
			branch.MergeWith(separatorKey, rightSibling)
			// Remove the separator key and the right sibling from the parent
			parent.DeleteKey(separatorKey, t.equal)
			parent.RemoveChild(branchIndex + 1)
			// If this was the last key in the parent, we need to return the first key of the merged node
			if len(parent.Keys()) == 0 && len(branch.Keys()) > 0 {
				return branch.Keys()[0]
			}
			return keyToReturn
		}
	}

	return keyToReturn
}

// GetAllKeys returns all keys in the tree
func (t *GenericBPlusTree[K]) GetAllKeys() []K {
	keys := make([]K, 0, t.size)
	t.collectKeys(t.root, &keys)
	return keys
}

// collectKeys collects all keys in the subtree rooted at node
func (t *GenericBPlusTree[K]) collectKeys(node GenericNode[K], keys *[]K) {
	switch n := node.(type) {
	case *GenericLeafNode[K]:
		*keys = append(*keys, n.Keys()...)
	case *GenericBranchNode[K]:
		for _, child := range n.Children() {
			t.collectKeys(child, keys)
		}
	}
}

// RangeQuery returns all keys in the range [start, end]
func (t *GenericBPlusTree[K]) RangeQuery(start, end K) []K {
	result := make([]K, 0)

	// Find the leaf containing the start key
	leaf := t.findLeafNode(t.root, start)
	if leaf == nil {
		return result
	}

	// Traverse the linked list of leaves until we reach the end key
	for leaf != nil {
		for _, key := range leaf.Keys() {
			if (t.less(start, key) || t.equal(start, key)) &&
				(t.less(key, end) || t.equal(key, end)) {
				result = append(result, key)
			}

			if t.less(end, key) {
				return result // We've reached the end
			}
		}

		leaf = leaf.next
	}

	return result
}

// String returns a string representation of the tree
func (t *GenericBPlusTree[K]) String() string {
	return fmt.Sprintf("Tree(size=%d, height=%d, branching=%d)", t.size, t.height, t.branchingFactor)
}
