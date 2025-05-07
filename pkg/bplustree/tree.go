package bplustree

// Helper functions for minimum key calculations
func minInternalKeys(branchingFactor int) int {
	return (branchingFactor+1)/2 - 1
}

func minLeafKeys(branchingFactor int) int {
	return (branchingFactor + 1) / 2
}
