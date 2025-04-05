package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigStruct(t *testing.T) {
	// Test creating a new Config instance with all fields
	cfg := &Config{
		APIKey:         "test-api-key",
		TargetDir:      "/test/directory",
		Force:          true,
		Verbose:        true,
		PromptTemplate: "test template {{.Directory}}",
		MaxRetries:     5,
		MaxFileBytes:   1024 * 1024, // 1MB
	}

	// Verify all fields are set correctly
	assert.Equal(t, "test-api-key", cfg.APIKey, "APIKey should match the provided value")
	assert.Equal(t, "/test/directory", cfg.TargetDir, "TargetDir should match the provided value")
	assert.True(t, cfg.Force, "Force should be true")
	assert.True(t, cfg.Verbose, "Verbose should be true")
	assert.Equal(t, "test template {{.Directory}}", cfg.PromptTemplate, "PromptTemplate should match the provided value")
	assert.Equal(t, 5, cfg.MaxRetries, "MaxRetries should match the provided value")
	assert.Equal(t, int64(1024*1024), cfg.MaxFileBytes, "MaxFileBytes should match the provided value")
}

func TestNewDefaultConfig(t *testing.T) {
	// Test creating a default configuration
	cfg := NewDefaultConfig()

	// Verify default values are set correctly
	assert.Empty(t, cfg.APIKey, "Default APIKey should be empty")
	assert.Empty(t, cfg.TargetDir, "Default TargetDir should be empty")
	assert.False(t, cfg.Force, "Default Force should be false")
	assert.False(t, cfg.Verbose, "Default Verbose should be false")
	assert.NotEmpty(t, cfg.PromptTemplate, "Default PromptTemplate should not be empty")
	assert.Equal(t, 3, cfg.MaxRetries, "Default MaxRetries should be 3")
	assert.Equal(t, int64(5*1024*1024), cfg.MaxFileBytes, "Default MaxFileBytes should be 5MB")
}

func TestWithAPIKey(t *testing.T) {
	// Start with default config
	cfg := NewDefaultConfig()
	
	// Apply APIKey
	result := cfg.WithAPIKey("new-api-key")
	
	// Verify the APIKey is updated and the original config is unchanged
	assert.Equal(t, "new-api-key", result.APIKey, "APIKey should be updated")
	assert.Empty(t, cfg.APIKey, "Original config should be unchanged")
	
	// Verify other fields remain unchanged
	assert.Equal(t, cfg.MaxRetries, result.MaxRetries, "MaxRetries should remain unchanged")
}

func TestWithTargetDir(t *testing.T) {
	// Start with default config
	cfg := NewDefaultConfig()
	
	// Apply TargetDir
	result := cfg.WithTargetDir("/new/target")
	
	// Verify the TargetDir is updated and the original config is unchanged
	assert.Equal(t, "/new/target", result.TargetDir, "TargetDir should be updated")
	assert.Empty(t, cfg.TargetDir, "Original config should be unchanged")
}

func TestWithForce(t *testing.T) {
	// Start with default config
	cfg := NewDefaultConfig()
	
	// Apply Force
	result := cfg.WithForce(true)
	
	// Verify the Force flag is updated and the original config is unchanged
	assert.True(t, result.Force, "Force should be updated to true")
	assert.False(t, cfg.Force, "Original config should be unchanged")
}

func TestWithVerbose(t *testing.T) {
	// Start with default config
	cfg := NewDefaultConfig()
	
	// Apply Verbose
	result := cfg.WithVerbose(true)
	
	// Verify the Verbose flag is updated and the original config is unchanged
	assert.True(t, result.Verbose, "Verbose should be updated to true")
	assert.False(t, cfg.Verbose, "Original config should be unchanged")
}

func TestWithPromptTemplate(t *testing.T) {
	// Start with default config
	cfg := NewDefaultConfig()
	defaultTemplate := cfg.PromptTemplate
	
	// Apply new prompt template
	newTemplate := "new template {{.Directory}}"
	result := cfg.WithPromptTemplate(newTemplate)
	
	// Verify the PromptTemplate is updated and the original config is unchanged
	assert.Equal(t, newTemplate, result.PromptTemplate, "PromptTemplate should be updated")
	assert.Equal(t, defaultTemplate, cfg.PromptTemplate, "Original config should be unchanged")
}

func TestWithMaxRetries(t *testing.T) {
	// Start with default config
	cfg := NewDefaultConfig()
	
	// Apply new MaxRetries
	newRetries := 10
	result := cfg.WithMaxRetries(newRetries)
	
	// Verify the MaxRetries is updated and the original config is unchanged
	assert.Equal(t, newRetries, result.MaxRetries, "MaxRetries should be updated")
	assert.Equal(t, DefaultMaxRetries, cfg.MaxRetries, "Original config should be unchanged")
}

func TestWithMaxFileBytes(t *testing.T) {
	// Start with default config
	cfg := NewDefaultConfig()
	
	// Apply new MaxFileBytes
	newMaxBytes := int64(10 * 1024 * 1024) // 10 MB
	result := cfg.WithMaxFileBytes(newMaxBytes)
	
	// Verify the MaxFileBytes is updated and the original config is unchanged
	assert.Equal(t, newMaxBytes, result.MaxFileBytes, "MaxFileBytes should be updated")
	assert.Equal(t, int64(DefaultMaxFileBytes), cfg.MaxFileBytes, "Original config should be unchanged")
}

func TestChainedWithMethods(t *testing.T) {
	// Start with default config
	cfg := NewDefaultConfig()
	
	// Apply multiple changes using method chaining
	result := cfg.
		WithAPIKey("chained-api-key").
		WithTargetDir("/chained/path").
		WithForce(true).
		WithVerbose(true).
		WithPromptTemplate("chained template").
		WithMaxRetries(7).
		WithMaxFileBytes(7 * 1024 * 1024)
	
	// Verify all changes were applied correctly
	assert.Equal(t, "chained-api-key", result.APIKey, "APIKey should be updated")
	assert.Equal(t, "/chained/path", result.TargetDir, "TargetDir should be updated")
	assert.True(t, result.Force, "Force should be true")
	assert.True(t, result.Verbose, "Verbose should be true")
	assert.Equal(t, "chained template", result.PromptTemplate, "PromptTemplate should be updated")
	assert.Equal(t, 7, result.MaxRetries, "MaxRetries should be updated")
	assert.Equal(t, int64(7*1024*1024), result.MaxFileBytes, "MaxFileBytes should be updated")
	
	// Verify original config remains unchanged
	assert.Empty(t, cfg.APIKey, "Original APIKey should be unchanged")
	assert.Empty(t, cfg.TargetDir, "Original TargetDir should be unchanged")
	assert.False(t, cfg.Force, "Original Force should be unchanged")
	assert.False(t, cfg.Verbose, "Original Verbose should be unchanged")
	assert.NotEqual(t, "chained template", cfg.PromptTemplate, "Original PromptTemplate should be unchanged")
	assert.Equal(t, DefaultMaxRetries, cfg.MaxRetries, "Original MaxRetries should be unchanged")
	assert.Equal(t, int64(DefaultMaxFileBytes), cfg.MaxFileBytes, "Original MaxFileBytes should be unchanged")
}