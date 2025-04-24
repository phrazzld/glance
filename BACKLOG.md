# BACKLOG

## Completed

- ✅ Simplify the progress bar tracking system by removing unnecessary interfaces and factory patterns (April 2025)
  - Radical simplification: completely removed abstraction layer for progress bars
  - Implemented direct progress bar usage in glance.go
  - Added test output suppression mechanism
  - Updated all tests to work with the new implementation

## Current

- improve performance -- make *fast*
- timestamp generated glance files
- remove force option
- audit whole codebase against dev philosophy, identify key things to hit
- refactor aggressively

## Code Review Issues

### High Priority

- Refactor integration tests to use production code directly without duplicating logic
- Revise the testing approach to avoid mocking internal collaborators

### Medium Priority

- Replace `After` checks with `NotEqual` where possible in integration tests
- Remove manual call to `BubbleUpParents` and assertion on `needsRegen` in tests
- Refactor `processDirectories` into smaller functions
- Remove interface-concrete pattern to prevent drift
- Rename test files to clarify their scope and approach

### Low Priority

- Remove deprecated `NewProcessor` function
- Remove redundant comments on trivial types
- Use consistent type naming in test code
- Remove confusing dual construction paths
- Use Go composition patterns where appropriate
- Add GoDoc comments to new public interfaces and methods
- Ensure `.gitignore` correctly reflects tracked files
- Clarify logging strategy in relation to backlog items
- Remove obsolete mock example code
- Add comments explaining ignored errors
