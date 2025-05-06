#!/bin/bash

# Create directory for results
mkdir -p results

# Run stress tests with different configurations
echo "Running stress tests..."

# Test with different tree sizes
echo "Testing different tree sizes..."
for size in 10000 100000 1000000; do
    echo "Running test with $size keys..."
    go run deletion_stress.go -keys=$size -patterns=true -verify=true > "results/size_${size}.txt"
done

# Test with different branching factors
echo "Testing different branching factors..."
for bf in 16 64 256 1024; do
    echo "Running test with branching factor $bf..."
    go run deletion_stress.go -keys=100000 -bf=$bf -patterns=true -verify=true > "results/bf_${bf}.txt"
done

# Run a CPU profile test
echo "Running CPU profile test..."
go run deletion_stress.go -keys=100000 -cpuprofile=results/cpu.prof -patterns=false -verify=true > "results/cpu_profile.txt"

# Run a memory profile test
echo "Running memory profile test..."
go run deletion_stress.go -keys=100000 -memprofile=results/mem.prof -patterns=false -verify=true > "results/mem_profile.txt"

echo "All tests completed. Results are in the 'results' directory."
