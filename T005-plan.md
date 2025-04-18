# T005: Refactor filesystem.ListDirsWithIgnores for consolidation - Plan

## Task Analysis
The task involves enhancing the `filesystem.ListDirsWithIgnores` function to handle all BFS (Breadth-First Search) use cases currently covered by both implementations. Currently, there are two separate BFS implementations:

1. In `filesystem/scanner.go`: `ListDirsWithIgnores` - Uses `IgnoreRule` and `IgnoreChain` types
2. In `glance.go`: `listAllDirsWithIgnores` - Uses a slice of `*gitignore.GitIgnore` objects

These implementations have different return types and slightly different implementations, which creates code duplication. The consolidated implementation needs to satisfy all requirements of both current implementations.

## Implementation Plan

1. Enhance `filesystem.ListDirsWithIgnores` to handle all use cases by:
   - Ensuring it properly matches directory patterns in gitignore rules
   - Ensuring the same directory-skipping logic (hidden dirs, node_modules)
   - Using the shared ignore functions from `filesystem/ignore.go`

2. The enhanced function will need to:
   - Maintain backward compatibility with its existing API
   - Support the same functionality as `glance.go:listAllDirsWithIgnores`
   - Improve the code by using the more robust and better-encapsulated ignore functions

## Approach
1. Review both current implementations to understand their differences
2. Modify `filesystem.ListDirsWithIgnores` to use the shared ignore functions
3. Ensure the enhanced function handles all edge cases correctly
4. Add appropriate comments and documentation
5. Update tests if needed

## Specific Changes

1. In `filesystem/scanner.go`:
   - Update `ListDirsWithIgnores` to use `ShouldIgnoreDir` for directory filtering
   - Simplify the ignore matching logic by leveraging existing shared functions
   - Maintain the same return signature to ensure backward compatibility
   - Add comprehensive comments to document the enhanced functionality

Once this task is complete, it will enable T006 (Update glance.go to use consolidated BFS logic) and eventually T007 (Remove duplicate BFS code from glance.go).
