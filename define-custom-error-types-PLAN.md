# Define Custom Error Types

## Task Goal
Create custom error types for more specific error handling in the application. This will enable better error classification, improve error reporting, and lay the groundwork for more sophisticated error handling throughout the codebase.

## Implementation Approach
I'll create a new package `errors` with the following components:

1. **Base Error Type:**
   - Define a `GlanceError` interface that extends the standard `error` interface
   - Include methods for error categorization, context, and unwrapping

2. **Specific Error Categories:**
   - `FileSystemError` - For file and directory access issues
   - `ConfigError` - For configuration and startup errors
   - `APIError` - For LLM API-related errors
   - `ValidationError` - For input validation failures

3. **Error Implementation:**
   - Create a base error struct that implements the `GlanceError` interface
   - Implement constructors for each error type
   - Support error wrapping (using `%w` in `fmt.Errorf`)
   - Include helpful metadata like error codes, severity levels, and suggestions

4. **Helper Functions:**
   - `Is<ErrorType>` functions to check error types (e.g., `IsAPIError(err error) bool`)
   - `New<ErrorType>` functions to create errors with appropriate context
   - `Wrap` function to add context to existing errors
   - Error formatting utilities for consistent error messages

5. **Sentinel Errors:**
   - Define common sentinel errors that can be used for equality comparison
   - Group them by category for better organization

## Key Reasoning
This approach is best because:

1. **Improved Error Classification**: Creating distinct error types allows us to distinguish between different categories of errors (e.g., file system errors vs. API errors), making error handling more precise.

2. **Better Error Context**: Custom error types can include additional context like suggested solutions, error codes, and severity levels, making debugging easier.

3. **Consistent Error Formatting**: By centralizing error creation, we ensure consistent error messages throughout the application, improving the user experience.

4. **Error Wrapping Support**: Building on Go's error wrapping capabilities, we can maintain error chains while adding context at each level, preserving the root cause.

5. **Future Extensibility**: This design allows for easy addition of new error types and helper functions as the application evolves.

This implementation follows Go best practices for error handling and aligns with the project's modular architecture, making it a good fit for the codebase.