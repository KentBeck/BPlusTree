#!/usr/bin/env python3

import os
import re
import matplotlib.pyplot as plt
import numpy as np

def parse_result_file(filename):
    """Parse a result file and extract key metrics."""
    with open(filename, 'r') as f:
        content = f.read()
    
    # Extract tree size
    tree_size_match = re.search(r'Keys: (\d+)', content)
    tree_size = int(tree_size_match.group(1)) if tree_size_match else 0
    
    # Extract branching factor
    bf_match = re.search(r'Branching Factor: (\d+)', content)
    branching_factor = int(bf_match.group(1)) if bf_match else 0
    
    # Extract pattern results
    pattern_results = []
    pattern_sections = re.findall(r'Running deletion test with pattern: (\w+).*?Deletion completed in ([\d.]+\w+) \(([\d.]+) keys/sec\)', 
                                 content, re.DOTALL)
    
    for pattern, time_str, keys_per_sec in pattern_sections:
        # Convert time string to seconds
        if 'ms' in time_str:
            time_sec = float(time_str.replace('ms', '')) / 1000
        elif 's' in time_str:
            time_sec = float(time_str.replace('s', ''))
        else:
            time_sec = float(time_str)
        
        pattern_results.append({
            'pattern': pattern,
            'time_sec': time_sec,
            'keys_per_sec': float(keys_per_sec)
        })
    
    return {
        'tree_size': tree_size,
        'branching_factor': branching_factor,
        'pattern_results': pattern_results
    }

def analyze_tree_sizes():
    """Analyze results for different tree sizes."""
    sizes = [10000, 100000, 1000000]
    results = []
    
    for size in sizes:
        filename = f"results/size_{size}.txt"
        if os.path.exists(filename):
            results.append(parse_result_file(filename))
    
    if not results:
        print("No tree size results found.")
        return
    
    # Plot deletion time vs tree size
    plt.figure(figsize=(10, 6))
    
    patterns = set()
    for result in results:
        for pattern_result in result['pattern_results']:
            patterns.add(pattern_result['pattern'])
    
    patterns = sorted(list(patterns))
    colors = plt.cm.tab10(np.linspace(0, 1, len(patterns)))
    
    for i, pattern in enumerate(patterns):
        x = []
        y = []
        for result in results:
            for pattern_result in result['pattern_results']:
                if pattern_result['pattern'] == pattern:
                    x.append(result['tree_size'])
                    y.append(pattern_result['keys_per_sec'])
        plt.plot(x, y, 'o-', label=pattern, color=colors[i])
    
    plt.xscale('log')
    plt.xlabel('Tree Size (number of keys)')
    plt.ylabel('Deletion Speed (keys/second)')
    plt.title('Deletion Performance vs Tree Size')
    plt.legend()
    plt.grid(True)
    plt.savefig('results/tree_size_performance.png')
    plt.close()

def analyze_branching_factors():
    """Analyze results for different branching factors."""
    factors = [16, 64, 256, 1024]
    results = []
    
    for bf in factors:
        filename = f"results/bf_{bf}.txt"
        if os.path.exists(filename):
            results.append(parse_result_file(filename))
    
    if not results:
        print("No branching factor results found.")
        return
    
    # Plot deletion time vs branching factor
    plt.figure(figsize=(10, 6))
    
    patterns = set()
    for result in results:
        for pattern_result in result['pattern_results']:
            patterns.add(pattern_result['pattern'])
    
    patterns = sorted(list(patterns))
    colors = plt.cm.tab10(np.linspace(0, 1, len(patterns)))
    
    for i, pattern in enumerate(patterns):
        x = []
        y = []
        for result in results:
            for pattern_result in result['pattern_results']:
                if pattern_result['pattern'] == pattern:
                    x.append(result['branching_factor'])
                    y.append(pattern_result['keys_per_sec'])
        plt.plot(x, y, 'o-', label=pattern, color=colors[i])
    
    plt.xscale('log')
    plt.xlabel('Branching Factor')
    plt.ylabel('Deletion Speed (keys/second)')
    plt.title('Deletion Performance vs Branching Factor')
    plt.legend()
    plt.grid(True)
    plt.savefig('results/branching_factor_performance.png')
    plt.close()

def analyze_patterns():
    """Analyze performance across different deletion patterns."""
    # Use the largest tree size result
    filename = "results/size_1000000.txt"
    if not os.path.exists(filename):
        print(f"File {filename} not found.")
        return
    
    result = parse_result_file(filename)
    
    # Plot performance by pattern
    plt.figure(figsize=(10, 6))
    
    patterns = []
    speeds = []
    
    for pattern_result in result['pattern_results']:
        patterns.append(pattern_result['pattern'])
        speeds.append(pattern_result['keys_per_sec'])
    
    # Sort by pattern name
    sorted_data = sorted(zip(patterns, speeds))
    patterns = [x[0] for x in sorted_data]
    speeds = [x[1] for x in sorted_data]
    
    plt.bar(patterns, speeds)
    plt.xlabel('Deletion Pattern')
    plt.ylabel('Deletion Speed (keys/second)')
    plt.title(f'Deletion Performance by Pattern (Tree Size: {result["tree_size"]})')
    plt.xticks(rotation=45)
    plt.tight_layout()
    plt.savefig('results/pattern_performance.png')
    plt.close()

def main():
    """Main function to analyze all results."""
    if not os.path.exists('results'):
        print("Results directory not found.")
        return
    
    analyze_tree_sizes()
    analyze_branching_factors()
    analyze_patterns()
    
    print("Analysis complete. Graphs saved to results directory.")

if __name__ == "__main__":
    main()
