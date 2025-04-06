# Centralize Ignore Logic

## Task Goal
Create filesystem/ignore.go with a unified ShouldIgnore function that provides consistent file/directory skipping logic across the application. This will centralize the currently duplicated logic for determining whether a file or directory should be ignored based on .gitignore rules and other criteria.

## Implementation Approach
I'll create a new file `filesystem/ignore.go` that will contain:

1. **ShouldIgnoreFile** - A function to determine if a file should be ignored
2. **ShouldIgnoreDir** - A function to determine if a directory should be ignored
3. **ShouldIgnorePath** - A generic function that handles both files and directories
4. **MatchesGitignore** - A helper function to check if a path matches any gitignore rule in a chain

Additionally, I'll define some constants for commonly ignored patterns:
- Default ignore patterns for hidden files (starting with ".")
- Special file/directory names to ignore (like "node_modules")
- Other common patterns to ignore ("GLANCE.md" files)

The implementation will leverage the existing IgnoreRule and IgnoreChain types from scanner.go, ensuring that the ignore logic is consistent throughout the application.

## Key Reasoning
This approach is best because:

1. **Elimination of Duplication**: Currently, similar ignore-checking logic is duplicated in scanner.go and reader.go, and partially in glance.go. Centralizing this logic in one file reduces code duplication and maintains consistency.

2. **Separation of Concerns**: By extracting ignore logic into its own file, we make the code more modular and focused. Each file in the filesystem package handles a specific aspect of filesystem operations.

3. **Improved Maintainability**: When ignore rules need to be updated or extended, there will be a single place to make changes, reducing the risk of inconsistencies or bugs.

4. **Better Testability**: Having dedicated functions for ignore logic makes it easier to write targeted tests that verify the behavior independently of the scanning or reading functionality.

This implementation aligns with the project's direction of modularizing functionality into separate packages and files, improving code organization and maintainability.