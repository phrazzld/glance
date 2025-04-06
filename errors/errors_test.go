package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseError(t *testing.T) {
	// Test creating a new base error
	err := New("test error")
	assert.NotNil(t, err)
	assert.Equal(t, "test error", err.Error())
	assert.Equal(t, ErrorSeverityNormal, err.Severity())
	assert.Empty(t, err.Suggestion())
}

func TestErrorWithSeverity(t *testing.T) {
	// Test setting severity
	err := New("critical error").WithSeverity(ErrorSeverityCritical)
	assert.Equal(t, ErrorSeverityCritical, err.Severity())
	assert.Contains(t, err.Error(), "critical error")
}

func TestErrorWithSuggestion(t *testing.T) {
	// Test adding suggestion
	suggestion := "try rebooting"
	err := New("system error").WithSuggestion(suggestion)
	assert.Equal(t, suggestion, err.Suggestion())
}

func TestWrappedError(t *testing.T) {
	// Create a standard error
	stdErr := fmt.Errorf("standard error")
	
	// Wrap it with our custom error
	glanceErr := Wrap(stdErr, "wrapped error")
	
	// Test error message formatting
	assert.Contains(t, glanceErr.Error(), "wrapped error")
	assert.Contains(t, glanceErr.Error(), "standard error")
	
	// Test unwrapping
	unwrapped := errors.Unwrap(glanceErr)
	assert.Equal(t, stdErr, unwrapped)
	
	// Test errors.Is
	assert.True(t, errors.Is(glanceErr, stdErr))
}

func TestErrorTypes(t *testing.T) {
	// Create different error types
	fsErr := NewFileSystemError("file not found", nil)
	configErr := NewConfigError("invalid config", nil)
	apiErr := NewAPIError("API timeout", nil)
	validationErr := NewValidationError("invalid input", nil)
	
	// Verify error types
	assert.True(t, IsFileSystemError(fsErr))
	assert.True(t, IsConfigError(configErr))
	assert.True(t, IsAPIError(apiErr))
	assert.True(t, IsValidationError(validationErr))
	
	// Verify cross-type checks
	assert.False(t, IsFileSystemError(apiErr))
	assert.False(t, IsConfigError(fsErr))
	assert.False(t, IsAPIError(validationErr))
	assert.False(t, IsValidationError(configErr))
}

func TestErrorCodes(t *testing.T) {
	// Create error with code
	err := NewFileSystemError("file error", nil).WithCode("FS-001")
	
	// Verify code was set
	assert.Equal(t, "FS-001", err.Code())
	
	// Verify code appears in error message
	assert.Contains(t, err.Error(), "FS-001")
}

func TestErrorUnwrapping(t *testing.T) {
	// Create a chain of errors
	baseErr := errors.New("original error")
	wrapped1 := Wrap(baseErr, "wrapped once")
	wrapped2 := NewAPIError("API error", wrapped1)
	
	// Unwrap to the original
	originalErr := errors.Unwrap(errors.Unwrap(wrapped2))
	assert.Equal(t, baseErr, originalErr)
	
	// Test Is functionality
	assert.True(t, errors.Is(wrapped2, baseErr))
}

func TestSentinelErrors(t *testing.T) {
	// Test using sentinel errors
	err1 := NewAPIError("timeout", ErrAPITimeout)
	err2 := NewConfigError("missing key", ErrConfigMissingKey)
	
	// Verify sentinel errors can be detected
	assert.True(t, errors.Is(err1, ErrAPITimeout))
	assert.True(t, errors.Is(err2, ErrConfigMissingKey))
}

func TestErrorFormat(t *testing.T) {
	// Create an error with all fields
	err := New("test error").
		WithCode("TST-001").
		WithSeverity(ErrorSeverityCritical).
		WithSuggestion("restart application")
	
	// Verify the error string format
	expectedParts := []string{
		"[TST-001]",
		"test error",
		"CRITICAL",
		"restart application",
	}
	
	for _, part := range expectedParts {
		assert.Contains(t, err.Error(), part)
	}
}