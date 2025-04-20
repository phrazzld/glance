# Glance Error Package

This package provides custom error types and error handling utilities for the Glance application.

## Features

- **Custom Error Types**: Specific error types for different categories of errors (FileSystem, API, Config, Validation)
- **Error Context**: Rich error information including codes, severity levels, and helpful suggestions
- **Error Wrapping**: Full support for Go's error wrapping capabilities
- **Sentinel Errors**: Predefined common errors for simpler error checking
- **Type Checking**: Helper functions to check error types

## Usage

### Basic Error Creation

```go
// Create a simple error
err := errors.New("something went wrong")

// Create an error with more context
detailedErr := errors.New("operation failed").
    WithCode("E001").
    WithSeverity(errors.ErrorSeverityCritical).
    WithSuggestion("check system logs")
```

### Error Wrapping

```go
// Wrap a standard error with additional context
fileErr := os.Open("config.txt")
if fileErr != nil {
    return errors.NewFileSystemError("failed to open config file", fileErr).
        WithSuggestion("verify the file exists and is readable")
}
```

### Error Type Checking

```go
// Handle different types of errors
func handleError(err error) {
    switch {
    case errors.IsFileSystemError(err):
        // Handle file system error
    case errors.IsAPIError(err):
        // Handle API error
    case errors.IsConfigError(err):
        // Handle configuration error
    default:
        // Handle other errors
    }
}
```

### Sentinel Errors

```go
// Use sentinel errors for common error conditions
if errors.Is(err, errors.ErrFileNotFound) {
    // Handle file not found error
}
```

## Error Types

- **FileSystemError**: For file and directory access issues
- **APIError**: For LLM API-related errors
- **ConfigError**: For configuration and startup errors
- **ValidationError**: For input validation failures

## Best Practices

1. **Add Context**: Always add context when wrapping errors
2. **Use Sentinel Errors**: For common error conditions, use the provided sentinel errors
3. **Check Error Types**: Use the Is* functions to check error types
4. **Include Suggestions**: When appropriate, include suggestions for how to resolve the error
5. **Set Severity**: Use appropriate severity levels for different errors
