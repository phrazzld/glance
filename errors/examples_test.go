package errors

import (
	"fmt"
	"os"
	stderrors "errors"
)

// ExampleBasicErrorCreation demonstrates creating and using basic errors.
func Example_basicErrorCreation() {
	// Create a simple error
	err := New("something went wrong")
	fmt.Println(err)
	
	// Create an error with more context
	detailedErr := New("operation failed").
		WithCode("E001").
		WithSeverity(ErrorSeverityCritical).
		WithSuggestion("check system logs")
	
	fmt.Println(detailedErr)
	
	// Output:
	// something went wrong
	// [E001] operation failed (CRITICAL) - Suggestion: check system logs
}

// ExampleErrorWrapping demonstrates wrapping errors to add context.
func Example_errorWrapping() {
	// Simulate an error from a standard library function
	baseErr := os.ErrNotExist
	
	// Wrap it with our custom error type
	err := NewFileSystemError("failed to read configuration file", baseErr).
		WithCode("FS-101").
		WithSuggestion("verify the file path is correct")
	
	fmt.Println(err)
	
	// We can still use errors.Is to check the original error
	if stderrors.Is(err, os.ErrNotExist) {
		fmt.Println("The underlying error is os.ErrNotExist")
	}
	
	// Output:
	// [FS-101] failed to read configuration file - Suggestion: verify the file path is correct: file does not exist
	// The underlying error is os.ErrNotExist
}

// ExampleErrorTypeChecking demonstrates checking error types.
func Example_errorTypeChecking() {
	// Create different types of errors
	var err1 error = NewAPIError("API call failed", nil)
	var err2 error = NewFileSystemError("file operation failed", nil)
	var err3 error = NewConfigError("invalid config", nil)
	
	// Check error types
	if IsAPIError(err1) {
		fmt.Println("err1 is an API error")
	}
	
	if IsFileSystemError(err2) {
		fmt.Println("err2 is a file system error")
	}
	
	if !IsAPIError(err3) {
		fmt.Println("err3 is not an API error")
	}
	
	// Output:
	// err1 is an API error
	// err2 is a file system error
	// err3 is not an API error
}

// ExampleSentinelErrors demonstrates using sentinel errors.
func Example_sentinelErrors() {
	// Function that could return various errors
	processFile := func(path string) error {
		if path == "" {
			return ErrInvalidPath
		}
		// In a real application, this would do actual file processing
		return ErrFilePermission
	}
	
	// Try with an invalid path
	err1 := processFile("")
	if stderrors.Is(err1, ErrInvalidPath) {
		fmt.Println("Invalid path error detected")
	}
	
	// Try with a path lacking permissions
	err2 := processFile("/etc/secure")
	if stderrors.Is(err2, ErrFilePermission) {
		fmt.Println("Permission error detected")
	}
	
	// Output:
	// Invalid path error detected
	// Permission error detected
}