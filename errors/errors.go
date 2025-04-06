// Package errors provides custom error types and error handling utilities
// for the glance application.
package errors

import (
	"errors"
	"fmt"
)

// -----------------------------------------------------------------------------
// Error Severity Levels
// -----------------------------------------------------------------------------

// ErrorSeverity represents the severity level of an error.
type ErrorSeverity int

const (
	// ErrorSeverityNormal indicates a standard error.
	ErrorSeverityNormal ErrorSeverity = iota
	
	// ErrorSeverityWarning indicates a warning condition.
	ErrorSeverityWarning
	
	// ErrorSeverityCritical indicates a critical error.
	ErrorSeverityCritical
)

// String returns a string representation of the severity level.
func (s ErrorSeverity) String() string {
	switch s {
	case ErrorSeverityNormal:
		return "ERROR"
	case ErrorSeverityWarning:
		return "WARNING"
	case ErrorSeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// -----------------------------------------------------------------------------
// Error Types
// -----------------------------------------------------------------------------

// GlanceError is the base interface for all custom error types in the application.
type GlanceError interface {
	error
	
	// Type returns the error type identifier
	Type() string
	
	// Code returns the error code
	Code() string
	
	// Severity returns the error severity level
	Severity() ErrorSeverity
	
	// Suggestion returns a recommended action to resolve the error
	Suggestion() string
	
	// Unwrap returns the wrapped error if any
	Unwrap() error
	
	// WithCode sets the error code and returns the error
	WithCode(code string) GlanceError
	
	// WithSeverity sets the error severity and returns the error
	WithSeverity(severity ErrorSeverity) GlanceError
	
	// WithSuggestion sets a suggestion for resolving the error and returns the error
	WithSuggestion(suggestion string) GlanceError
}

// baseError is the common implementation of the GlanceError interface.
type baseError struct {
	errorType  string
	message    string
	code       string
	severity   ErrorSeverity
	suggestion string
	cause      error
}

// Error returns the error message.
func (e *baseError) Error() string {
	var result string
	
	// Include code if present
	if e.code != "" {
		result = fmt.Sprintf("[%s] ", e.code)
	}
	
	// Add the main message
	result += e.message
	
	// Add severity if not normal
	if e.severity != ErrorSeverityNormal {
		result += fmt.Sprintf(" (%s)", e.severity)
	}
	
	// Add suggestion if present
	if e.suggestion != "" {
		result += fmt.Sprintf(" - Suggestion: %s", e.suggestion)
	}
	
	// Add wrapped error if present
	if e.cause != nil {
		result += fmt.Sprintf(": %v", e.cause)
	}
	
	return result
}

// Type returns the error type.
func (e *baseError) Type() string {
	return e.errorType
}

// Code returns the error code.
func (e *baseError) Code() string {
	return e.code
}

// Severity returns the error severity level.
func (e *baseError) Severity() ErrorSeverity {
	return e.severity
}

// Suggestion returns the recommended action to resolve the error.
func (e *baseError) Suggestion() string {
	return e.suggestion
}

// Unwrap returns the wrapped error.
func (e *baseError) Unwrap() error {
	return e.cause
}

// WithCode sets the error code.
func (e *baseError) WithCode(code string) GlanceError {
	e.code = code
	return e
}

// WithSeverity sets the error severity level.
func (e *baseError) WithSeverity(severity ErrorSeverity) GlanceError {
	e.severity = severity
	return e
}

// WithSuggestion sets a suggestion for resolving the error.
func (e *baseError) WithSuggestion(suggestion string) GlanceError {
	e.suggestion = suggestion
	return e
}

// -----------------------------------------------------------------------------
// Specific Error Types
// -----------------------------------------------------------------------------

// FileSystemError represents an error related to file system operations.
type FileSystemError struct {
	baseError
}

// APIError represents an error related to LLM API operations.
type APIError struct {
	baseError
}

// ConfigError represents an error related to configuration.
type ConfigError struct {
	baseError
}

// ValidationError represents an error related to input validation.
type ValidationError struct {
	baseError
}

// -----------------------------------------------------------------------------
// Error Creation Functions
// -----------------------------------------------------------------------------

// New creates a new base error.
func New(message string) GlanceError {
	return &baseError{
		errorType: "General",
		message:   message,
		severity:  ErrorSeverityNormal,
	}
}

// Wrap wraps an existing error with additional context.
func Wrap(err error, message string) GlanceError {
	return &baseError{
		errorType: "General",
		message:   message,
		severity:  ErrorSeverityNormal,
		cause:     err,
	}
}

// NewFileSystemError creates a new file system error.
func NewFileSystemError(message string, cause error) GlanceError {
	return &FileSystemError{
		baseError: baseError{
			errorType: "FileSystem",
			message:   message,
			severity:  ErrorSeverityNormal,
			cause:     cause,
		},
	}
}

// NewAPIError creates a new API error.
func NewAPIError(message string, cause error) GlanceError {
	return &APIError{
		baseError: baseError{
			errorType: "API",
			message:   message,
			severity:  ErrorSeverityNormal,
			cause:     cause,
		},
	}
}

// NewConfigError creates a new configuration error.
func NewConfigError(message string, cause error) GlanceError {
	return &ConfigError{
		baseError: baseError{
			errorType: "Config",
			message:   message,
			severity:  ErrorSeverityNormal,
			cause:     cause,
		},
	}
}

// NewValidationError creates a new validation error.
func NewValidationError(message string, cause error) GlanceError {
	return &ValidationError{
		baseError: baseError{
			errorType: "Validation",
			message:   message,
			severity:  ErrorSeverityNormal,
			cause:     cause,
		},
	}
}

// -----------------------------------------------------------------------------
// Error Type Checking Functions
// -----------------------------------------------------------------------------

// IsFileSystemError checks if an error is a FileSystemError.
func IsFileSystemError(err error) bool {
	var e *FileSystemError
	return errors.As(err, &e)
}

// IsAPIError checks if an error is an APIError.
func IsAPIError(err error) bool {
	var e *APIError
	return errors.As(err, &e)
}

// IsConfigError checks if an error is a ConfigError.
func IsConfigError(err error) bool {
	var e *ConfigError
	return errors.As(err, &e)
}

// IsValidationError checks if an error is a ValidationError.
func IsValidationError(err error) bool {
	var e *ValidationError
	return errors.As(err, &e)
}

// -----------------------------------------------------------------------------
// Sentinel Errors
// -----------------------------------------------------------------------------

// Common file system errors
var (
	ErrFileNotFound       = NewFileSystemError("file not found", nil).WithCode("FS-001")
	ErrFilePermission     = NewFileSystemError("permission denied", nil).WithCode("FS-002")
	ErrDirectoryNotFound  = NewFileSystemError("directory not found", nil).WithCode("FS-003")
	ErrInvalidPath        = NewFileSystemError("invalid path", nil).WithCode("FS-004")
	ErrFileAlreadyExists  = NewFileSystemError("file already exists", nil).WithCode("FS-005")
)

// Common API errors
var (
	ErrAPITimeout         = NewAPIError("API request timed out", nil).WithCode("API-001")
	ErrAPIRateLimit       = NewAPIError("API rate limit exceeded", nil).WithCode("API-002")
	ErrAPIAuthentication  = NewAPIError("API authentication failed", nil).WithCode("API-003")
	ErrAPIQuota           = NewAPIError("API quota exceeded", nil).WithCode("API-004")
	ErrAPIResponseFormat  = NewAPIError("invalid API response format", nil).WithCode("API-005")
)

// Common configuration errors
var (
	ErrConfigMissingKey   = NewConfigError("required configuration key missing", nil).WithCode("CFG-001")
	ErrConfigFormat       = NewConfigError("invalid configuration format", nil).WithCode("CFG-002")
	ErrConfigEnvVar       = NewConfigError("environment variable not set", nil).WithCode("CFG-003")
	ErrConfigValidation   = NewConfigError("configuration validation failed", nil).WithCode("CFG-004")
)

// Common validation errors
var (
	ErrValidationRequired = NewValidationError("required field missing", nil).WithCode("VAL-001")
	ErrValidationFormat   = NewValidationError("invalid format", nil).WithCode("VAL-002")
	ErrValidationRange    = NewValidationError("value out of range", nil).WithCode("VAL-003")
)