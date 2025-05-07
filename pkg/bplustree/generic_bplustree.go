// Package bplustree provides a generic implementation of a B+ tree data structure.
//
// A B+ tree is a self-balancing tree data structure that maintains sorted data
// and allows searches, sequential access, insertions, and deletions in logarithmic time.
// This implementation is generic and can work with any comparable type.
//
// The B+ tree is optimized with a bloom filter for faster lookups of non-existent keys.
package bplustree

import (
	"fmt"
)

// GenericBPlusTree is a B+ tree that works with any comparable type.
// It provides efficient operations for inserting, deleting, and querying keys.
// The tree is self-balancing and maintains its height automatically.
//
// The generic parameter K represents the type of keys stored in the tree.
// The tree requires three functions to work with the keys:
// - less: a function that returns true if a < b
// - equal: a function that returns true if a == b
// - hashFunc: a function that converts a key to a uint64 for bloom filter usage
type GenericBPlusTree[K comparable] struct {
	root            GenericNode[K]       // Root node of the tree
	branchingFactor int                  // Maximum number of children per node
	height          int                  // Current height of the tree
	size            int                  // Number of keys in the tree
	less            func(a, b K) bool    // Function to compare keys (a < b)
	equal           func(a, b K) bool    // Function to check equality of keys (a == b)
	hashFunc        func(K) uint64       // Function to hash keys for bloom filter
	bloomFilter     BloomFilterInterface // Bloom filter for faster lookups
}

// NewGenericBPlusTree creates a new generic B+ tree with the specified parameters.
//
// Parameters:
//   - branchingFactor: The maximum number of children per node. Must be at least 3.
//   - less: A function that returns true if a < b for keys of type K.
//   - equal: A function that returns true if a == b for keys of type K.
//   - hashFunc: A function that converts a key of type K to a uint64 for bloom filter usage.
//
// Returns a new empty B+ tree with a bloom filter enabled for faster lookups.
func NewGenericBPlusTree[K comparable](
	branchingFactor int,
	less func(a, b K) bool,
	equal func(a, b K) bool,
	hashFunc func(K) uint64,
) *GenericBPlusTree[K] {
	if branchingFactor < 3 {
		branchingFactor = 3 // Minimum branching factor
	}

	// Create a Bloom filter with reasonable default parameters
	// Initial size is set for 1000 elements with 1% false positive rate
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

// NewGenericBPlusTreeWithoutBloom creates a new generic B+ tree without a bloom filter.
// This can be more efficient for small trees or when memory usage is a concern.
//
// Parameters:
//   - branchingFactor: The maximum number of children per node. Must be at least 3.
//   - less: A function that returns true if a < b for keys of type K.
//   - equal: A function that returns true if a == b for keys of type K.
//   - hashFunc: A function that converts a key of type K to a uint64 (not used without bloom filter).
//
// Returns a new empty B+ tree with no bloom filter.
func NewGenericBPlusTreeWithoutBloom[K comparable](
	branchingFactor int,
	less func(a, b K) bool,
	equal func(a, b K) bool,
	hashFunc func(K) uint64,
) *GenericBPlusTree[K] {
	if branchingFactor < 3 {
		branchingFactor = 3 // Minimum branching factor
	}

	return &GenericBPlusTree[K]{
		root:            NewGenericLeafNode[K](),
		branchingFactor: branchingFactor,
		height:          1,
		size:            0,
		less:            less,
		equal:           equal,
		hashFunc:        hashFunc,
		bloomFilter:     NewNullBloomFilter(), // Use null bloom filter (always returns "maybe")
	}
}

// Size returns the number of keys in the tree.
// Time complexity: O(1)
func (t *GenericBPlusTree[K]) Size() int {
	return t.size
}

// Height returns the height of the tree.
// The height is the number of levels in the tree, including the leaf level.
// A tree with just a root leaf node has a height of 1.
// Time complexity: O(1)
func (t *GenericBPlusTree[K]) Height() int {
	return t.height
}

// IsEmpty returns true if the tree contains no keys.
// Time complexity: O(1)
func (t *GenericBPlusTree[K]) IsEmpty() bool {
	return t.size == 0
}

// BranchingFactor returns the maximum number of children per node.
// Time complexity: O(1)
func (t *GenericBPlusTree[K]) BranchingFactor() int {
	return t.branchingFactor
}

// Insert inserts a key into the tree.
// Returns true if the key was inserted, false if it already existed.
// Time complexity: O(log n) where n is the number of keys in the tree.
func (t *GenericBPlusTree[K]) Insert(key K) bool {
	// If the root is full, split it before inserting
	if t.root.IsFull(t.branchingFactor) {
		t.splitRoot()
	}

	// Insert the key into the tree
	inserted := t.insertNonFull(t.root, key)

	// Update size and bloom filter if the key was inserted
	if inserted {
		t.size++
		t.updateBloomFilter(key)
	}

	return inserted
}

// splitRoot handles splitting the root when it's full.
// This increases the height of the tree by 1.
// Time complexity: O(B) where B is the branching factor.
func (t *GenericBPlusTree[K]) splitRoot() {
	// Save the old root
	oldRoot := t.root

	// Create a new root as a branch node
	t.root = NewGenericBranchNode[K]()
	newRoot := t.root.(*GenericBranchNode[K])

	// Make the old root the first child of the new root
	newRoot.SetChild(0, oldRoot)

	// Split the old root (now the first child of the new root)
	t.splitChild(newRoot, 0)

	// Increment the height of the tree
	// This is done unconditionally because splitting the root always increases the height
	t.height++
}

// updateBloomFilter adds a key to the bloom filter if it's valid.
// Time complexity: O(k) where k is the number of hash functions in the bloom filter.
func (t *GenericBPlusTree[K]) updateBloomFilter(key K) {
	// Hash the key
	hash := t.hashFunc(key)

	// Add the hash to the bloom filter if it's valid
	if t.bloomFilter.IsValid() {
		t.bloomFilter.Add(hash)
	} else {
		// If the bloom filter is invalid, we could recompute it here,
		// but that would be expensive. Instead, we leave it invalid
		// until Contains is called, which will recompute it if needed.
	}
}

// insertNonFull inserts a key into a non-full node.
// Returns true if the key was inserted, false if it already existed.
// Time complexity: O(log n) where n is the number of keys in the tree.
func (t *GenericBPlusTree[K]) insertNonFull(node GenericNode[K], key K) bool {
	switch n := node.(type) {
	case *GenericLeafNode[K]:
		// If we've reached a leaf node, insert the key
		return n.InsertKey(key, t.less)

	case *GenericBranchNode[K]:
		// Find the child that should contain the key
		childIndex := n.FindChildIndex(key, t.less)

		// Safety check to avoid index out of range
		if childIndex >= len(n.Children()) {
			return false
		}

		child := n.Children()[childIndex]

		// If the child is full, split it before inserting
		if child.IsFull(t.branchingFactor) {
			t.splitChild(n, childIndex)

			// After splitting, determine which child to go to
			// If the key is greater than or equal to the new separator key,
			// we need to go to the right child (childIndex + 1)
			if childIndex < len(n.Keys()) && (t.less(n.Keys()[childIndex], key) || t.equal(n.Keys()[childIndex], key)) {
				childIndex++
			}

			// Safety check again after potential increment
			if childIndex >= len(n.Children()) {
				return false
			}

			// Get the new child
			child = n.Children()[childIndex]
		}

		// Recursively insert into the child
		return t.insertNonFull(child, key)
	}

	// This should never happen if the tree is properly structured
	return false
}

// splitChild splits a full child of a branch node.
// This is a key operation in maintaining the B+ tree property.
// Time complexity: O(B) where B is the branching factor.
func (t *GenericBPlusTree[K]) splitChild(parent *GenericBranchNode[K], childIndex int) {
	// Get the child to split
	child := parent.Children()[childIndex]

	switch c := child.(type) {
	case *GenericBranchNode[K]:
		// Split branch node (internal node)

		// Create a new branch node for the right half
		newChildImpl := NewGenericBranchNode[K]()

		// Calculate the middle index
		midIndex := t.branchingFactor/2 - 1

		// Get the middle key that will move up to the parent
		midKey := c.Keys()[midIndex]

		// Move keys and children to the new node (right half)
		newChildImpl.keys = append(newChildImpl.keys, c.Keys()[midIndex+1:]...)
		newChildImpl.children = append(newChildImpl.children, c.Children()[midIndex+1:]...)

		// Update the original child (left half)
		c.keys = c.Keys()[:midIndex]
		c.children = c.Children()[:midIndex+1]

		// Insert the new child into the parent
		parent.InsertKeyWithChild(midKey, newChildImpl, t.less)

	case *GenericLeafNode[K]:
		// Split leaf node

		// Create a new leaf node for the right half
		newLeafImpl := NewGenericLeafNode[K]()

		// Calculate the middle index
		// For leaf nodes, we include the middle key in the right node
		midIndex := t.branchingFactor / 2

		// Move keys to the new leaf (right half)
		newLeafImpl.keys = append(newLeafImpl.keys, c.Keys()[midIndex:]...)

		// Update the original leaf (left half)
		c.keys = c.Keys()[:midIndex]

		// Update the linked list of leaves for range queries
		newLeafImpl.next = c.next
		c.next = newLeafImpl

		// Insert the new leaf into the parent
		// Use the first key of the new leaf as the separator key
		if len(newLeafImpl.Keys()) > 0 {
			parent.InsertKeyWithChild(newLeafImpl.Keys()[0], newLeafImpl, t.less)
		} else {
			// This should not happen in a properly structured tree
			// But handle it gracefully just in case
			var zeroKey K
			parent.InsertKeyWithChild(zeroKey, newLeafImpl, t.less)
		}
	}

	// Note: We don't update the height here.
	// Height is only updated in splitRoot, which is the only place
	// where the height of the tree actually increases.
}

// Contains returns true if the tree contains the key.
// This method uses a bloom filter for faster lookups of non-existent keys.
// Time complexity: O(log n) where n is the number of keys in the tree.
// In the case of non-existent keys that can be filtered by the bloom filter,
// the time complexity is O(k) where k is the number of hash functions.
func (t *GenericBPlusTree[K]) Contains(key K) bool {
	// Special case for empty tree
	if t.size == 0 {
		return false
	}

	// Ensure bloom filter is valid
	if !t.bloomFilter.IsValid() {
		t.recomputeBloomFilter()
	}

	// Hash the key for bloom filter lookup
	hash := t.hashFunc(key)

	// Early return if bloom filter says key is definitely not present
	// This is a key optimization for lookups of non-existent keys
	if !t.bloomFilter.Contains(hash) {
		return false
	}

	// Check the tree since bloom filter says key might be present
	// (bloom filters can have false positives but not false negatives)
	return t.findLeaf(t.root, key)
}

// recomputeBloomFilter recomputes the Bloom filter from all keys in the tree.
// This is called when the bloom filter is invalid and needs to be rebuilt.
// Time complexity: O(n) where n is the number of keys in the tree.
func (t *GenericBPlusTree[K]) recomputeBloomFilter() {
	// Clear the Bloom filter
	t.bloomFilter.Clear()

	// Add all keys to the Bloom filter
	t.addKeysToBloomFilter(t.root)

	// Mark the Bloom filter as valid
	t.bloomFilter.SetValid()
}

// addKeysToBloomFilter adds all keys in the subtree rooted at node to the Bloom filter.
// Time complexity: O(n) where n is the number of keys in the subtree.
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

// ResizeBloomFilter resizes the bloom filter with new parameters.
// This can be useful when the tree has grown significantly and the
// current bloom filter parameters are no longer optimal.
// Time complexity: O(n) where n is the number of keys in the tree.
func (t *GenericBPlusTree[K]) ResizeBloomFilter(expectedElements int, falsePositiveRate float64) {
	// Calculate optimal bloom filter parameters
	size, hashFunctions := OptimalBloomFilterSize(expectedElements, falsePositiveRate)

	// Create a new bloom filter with the new parameters
	t.bloomFilter = NewBloomFilter(size, hashFunctions)

	// Recompute the bloom filter with all keys in the tree
	t.recomputeBloomFilter()
}

// findLeaf finds the leaf node that should contain the key and checks if it's present.
// Time complexity: O(log n) where n is the number of keys in the tree.
func (t *GenericBPlusTree[K]) findLeaf(node GenericNode[K], key K) bool {
	switch n := node.(type) {
	case *GenericLeafNode[K]:
		// We've reached a leaf node, check if it contains the key
		for _, k := range n.Keys() {
			if t.equal(k, key) {
				return true
			}
		}
		return false

	case *GenericBranchNode[K]:
		// Find the child that should contain the key
		childIndex := n.FindChildIndex(key, t.less)

		// Safety check to avoid index out of range
		if childIndex >= len(n.Children()) {
			return false
		}

		// Recursively search in the appropriate child
		return t.findLeaf(n.Children()[childIndex], key)
	}

	// This should never happen if the tree is properly structured
	return false
}

// findLeafNode finds and returns the leaf node that should contain the key.
// This is used for operations that need to modify the leaf, like range queries.
// Time complexity: O(log n) where n is the number of keys in the tree.
func (t *GenericBPlusTree[K]) findLeafNode(node GenericNode[K], key K) *GenericLeafNode[K] {
	switch n := node.(type) {
	case *GenericLeafNode[K]:
		// We've reached a leaf node, return it
		return n

	case *GenericBranchNode[K]:
		// Find the child that should contain the key
		childIndex := n.FindChildIndex(key, t.less)

		// Safety check to avoid index out of range
		if childIndex >= len(n.Children()) {
			return nil
		}

		// Recursively search in the appropriate child
		return t.findLeafNode(n.Children()[childIndex], key)
	}

	// This should never happen if the tree is properly structured
	return nil
}

// Delete removes a key from the tree.
// Returns true if the key was deleted, false if it didn't exist.
// Time complexity: O(log n) where n is the number of keys in the tree.
func (t *GenericBPlusTree[K]) Delete(key K) bool {
	// Special case for empty tree
	if t.size == 0 {
		return false
	}

	// First, check if the key exists using the bloom filter
	// This is an optimization to avoid the deletion process for non-existent keys
	if t.bloomFilter.IsValid() {
		hash := t.hashFunc(key)
		if !t.bloomFilter.Contains(hash) {
			// If the bloom filter says the key is definitely not present, return false
			return false
		}
	}

	// Delete the key and balance the tree if necessary
	deleted, _ := t.deleteAndBalance(t.root, nil, -1, key)

	if deleted {
		// Update tree state after successful deletion
		t.decrementSize()
		t.handleRootUnderflow()
		t.invalidateBloomFilter()
	}

	return deleted
}

// decrementSize decrements the size counter if it's greater than 0.
// Time complexity: O(1)
func (t *GenericBPlusTree[K]) decrementSize() {
	if t.size > 0 {
		t.size--
	}
}

// handleRootUnderflow handles the case where the root has no keys.
// This happens when the last key is deleted from the root or when
// all keys in the root are moved to its children during balancing.
// Time complexity: O(1)
func (t *GenericBPlusTree[K]) handleRootUnderflow() {
	if t.isEmptyInternalRoot() {
		t.promoteOnlyChild()
	}
}

// isEmptyInternalRoot returns true if the root is an internal node with no keys.
// Time complexity: O(1)
func (t *GenericBPlusTree[K]) isEmptyInternalRoot() bool {
	if branch, ok := t.root.(*GenericBranchNode[K]); ok {
		return len(branch.Keys()) == 0 && len(branch.Children()) > 0
	}
	return false
}

// promoteOnlyChild makes the only child of the root the new root.
// This decreases the height of the tree by 1.
// Time complexity: O(1)
func (t *GenericBPlusTree[K]) promoteOnlyChild() {
	if branch, ok := t.root.(*GenericBranchNode[K]); ok {
		if len(branch.Children()) > 0 {
			t.root = branch.Children()[0]
			t.height--
		}
	}
}

// invalidateBloomFilter clears the bloom filter.
// This is called after a key is deleted, as the bloom filter
// cannot efficiently remove elements.
// Time complexity: O(1)
func (t *GenericBPlusTree[K]) invalidateBloomFilter() {
	t.bloomFilter.Clear()
}

// deleteAndBalance removes a key from a node and balances the tree if necessary.
// This is the core deletion algorithm for the B+ tree.
//
// Parameters:
// - node: The current node being processed
// - parent: The parent of the current node (nil for root)
// - parentChildIndex: The index of the current node in its parent's children array (-1 for root)
// - key: The key to delete
//
// Returns:
// - deleted: true if the key was deleted
// - keyToReplaceInParent: a key that needs to be replaced in the parent (for internal nodes)
//
// Time complexity: O(log n) where n is the number of keys in the tree.
func (t *GenericBPlusTree[K]) deleteAndBalance(node GenericNode[K], parent *GenericBranchNode[K], parentChildIndex int, key K) (bool, K) {
	var zeroKey K // Zero value of K

	switch n := node.(type) {
	case *GenericLeafNode[K]:
		// Case 1: Leaf node

		// Try to delete the key from the leaf
		// First, check if the key exists in this leaf
		keyExists := false
		for _, k := range n.Keys() {
			if t.equal(k, key) {
				keyExists = true
				break
			}
		}

		if !keyExists {
			return false, zeroKey // Key not found
		}

		// Delete the key
		if !n.DeleteKey(key, t.equal) {
			return false, zeroKey // Key not found (should not happen)
		}

		// If this is the root or it doesn't underflow, we're done
		if parent == nil || !n.IsUnderflow(t.branchingFactor) {
			return true, zeroKey
		}

		// Handle underflow by borrowing or merging
		return t.handleLeafUnderflow(n, parent, parentChildIndex), zeroKey

	case *GenericBranchNode[K]:
		// Case 2: Branch node (internal node)

		// Check if the key is in this node
		keyIndex := n.FindKey(key, t.equal)
		if keyIndex != -1 {
			// The key is in this internal node, so we need to find a replacement
			// Get the rightmost key from the left subtree (predecessor)
			if keyIndex < len(n.Children()) {
				leftChild := n.Children()[keyIndex]
				replacementKey, success := t.findAndRemoveMax(leftChild, n, keyIndex)
				if success {
					// Replace the key in this node
					n.keys[keyIndex] = replacementKey
					return true, zeroKey
				}
			}
			return false, zeroKey
		}

		// The key is not in this node, so we need to find the child that should contain it
		childIndex := n.FindChildIndex(key, t.less)

		// Safety check to avoid index out of range
		if childIndex >= len(n.Children()) {
			return false, zeroKey
		}

		// Recursively delete from the child
		deleted, keyToReplace := t.deleteAndBalance(n.Children()[childIndex], n, childIndex, key)
		if !deleted {
			return false, zeroKey // Key not found in the subtree
		}

		// If we need to replace a key in this node
		// This happens when a key in an internal node is replaced during deletion
		if !t.equal(keyToReplace, zeroKey) {
			if childIndex > 0 {
				n.keys[childIndex-1] = keyToReplace
			}
		}

		// Check if the child underflowed and needs rebalancing
		if childIndex < len(n.Children()) {
			child := n.Children()[childIndex]
			if child.IsUnderflow(t.branchingFactor) {
				return true, t.handleBranchUnderflow(n, childIndex)
			}
		}

		return true, zeroKey
	}

	// This should never happen if the tree is properly structured
	return false, zeroKey
}

// findAndRemoveMax finds and removes the maximum key in the subtree rooted at node.
// This is used during deletion when a key in an internal node needs to be replaced.
//
// Parameters:
// - node: The current node being processed
// - parent: The parent of the current node
// - parentChildIndex: The index of the current node in its parent's children array
//
// Returns:
// - The maximum key in the subtree
// - true if a key was found and removed, false otherwise
//
// Time complexity: O(log n) where n is the number of keys in the subtree.
func (t *GenericBPlusTree[K]) findAndRemoveMax(node GenericNode[K], parent *GenericBranchNode[K], parentChildIndex int) (K, bool) {
	var zeroKey K // Zero value of K

	switch n := node.(type) {
	case *GenericLeafNode[K]:
		// Case 1: Leaf node - the maximum key is the last key in the leaf

		// Check if the leaf has any keys
		if len(n.Keys()) == 0 {
			return zeroKey, false
		}

		// Get the maximum key (last key in the leaf)
		maxKeyIndex := len(n.Keys()) - 1
		maxKey := n.Keys()[maxKeyIndex]

		// Remove the maximum key
		n.keys = n.Keys()[:maxKeyIndex]

		// Handle underflow if necessary
		if parent != nil && n.IsUnderflow(t.branchingFactor) {
			t.handleLeafUnderflow(n, parent, parentChildIndex)
		}

		return maxKey, true

	case *GenericBranchNode[K]:
		// Case 2: Branch node - the maximum key is in the rightmost subtree

		// Recursively find and remove the maximum key from the rightmost child
		childIndex := len(n.Children()) - 1

		// Safety check to avoid index out of range
		if childIndex < 0 {
			return zeroKey, false
		}

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

	// This should never happen if the tree is properly structured
	return zeroKey, false
}

// handleLeafUnderflow handles the case where a leaf node has too few keys.
// This is a key part of maintaining the B+ tree property during deletion.
//
// Parameters:
// - leaf: The leaf node that has underflowed
// - parent: The parent of the leaf node
// - leafIndex: The index of the leaf in its parent's children array
//
// Returns:
// - true if the underflow was handled successfully
//
// Time complexity: O(B) where B is the branching factor.
func (t *GenericBPlusTree[K]) handleLeafUnderflow(leaf *GenericLeafNode[K], parent *GenericBranchNode[K], leafIndex int) bool {
	// First try to borrow keys from siblings
	if t.tryBorrowFromSiblingLeaf(leaf, parent, leafIndex) {
		return true
	}

	// If borrowing fails, merge with a sibling
	return t.mergeLeafWithSibling(leaf, parent, leafIndex)
}

// tryBorrowFromSiblingLeaf tries to borrow a key from a sibling leaf.
// This is called when a leaf node has too few keys after deletion.
//
// Parameters:
// - leaf: The leaf node that needs to borrow
// - parent: The parent of the leaf node
// - leafIndex: The index of the leaf in its parent's children array
//
// Returns:
// - true if borrowing was successful
//
// Time complexity: O(B) where B is the branching factor.
func (t *GenericBPlusTree[K]) tryBorrowFromSiblingLeaf(leaf *GenericLeafNode[K], parent *GenericBranchNode[K], leafIndex int) bool {
	// Try to borrow from right sibling first (if it exists)
	if leafIndex < len(parent.Children())-1 {
		rightSibling, ok := parent.Children()[leafIndex+1].(*GenericLeafNode[K])
		if ok && len(rightSibling.Keys()) > minLeafKeys(t.branchingFactor) {
			// Right sibling has enough keys to spare one
			leaf.BorrowFromRight(rightSibling, leafIndex, parent)
			return true
		}
	}

	// If borrowing from right failed, try to borrow from left sibling
	if leafIndex > 0 {
		leftSibling, ok := parent.Children()[leafIndex-1].(*GenericLeafNode[K])
		if ok && len(leftSibling.Keys()) > minLeafKeys(t.branchingFactor) {
			// Left sibling has enough keys to spare one
			leaf.BorrowFromLeft(leftSibling, leafIndex, parent)
			return true
		}
	}

	// Borrowing failed
	return false
}

// mergeLeafWithSibling merges a leaf node with one of its siblings.
// This is called when a leaf node has too few keys and borrowing failed.
//
// Parameters:
// - leaf: The leaf node to merge
// - parent: The parent of the leaf node
// - leafIndex: The index of the leaf in its parent's children array
//
// Returns:
// - true if merging was successful
//
// Time complexity: O(B) where B is the branching factor.
func (t *GenericBPlusTree[K]) mergeLeafWithSibling(leaf *GenericLeafNode[K], parent *GenericBranchNode[K], leafIndex int) bool {
	// Try to merge with left sibling first (if it exists)
	if leafIndex > 0 {
		leftSibling, ok := parent.Children()[leafIndex-1].(*GenericLeafNode[K])
		if ok {
			// Merge leaf into left sibling
			leftSibling.MergeWith(leaf)

			// Update the linked list of leaves
			// (leftSibling.next is already set to leaf.next by MergeWith)

			// Remove the separator key and the leaf from the parent
			parent.DeleteKey(parent.Keys()[leafIndex-1], t.equal)
			parent.RemoveChild(leafIndex)
			return true
		}
	}

	// If merging with left failed, try to merge with right sibling
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

	// Merging failed (this should not happen in a properly structured tree)
	return false
}

// handleBranchUnderflow handles the case where a branch node has too few keys.
// This is a key part of maintaining the B+ tree property during deletion.
//
// Parameters:
// - parent: The parent of the branch node that has underflowed
// - childIndex: The index of the branch in its parent's children array
//
// Returns:
// - A key that needs to be replaced in the parent, or the zero value if no replacement is needed
//
// Time complexity: O(B) where B is the branching factor.
func (t *GenericBPlusTree[K]) handleBranchUnderflow(parent *GenericBranchNode[K], childIndex int) K {
	var zeroKey K // Zero value of K

	// Ensure the child is a branch node
	child, ok := parent.Children()[childIndex].(*GenericBranchNode[K])
	if !ok {
		return zeroKey
	}

	// First try to borrow keys from siblings
	if t.tryBorrowFromSiblingBranch(child, parent, childIndex) {
		return zeroKey
	}

	// If borrowing fails, merge with a sibling
	return t.mergeBranchWithSibling(child, parent, childIndex)
}

// tryBorrowFromSiblingBranch tries to borrow a key from a sibling branch.
// This is called when a branch node has too few keys after deletion.
//
// Parameters:
// - branch: The branch node that needs to borrow
// - parent: The parent of the branch node
// - branchIndex: The index of the branch in its parent's children array
//
// Returns:
// - true if borrowing was successful
//
// Time complexity: O(B) where B is the branching factor.
func (t *GenericBPlusTree[K]) tryBorrowFromSiblingBranch(branch *GenericBranchNode[K], parent *GenericBranchNode[K], branchIndex int) bool {
	// Try to borrow from right sibling first (if it exists)
	if branchIndex < len(parent.Children())-1 {
		rightSibling, ok := parent.Children()[branchIndex+1].(*GenericBranchNode[K])
		if ok && len(rightSibling.Keys()) > minInternalKeys(t.branchingFactor) {
			// Right sibling has enough keys to spare one
			separatorKey := parent.Keys()[branchIndex]
			branch.BorrowFromRight(separatorKey, rightSibling, branchIndex, parent)
			return true
		}
	}

	// If borrowing from right failed, try to borrow from left sibling
	if branchIndex > 0 {
		leftSibling, ok := parent.Children()[branchIndex-1].(*GenericBranchNode[K])
		if ok && len(leftSibling.Keys()) > minInternalKeys(t.branchingFactor) {
			// Left sibling has enough keys to spare one
			separatorKey := parent.Keys()[branchIndex-1]
			branch.BorrowFromLeft(separatorKey, leftSibling, branchIndex, parent)
			return true
		}
	}

	// Borrowing failed
	return false
}

// mergeBranchWithSibling merges a branch node with one of its siblings.
// This is called when a branch node has too few keys and borrowing failed.
//
// Parameters:
// - branch: The branch node to merge
// - parent: The parent of the branch node
// - branchIndex: The index of the branch in its parent's children array
//
// Returns:
// - A key that needs to be replaced in the parent, or the zero value if no replacement is needed
//
// Time complexity: O(B) where B is the branching factor.
func (t *GenericBPlusTree[K]) mergeBranchWithSibling(branch *GenericBranchNode[K], parent *GenericBranchNode[K], branchIndex int) K {
	var keyToReturn K // Zero value of K

	// Try to merge with left sibling first (if it exists)
	if branchIndex > 0 {
		leftSibling, ok := parent.Children()[branchIndex-1].(*GenericBranchNode[K])
		if ok {
			// Get the separator key from the parent
			separatorKey := parent.Keys()[branchIndex-1]

			// Merge branch into left sibling
			// The separator key from the parent goes into the left sibling
			leftSibling.MergeWith(separatorKey, branch)

			// Remove the separator key and the branch from the parent
			parent.DeleteKey(separatorKey, t.equal)
			parent.RemoveChild(branchIndex)

			return keyToReturn
		}
	}

	// If merging with left failed, try to merge with right sibling
	if branchIndex < len(parent.Children())-1 {
		rightSibling, ok := parent.Children()[branchIndex+1].(*GenericBranchNode[K])
		if ok {
			// Get the separator key from the parent
			separatorKey := parent.Keys()[branchIndex]

			// Merge right sibling into branch
			// The separator key from the parent goes into the branch
			branch.MergeWith(separatorKey, rightSibling)

			// Remove the separator key and the right sibling from the parent
			parent.DeleteKey(separatorKey, t.equal)
			parent.RemoveChild(branchIndex + 1)

			// If this was the last key in the parent, we need to return the first key of the merged node
			// This key will be used to replace a key in a higher level of the tree
			if len(parent.Keys()) == 0 && len(branch.Keys()) > 0 {
				return branch.Keys()[0]
			}

			return keyToReturn
		}
	}

	// Merging failed (this should not happen in a properly structured tree)
	return keyToReturn
}

// GetAllKeys returns all keys in the tree as an unsorted slice.
// Time complexity: O(n) where n is the number of keys in the tree.
func (t *GenericBPlusTree[K]) GetAllKeys() []K {
	// Pre-allocate the slice with the known size for efficiency
	keys := make([]K, 0, t.size)

	// Collect all keys from the tree
	t.collectKeys(t.root, &keys)

	return keys
}

// collectKeys collects all keys in the subtree rooted at node.
// Time complexity: O(n) where n is the number of keys in the subtree.
func (t *GenericBPlusTree[K]) collectKeys(node GenericNode[K], keys *[]K) {
	switch n := node.(type) {
	case *GenericLeafNode[K]:
		// For leaf nodes, add all keys to the result
		*keys = append(*keys, n.Keys()...)

	case *GenericBranchNode[K]:
		// For branch nodes, recursively collect keys from all children
		for _, child := range n.Children() {
			t.collectKeys(child, keys)
		}
	}
}

// RangeQuery returns all keys in the range [start, end], inclusive.
// The keys are returned in sorted order.
// Time complexity: O(log n + k) where n is the number of keys in the tree
// and k is the number of keys in the range.
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
			// Check if the key is in the range [start, end]
			inRange := (t.less(start, key) || t.equal(start, key)) &&
				(t.less(key, end) || t.equal(key, end))

			if inRange {
				result = append(result, key)
			}

			// If we've passed the end key, we're done
			if t.less(end, key) {
				return result
			}
		}

		// Move to the next leaf in the linked list
		leaf = leaf.next
	}

	return result
}

// Clear removes all keys from the tree.
// Time complexity: O(1)
func (t *GenericBPlusTree[K]) Clear() {
	// Create a new empty leaf node as the root
	t.root = NewGenericLeafNode[K]()

	// Reset tree properties
	t.height = 1
	t.size = 0

	// Clear the bloom filter
	t.bloomFilter.Clear()
}

// String returns a string representation of the tree.
// Time complexity: O(1)
func (t *GenericBPlusTree[K]) String() string {
	return fmt.Sprintf("GenericBPlusTree(size=%d, height=%d, branching=%d)",
		t.size, t.height, t.branchingFactor)
}

// CountKeys counts the actual number of keys in the tree by traversing it.
// This is useful for debugging and verification.
// Time complexity: O(n) where n is the number of keys in the tree.
func (t *GenericBPlusTree[K]) CountKeys() int {
	count := 0
	t.traverseTree(t.root, func(key K) {
		count++
	})
	return count
}

// traverseTree traverses the tree in-order and calls the visitor function for each key.
// Time complexity: O(n) where n is the number of keys in the tree.
func (t *GenericBPlusTree[K]) traverseTree(node GenericNode[K], visitor func(K)) {
	switch n := node.(type) {
	case *GenericLeafNode[K]:
		for _, key := range n.Keys() {
			visitor(key)
		}
	case *GenericBranchNode[K]:
		for i, child := range n.Children() {
			t.traverseTree(child, visitor)
			if i < len(n.Keys()) {
				// Skip the separator keys in branch nodes
				// They are duplicated in the leaf nodes
			}
		}
	}
}

// ForceDeleteKeys forcibly deletes keys from the tree.
// This is a utility method for testing and debugging.
// It returns the number of keys that were actually deleted.
// Time complexity: O(n*log(n)) where n is the number of keys to delete.
func (t *GenericBPlusTree[K]) ForceDeleteKeys(keys []K) int {
	// First, collect all keys in the tree
	treeKeys := t.GetAllKeys()

	// Create a map of keys to delete for faster lookup
	keysToDelete := make(map[K]bool)
	for _, key := range keys {
		keysToDelete[key] = true
	}

	// Delete all keys that are in the tree and in the keys to delete
	count := 0
	for _, key := range treeKeys {
		if keysToDelete[key] {
			// Use a direct approach to delete the key
			leaf := t.findLeafNode(t.root, key)
			if leaf != nil && leaf.DeleteKey(key, t.equal) {
				t.decrementSize()
				count++
			}
		}
	}

	// Invalidate the bloom filter
	t.invalidateBloomFilter()

	return count
}

// ResetSize resets the size counter to match the actual number of keys in the tree.
// This is useful for debugging and verification.
// Time complexity: O(n) where n is the number of keys in the tree.
func (t *GenericBPlusTree[K]) ResetSize() {
	t.size = t.CountKeys()
}
