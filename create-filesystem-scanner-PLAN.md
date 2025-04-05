# Create Filesystem Scanner

## Goal
Create a dedicated filesystem scanner package that handles directory traversal and .gitignore handling, extracting this functionality from the main.go file to improve modularity and maintainability.

## Implementation Approach
1. Create a `filesystem` package with a `scanner.go` file
2. Move the directory traversal functionality (BFS scanning) from main.go to this new package
3. Design a clean API with exported functions and proper error handling
4. Maintain compatibility with the existing code while preparing for future refactorings

The core functions to implement include:
- `ListDirsWithIgnores` - Exported function that performs BFS traversal and returns directories with their gitignore chains
- Supporting internal functionality like handling gitignore files, path matching, etc.

## Reasoning
I considered three potential approaches:

1. **Direct extraction** - Simply move the exact code from main.go to the new package with minimal changes
   - Pros: Simple, low risk of introducing bugs
   - Cons: Misses opportunity to improve the design, may not be optimal for future use

2. **Complete redesign** - Redesign the entire filesystem scanning approach from scratch
   - Pros: Could potentially be more elegant and efficient
   - Cons: High risk, requires substantial testing, might introduce compatibility issues

3. **Balanced approach** (chosen) - Extract the core functionality while making targeted improvements to the API
   - Pros: Maintains compatibility while cleaning up the interface and preparing for future work
   - Cons: Requires careful consideration of what to change vs. preserve

I've selected the balanced approach because:
- It minimizes risk while still improving the codebase
- It aligns with the incremental refactoring strategy outlined in the plan
- It provides a clean abstraction that can be further refined in subsequent tasks
- The existing BFS-based directory traversal approach is sound and doesn't need a complete redesign