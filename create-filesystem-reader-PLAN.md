# Create Filesystem Reader

## Task Goal
Create a filesystem/reader.go file with functions for reading files and detecting text content. This will move the file reading logic from glance.go into the filesystem package to better organize the code and adhere to the package structure being established.

## Implementation Approach
I'll create a filesystem/reader.go file that contains the following functions extracted from glance.go:
1. `ReadTextFile` - A function to read a file and return its contents as a string with UTF-8 validation
2. `IsTextFile` - The existing function to detect if a file is text-based
3. `GatherLocalFiles` - Modified version of the current function to collect text files in a directory

I'll also add a new utility function:
- `TruncateContent` - For truncating file content to a maximum size

## Key Reasoning
This approach is best because:

1. **Separation of Concerns**: It moves filesystem-specific functionality from the main package to the dedicated filesystem package, improving code organization.

2. **Reusability**: The extracted functions can be reused by other parts of the application and future components without duplicating code.

3. **Testability**: Having these functions in a dedicated package makes them easier to test in isolation.

4. **Maintainability**: The refactoring aligns with the project's direction of splitting functionality into modular packages, making the codebase easier to understand and maintain.

This implementation will follow Go best practices by using proper error handling through returns and maintaining the existing code style patterns seen in the scanner.go file.