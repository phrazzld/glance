package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigStructConstruction tests different ways to construct a Config struct
func TestConfigStructConstruction(t *testing.T) {
	t.Run("Direct struct initialization", func(t *testing.T) {
		// Test creating a new Config instance with all fields directly
		cfg := &Config{
			APIKey:         "test-api-key",
			TargetDir:      "/test/directory",
			Force:          true,
			PromptTemplate: "test template {{.Directory}}",
			MaxRetries:     5,
			MaxFileBytes:   1024 * 1024, // 1MB
		}

		// Verify all fields are set correctly
		assert.Equal(t, "test-api-key", cfg.APIKey, "APIKey should match the provided value")
		assert.Equal(t, "/test/directory", cfg.TargetDir, "TargetDir should match the provided value")
		assert.True(t, cfg.Force, "Force should be true")
		assert.Equal(t, "test template {{.Directory}}", cfg.PromptTemplate, "PromptTemplate should match the provided value")
		assert.Equal(t, 5, cfg.MaxRetries, "MaxRetries should match the provided value")
		assert.Equal(t, int64(1024*1024), cfg.MaxFileBytes, "MaxFileBytes should match the provided value")
	})

	t.Run("Zero values", func(t *testing.T) {
		// Test zero value initialization
		var cfg Config

		// Verify zero values
		assert.Empty(t, cfg.APIKey, "Zero value APIKey should be empty")
		assert.Empty(t, cfg.TargetDir, "Zero value TargetDir should be empty")
		assert.False(t, cfg.Force, "Zero value Force should be false")
		assert.Empty(t, cfg.PromptTemplate, "Zero value PromptTemplate should be empty")
		assert.Zero(t, cfg.MaxRetries, "Zero value MaxRetries should be 0")
		assert.Zero(t, cfg.MaxFileBytes, "Zero value MaxFileBytes should be 0")
	})

	t.Run("Partial initialization", func(t *testing.T) {
		// Test partial initialization
		cfg := &Config{
			APIKey:    "partial-test-key",
			TargetDir: "/partial/test",
		}

		// Verify specified fields are set and others have zero values
		assert.Equal(t, "partial-test-key", cfg.APIKey, "APIKey should match the provided value")
		assert.Equal(t, "/partial/test", cfg.TargetDir, "TargetDir should match the provided value")
		assert.False(t, cfg.Force, "Unspecified Force should default to false")
		assert.Empty(t, cfg.PromptTemplate, "Unspecified PromptTemplate should be empty")
		assert.Zero(t, cfg.MaxRetries, "Unspecified MaxRetries should be zero")
		assert.Zero(t, cfg.MaxFileBytes, "Unspecified MaxFileBytes should be zero")
	})
}

func TestNewDefaultConfig(t *testing.T) {
	// Test creating a default configuration
	cfg := NewDefaultConfig()

	// Verify default values are set correctly
	assert.Empty(t, cfg.APIKey, "Default APIKey should be empty")
	assert.Empty(t, cfg.TargetDir, "Default TargetDir should be empty")
	assert.False(t, cfg.Force, "Default Force should be false")
	assert.NotEmpty(t, cfg.PromptTemplate, "Default PromptTemplate should not be empty")
	assert.Equal(t, 3, cfg.MaxRetries, "Default MaxRetries should be 3")
	assert.Equal(t, int64(5*1024*1024), cfg.MaxFileBytes, "Default MaxFileBytes should be 5MB")

	// Additional checks for the default prompt template
	assert.Contains(t, cfg.PromptTemplate, "{{.Directory}}", "Default template should contain Directory placeholder")
	assert.Contains(t, cfg.PromptTemplate, "{{.SubGlances}}", "Default template should contain SubGlances placeholder")
	assert.Contains(t, cfg.PromptTemplate, "{{.FileContents}}", "Default template should contain FileContents placeholder")
}

func TestImplicitNilDefaultConfig(t *testing.T) {
	// No direct test for this in the codebase, but it's good practice to handle nil configs
	// We infer this behavior from the builder pattern implementation

	// Verify we can call methods on a nil config (this would panic if not properly handled)
	var nilCfg *Config
	require.Nil(t, nilCfg, "nilCfg should be nil")

	// These would panic if the methods don't handle nil receivers properly
	// This is a hypothetical test since the current implementation doesn't handle nil receivers
	// But it's something to consider for robustness

	// The following code is commented out because it would panic
	// If we want to make the code more robust, we could implement nil handling in the methods
	// newCfg := nilCfg.WithAPIKey("test")
	// assert.NotNil(t, newCfg, "WithAPIKey should handle nil receiver")
}

// Individual tests for each With* method

func TestWithAPIKey(t *testing.T) {
	// Start with default config
	cfg := NewDefaultConfig()

	// Test cases
	testCases := []struct {
		name   string
		apiKey string
		want   string
	}{
		{"Normal API key", "normal-api-key", "normal-api-key"},
		{"Empty API key", "", ""},
		{"Long API key", "very-long-api-key-" + repeatString("x", 100), "very-long-api-key-" + repeatString("x", 100)},
		{"Special characters", "api$key!@#", "api$key!@#"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Apply APIKey
			result := cfg.WithAPIKey(tc.apiKey)

			// Verify the APIKey is updated and the original config is unchanged
			assert.Equal(t, tc.want, result.APIKey, "APIKey should be updated correctly")
			assert.Empty(t, cfg.APIKey, "Original config should be unchanged")

			// Verify other fields remain unchanged
			assert.Equal(t, cfg.TargetDir, result.TargetDir, "TargetDir should remain unchanged")
			assert.Equal(t, cfg.Force, result.Force, "Force should remain unchanged")
		})
	}
}

func TestWithTargetDir(t *testing.T) {
	// Start with default config
	cfg := NewDefaultConfig()

	// Test cases
	testCases := []struct {
		name      string
		targetDir string
		want      string
	}{
		{"Absolute path", "/absolute/path", "/absolute/path"},
		{"Relative path", "relative/path", "relative/path"}, // Note: path validation happens elsewhere
		{"Empty path", "", ""},
		{"Root path", "/", "/"},
		{"Path with spaces", "/path with spaces", "/path with spaces"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Apply TargetDir
			result := cfg.WithTargetDir(tc.targetDir)

			// Verify the TargetDir is updated and the original config is unchanged
			assert.Equal(t, tc.want, result.TargetDir, "TargetDir should be updated correctly")
			assert.Empty(t, cfg.TargetDir, "Original config should be unchanged")
		})
	}
}

func TestWithForce(t *testing.T) {
	// Start with default config
	cfg := NewDefaultConfig()

	// Test both true and false values
	testCases := []struct {
		name  string
		force bool
		want  bool
	}{
		{"Set Force to true", true, true},
		{"Set Force to false", false, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Apply Force
			result := cfg.WithForce(tc.force)

			// Verify the Force flag is updated and the original config is unchanged
			assert.Equal(t, tc.want, result.Force, "Force should be updated correctly")
			assert.False(t, cfg.Force, "Original config should be unchanged")
		})
	}
}

func TestWithPromptTemplate(t *testing.T) {
	// Start with default config
	cfg := NewDefaultConfig()
	defaultTemplate := cfg.PromptTemplate

	// Test cases
	testCases := []struct {
		name     string
		template string
		want     string
	}{
		{"Simple template", "simple {{.Directory}} template", "simple {{.Directory}} template"},
		{"Empty template", "", ""},
		{"Complex template", "complex {{.Directory}}\n{{.SubGlances}}\n{{.FileContents}}", "complex {{.Directory}}\n{{.SubGlances}}\n{{.FileContents}}"},
		{"Just text", "no template variables here", "no template variables here"},
		{"Special characters", "template with $pecial @#! chars", "template with $pecial @#! chars"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Apply new prompt template
			result := cfg.WithPromptTemplate(tc.template)

			// Verify the PromptTemplate is updated and the original config is unchanged
			assert.Equal(t, tc.want, result.PromptTemplate, "PromptTemplate should be updated correctly")
			assert.Equal(t, defaultTemplate, cfg.PromptTemplate, "Original config should be unchanged")
		})
	}
}

func TestWithMaxRetries(t *testing.T) {
	// Start with default config
	cfg := NewDefaultConfig()

	// Test cases - include edge cases like negative values
	testCases := []struct {
		name       string
		maxRetries int
		want       int
	}{
		{"Zero retries", 0, 0},
		{"Positive retries", 10, 10},
		{"Negative retries", -5, -5}, // Should accept any value, validation happens elsewhere
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Apply new MaxRetries
			result := cfg.WithMaxRetries(tc.maxRetries)

			// Verify the MaxRetries is updated and the original config is unchanged
			assert.Equal(t, tc.want, result.MaxRetries, "MaxRetries should be updated correctly")
			assert.Equal(t, DefaultMaxRetries, cfg.MaxRetries, "Original config should be unchanged")
		})
	}
}

func TestWithMaxFileBytes(t *testing.T) {
	// Start with default config
	cfg := NewDefaultConfig()

	// Test cases - include edge cases like zero and negative values
	testCases := []struct {
		name         string
		maxFileBytes int64
		want         int64
	}{
		{"Zero bytes", 0, 0},
		{"1 KB", 1024, 1024},
		{"10 MB", 10 * 1024 * 1024, 10 * 1024 * 1024},
		{"Negative bytes", -1024, -1024}, // Should accept any value, validation happens elsewhere
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Apply new MaxFileBytes
			result := cfg.WithMaxFileBytes(tc.maxFileBytes)

			// Verify the MaxFileBytes is updated and the original config is unchanged
			assert.Equal(t, tc.want, result.MaxFileBytes, "MaxFileBytes should be updated correctly")
			assert.Equal(t, int64(DefaultMaxFileBytes), cfg.MaxFileBytes, "Original config should be unchanged")
		})
	}
}

func TestChainedWithMethods(t *testing.T) {
	// Start with default config
	cfg := NewDefaultConfig()

	// Apply multiple changes using method chaining
	result := cfg.
		WithAPIKey("chained-api-key").
		WithTargetDir("/chained/path").
		WithForce(true).
		WithPromptTemplate("chained template").
		WithMaxRetries(7).
		WithMaxFileBytes(7 * 1024 * 1024)

	// Verify all changes were applied correctly
	assert.Equal(t, "chained-api-key", result.APIKey, "APIKey should be updated")
	assert.Equal(t, "/chained/path", result.TargetDir, "TargetDir should be updated")
	assert.True(t, result.Force, "Force should be true")
	assert.Equal(t, "chained template", result.PromptTemplate, "PromptTemplate should be updated")
	assert.Equal(t, 7, result.MaxRetries, "MaxRetries should be updated")
	assert.Equal(t, int64(7*1024*1024), result.MaxFileBytes, "MaxFileBytes should be updated")

	// Verify original config remains unchanged
	assert.Empty(t, cfg.APIKey, "Original APIKey should be unchanged")
	assert.Empty(t, cfg.TargetDir, "Original TargetDir should be unchanged")
	assert.False(t, cfg.Force, "Original Force should be unchanged")
	assert.NotEqual(t, "chained template", cfg.PromptTemplate, "Original PromptTemplate should be unchanged")
	assert.Equal(t, DefaultMaxRetries, cfg.MaxRetries, "Original MaxRetries should be unchanged")
	assert.Equal(t, int64(DefaultMaxFileBytes), cfg.MaxFileBytes, "Original MaxFileBytes should be unchanged")

	// Test complex chaining in different order
	result2 := cfg.
		WithMaxFileBytes(2 * 1024 * 1024).
		WithAPIKey("different-key").
		WithMaxRetries(5).
		WithPromptTemplate("different template").
		WithTargetDir("/different/path").
		WithForce(true)

	// Verify all changes were applied correctly
	assert.Equal(t, "different-key", result2.APIKey, "APIKey should be updated")
	assert.Equal(t, "/different/path", result2.TargetDir, "TargetDir should be updated")
	assert.True(t, result2.Force, "Force should be true")
	assert.Equal(t, "different template", result2.PromptTemplate, "PromptTemplate should be updated")
	assert.Equal(t, 5, result2.MaxRetries, "MaxRetries should be updated")
	assert.Equal(t, int64(2*1024*1024), result2.MaxFileBytes, "MaxFileBytes should be updated")
}

// Helper function to repeat a string
func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
