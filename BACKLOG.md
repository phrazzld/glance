# BACKLOG

- improve performance -- make *fast*
- timestamp generated glance files
- remove force option
- audit whole codebase against dev philosophy, identify key things to hit
- refactor aggressively

## Code Review Issues

### High Priority

- Simplify the progress bar tracking system by removing unnecessary interfaces and factory patterns
- Refactor integration tests to use production code directly without duplicating logic
- Revise the testing approach to avoid mocking internal collaborators

### Medium Priority

- Replace `After` checks with `NotEqual` where possible in integration tests
- Remove manual call to `BubbleUpParents` and assertion on `needsRegen` in tests
- Remove interface/factory pattern for progress bar in favor of direct construction
- Refactor `processDirectories` into smaller functions
- Develop clear strategy for suppressing or capturing progress bar output in tests
- Simplify progress bar mocking in tests
- Remove interface-concrete pattern to prevent drift
- Rename test files to clarify their scope and approach

### Low Priority

- Rename `ConcreteProgressBar` to follow idiomatic Go naming
- Remove deprecated `NewProcessor` function
- Remove redundant comments on trivial types
- Use consistent type naming in test code
- Remove confusing dual construction paths
- Trim overly verbose comments on options
- Document security rationale for progress bar output
- Use Go composition patterns where appropriate
- Refactor progress bar option tests for single responsibility
- Add GoDoc comments to new public interfaces and methods
- Ensure `.gitignore` correctly reflects tracked files
- Clarify logging strategy in relation to backlog items
- Remove obsolete mock example code
- Add comments explaining ignored errors
