# Create Filesystem Utilities

## Task Goal
Create filesystem/utils.go with helper functions like latestModTime and shouldRegenerate to centralize filesystem utility functions that are currently in the main package. This will improve code organization and maintainability by moving these functions to the appropriate package.

## Implementation Approach
I'll create a new file `filesystem/utils.go` that will contain the following functions extracted from glance.go:

1. `LatestModTime` - Determine the latest modification time of any file in a directory (recursively)
2. `ShouldRegenerate` - Determine if a GLANCE.md file needs to be regenerated based on modification times
3. `BubbleUpParents` - Mark parent directories for regeneration when a child directory is updated

In addition to simply moving these functions, I'll:

1. Update the functions to use the new IgnoreChain type instead of []*gitignore.GitIgnore
2. Utilize the ShouldIgnorePath function from the ignore.go module for consistent ignore logic
3. Add proper documentation to each function
4. Ensure the functions follow Go best practices (proper error handling, descriptive parameter names)

## Key Reasoning
This approach is best because:

1. **Appropriate Package Organization**: These functions are filesystem-related utilities and should be in the filesystem package, not the main package.

2. **Consistency with New Architecture**: Moving these functions aligns with the project's recent refactoring to organize functionality into dedicated packages.

3. **Code Reusability**: Placing these utilities in a dedicated file makes them more discoverable and reusable by other parts of the application.

4. **Improved Maintainability**: By using the centralized ignore logic from ignore.go, we ensure consistent behavior and make future changes easier.

5. **Clean Dependencies**: This approach maintains a clean dependency flow: main package can depend on filesystem package, but filesystem package should not depend on main.

This implementation approach integrates well with the existing filesystem package structure (scanner.go, reader.go, ignore.go) and follows the established patterns for code organization and documentation.