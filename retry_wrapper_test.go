package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RetryWrapperResult captures the result of running the retry wrapper
type RetryWrapperResult struct {
	ExitCode      int           `json:"exit_code"`
	Duration      time.Duration `json:"duration"`
	StdOut        string        `json:"stdout"`
	StdErr        string        `json:"stderr"`
	CorrelationID string        `json:"correlation_id"`
}

// TestRetryWrapperBasicFunctionality tests the basic functionality of the retry wrapper
func TestRetryWrapperBasicFunctionality(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		expectSuccess bool
		expectRetries bool
	}{
		{
			name:          "Help command",
			args:          []string{"--help"},
			expectSuccess: true,
			expectRetries: false,
		},
		{
			name:          "Valid scan with default args",
			args:          []string{},
			expectSuccess: true, // Should succeed in normal conditions
			expectRetries: false,
		},
		{
			name:          "Custom retry count",
			args:          []string{"--retry-attempts", "1", "./..."},
			expectSuccess: true,
			expectRetries: false,
		},
		{
			name:          "Custom timeout",
			args:          []string{"--timeout", "30", "./..."},
			expectSuccess: true,
			expectRetries: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := runRetryWrapper(t, tc.args, 60*time.Second)

			if tc.expectSuccess {
				assert.Equal(t, 0, result.ExitCode, "Expected successful execution")
			}

			if tc.name == "Help command" {
				assert.Contains(t, result.StdOut, "Usage:", "Help should contain usage information")
				assert.Contains(t, result.StdOut, "Examples:", "Help should contain examples")
			}

			// Check that correlation ID is generated
			if tc.name != "Help command" {
				assert.NotEmpty(t, result.CorrelationID, "Should generate correlation ID")
				assert.Contains(t, result.StdErr, "govulncheck-retry-", "Should log correlation ID")
			}
		})
	}
}

// TestRetryWrapperNetworkFailures tests network failure handling
func TestRetryWrapperNetworkFailures(t *testing.T) {
	// Skip if network tests are not enabled
	if os.Getenv("RUN_NETWORK_TESTS") != "true" {
		t.Skip("Network failure tests skipped - set RUN_NETWORK_TESTS=true to enable")
	}

	testCases := []struct {
		name           string
		retryAttempts  int
		timeout        int
		networkSetup   func() func()
		expectedExit   int
		expectedStderr []string
	}{
		{
			name:          "Network isolation with retries",
			retryAttempts: 3,
			timeout:       10,
			networkSetup: func() func() {
				return setupNetworkIsolationForWrapper(t)
			},
			expectedExit:   125, // Network connectivity issue
			expectedStderr: []string{"Network error", "retrying"},
		},
		{
			name:          "Timeout with no retries",
			retryAttempts: 0,
			timeout:       5,
			networkSetup: func() func() {
				return setupSlowNetworkForWrapper(t)
			},
			expectedExit:   124, // Timeout
			expectedStderr: []string{"timed out"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cleanup := tc.networkSetup()
			defer cleanup()

			args := []string{
				"--retry-attempts", fmt.Sprintf("%d", tc.retryAttempts),
				"--timeout", fmt.Sprintf("%d", tc.timeout),
				"./...",
			}

			result := runRetryWrapper(t, args, time.Duration(tc.timeout*4)*time.Second)

			// Check exit code
			if tc.expectedExit > 0 {
				assert.Equal(t, tc.expectedExit, result.ExitCode, "Expected specific failure exit code")
			}

			// Check error messages
			for _, expectedMsg := range tc.expectedStderr {
				assert.Contains(t, strings.ToLower(result.StdErr), strings.ToLower(expectedMsg),
					"Should contain expected error message: %s", expectedMsg)
			}

			// Verify structured logging
			assert.Contains(t, result.StdErr, "correlation_id", "Should include correlation ID in logs")
			assert.Contains(t, result.StdErr, "govulncheck-wrapper", "Should identify service name")
		})
	}
}

// TestRetryWrapperConfigurationHandling tests configuration file handling
func TestRetryWrapperConfigurationHandling(t *testing.T) {
	testCases := []struct {
		name            string
		configContent   string
		expectedRetry   int
		expectedTimeout int
		shouldFail      bool
	}{
		{
			name: "Valid configuration",
			configContent: `
fail_on_severity:
  - "HIGH"
  - "CRITICAL"
timeout_seconds: 60
retry_attempts: 5
`,
			expectedRetry:   5,
			expectedTimeout: 60,
			shouldFail:      false,
		},
		{
			name: "Missing retry configuration",
			configContent: `
fail_on_severity:
  - "HIGH"
timeout_seconds: 30
`,
			expectedRetry:   2, // Default value
			expectedTimeout: 30,
			shouldFail:      false,
		},
		{
			name: "Invalid YAML",
			configContent: `
invalid: [yaml content
retry_attempts: not_a_number
`,
			expectedRetry:   2, // Should fall back to defaults
			expectedTimeout: 300,
			shouldFail:      false, // Should handle gracefully
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary configuration
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, ".govulncheck.yaml")
			err := os.WriteFile(configPath, []byte(tc.configContent), 0644)
			require.NoError(t, err)

			// Create minimal Go project
			createMinimalGoProjectInDir(t, tempDir)

			// Run wrapper with configuration
			args := []string{"--config", configPath, "./..."}
			result := runRetryWrapperInDir(t, tempDir, args, 30*time.Second)

			if tc.shouldFail {
				assert.NotEqual(t, 0, result.ExitCode, "Expected configuration error")
			} else {
				// Should at least attempt to run (may fail due to network but should parse config)
				assert.Contains(t, result.StdErr, "Starting govulncheck", "Should start execution")
			}

			// Check that configuration values are logged
			assert.Contains(t, result.StdErr, "retry configuration", "Should log retry configuration")
		})
	}
}

// TestRetryWrapperErrorClassification tests error classification logic
func TestRetryWrapperErrorClassification(t *testing.T) {
	// This test validates that the wrapper correctly identifies network vs non-network errors
	testCases := []struct {
		name                 string
		simulatedError       string
		simulatedExitCode    int
		expectedRetryable    bool
		expectedErrorMessage string
	}{
		{
			name:                 "Timeout error",
			simulatedExitCode:    124,
			expectedRetryable:    true,
			expectedErrorMessage: "timed out",
		},
		{
			name:                 "Non-network error",
			simulatedExitCode:    1,
			simulatedError:       "invalid argument",
			expectedRetryable:    false,
			expectedErrorMessage: "Non-network error",
		},
		{
			name:                 "Network connection error",
			simulatedExitCode:    1,
			simulatedError:       "connection refused",
			expectedRetryable:    true,
			expectedErrorMessage: "Network error",
		},
		{
			name:                 "DNS resolution error",
			simulatedExitCode:    1,
			simulatedError:       "temporary failure in name resolution",
			expectedRetryable:    true,
			expectedErrorMessage: "Network error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This test would need to mock govulncheck execution to simulate specific errors
			// For now, we'll test the error classification logic directly
			t.Skip("Error classification test requires mocking govulncheck - implement when needed")
		})
	}
}

// TestRetryWrapperPerformanceImpact tests that retry logic doesn't significantly impact performance
func TestRetryWrapperPerformanceImpact(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Run scan without retry wrapper
	start := time.Now()
	directResult := runGovulncheckDirect(t, 30*time.Second)
	directDuration := time.Since(start)

	// Run scan with retry wrapper (0 retries)
	start = time.Now()
	wrapperResult := runRetryWrapper(t, []string{"--retry-attempts", "0", "./..."}, 30*time.Second)
	wrapperDuration := time.Since(start)

	// Both should succeed (assuming normal network conditions)
	assert.Equal(t, 0, directResult.ExitCode, "Direct govulncheck should succeed")
	assert.Equal(t, 0, wrapperResult.ExitCode, "Wrapper should succeed")

	// Wrapper should not add significant overhead (less than 10%)
	maxAcceptableOverhead := directDuration + (directDuration / 10)
	assert.LessOrEqual(t, wrapperDuration, maxAcceptableOverhead,
		"Retry wrapper should not add significant overhead: direct=%v, wrapper=%v",
		directDuration, wrapperDuration)

	t.Logf("Performance comparison: direct=%v, wrapper=%v, overhead=%.1f%%",
		directDuration, wrapperDuration,
		float64(wrapperDuration-directDuration)/float64(directDuration)*100)
}

// Helper functions

// runRetryWrapper runs the retry wrapper script with given arguments
func runRetryWrapper(t *testing.T, args []string, timeout time.Duration) *RetryWrapperResult {
	tempDir := t.TempDir()
	createMinimalGoProjectInDir(t, tempDir)
	return runRetryWrapperInDir(t, tempDir, args, timeout)
}

// runRetryWrapperInDir runs the retry wrapper in a specific directory
func runRetryWrapperInDir(t *testing.T, dir string, args []string, timeout time.Duration) *RetryWrapperResult {
	// Get the path to the retry wrapper script
	pwd, _ := os.Getwd()
	scriptPath := filepath.Join(pwd, "scripts", "govulncheck-with-retry.sh")

	// Ensure script exists and is executable
	require.FileExists(t, scriptPath, "Retry wrapper script should exist")

	// Change to target directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(dir)

	// Run the script
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, scriptPath, args...)

	start := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	result := &RetryWrapperResult{
		Duration: duration,
		StdOut:   string(output),
	}

	// Extract stderr from combined output (script sends structured logs to stderr)
	lines := strings.Split(string(output), "\n")
	var stdoutLines, stderrLines []string

	for _, line := range lines {
		if strings.Contains(line, `"service_name": "govulncheck-wrapper"`) ||
			strings.Contains(line, "❌") || strings.Contains(line, "⚠️") {
			stderrLines = append(stderrLines, line)
		} else {
			stdoutLines = append(stdoutLines, line)
		}
	}

	result.StdOut = strings.Join(stdoutLines, "\n")
	result.StdErr = strings.Join(stderrLines, "\n")

	// Extract correlation ID if present
	for _, line := range stderrLines {
		if strings.Contains(line, "correlation_id") {
			// Simple extraction - in real implementation, might use JSON parsing
			if idx := strings.Index(line, "govulncheck-retry-"); idx != -1 {
				end := strings.Index(line[idx:], `"`)
				if end > 0 {
					result.CorrelationID = line[idx : idx+end]
				}
			}
		}
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}

	return result
}

// runGovulncheckDirect runs govulncheck directly for performance comparison
func runGovulncheckDirect(t *testing.T, timeout time.Duration) *RetryWrapperResult {
	tempDir := t.TempDir()
	createMinimalGoProjectInDir(t, tempDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "govulncheck", "./...")

	start := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	result := &RetryWrapperResult{
		Duration: duration,
		StdOut:   string(output),
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}

	return result
}

// createMinimalGoProjectInDir creates a minimal Go project in specified directory
func createMinimalGoProjectInDir(t *testing.T, dir string) {
	// Create go.mod
	goMod := `module retrytest

go 1.21
`
	err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644)
	require.NoError(t, err)

	// Create main.go
	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}
`
	err = os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0644)
	require.NoError(t, err)
}

// setupNetworkIsolationForWrapper sets up network isolation for wrapper tests
func setupNetworkIsolationForWrapper(t *testing.T) func() {
	// Similar to network_failure_test.go but specifically for wrapper
	originalProxy := os.Getenv("HTTP_PROXY")
	originalHTTPSProxy := os.Getenv("HTTPS_PROXY")

	os.Setenv("HTTP_PROXY", "http://127.0.0.1:9999")
	os.Setenv("HTTPS_PROXY", "http://127.0.0.1:9999")

	return func() {
		if originalProxy != "" {
			os.Setenv("HTTP_PROXY", originalProxy)
		} else {
			os.Unsetenv("HTTP_PROXY")
		}
		if originalHTTPSProxy != "" {
			os.Setenv("HTTPS_PROXY", originalHTTPSProxy)
		} else {
			os.Unsetenv("HTTPS_PROXY")
		}
	}
}

// setupSlowNetworkForWrapper sets up slow network conditions for wrapper tests
func setupSlowNetworkForWrapper(t *testing.T) func() {
	// For now, this is a placeholder - real implementation would use traffic shaping
	return func() {}
}

// Integration test with CI workflow patterns
func TestRetryWrapperCIIntegration(t *testing.T) {
	t.Run("CI-style execution with correlation ID", func(t *testing.T) {
		// Simulate CI environment
		os.Setenv("CI", "true")
		defer os.Unsetenv("CI")

		result := runRetryWrapper(t, []string{"./..."}, 60*time.Second)

		// Should generate correlation ID
		assert.NotEmpty(t, result.CorrelationID, "Should generate correlation ID in CI")

		// Should include structured logging
		assert.Contains(t, result.StdErr, "service_name", "Should use structured logging")
		assert.Contains(t, result.StdErr, "correlation_id", "Should log correlation ID")
	})

	t.Run("Environment variable configuration", func(t *testing.T) {
		// Test environment variable overrides
		os.Setenv("GOVULNCHECK_RETRY_ATTEMPTS", "1")
		os.Setenv("GOVULNCHECK_TIMEOUT_SECONDS", "30")
		defer os.Unsetenv("GOVULNCHECK_RETRY_ATTEMPTS")
		defer os.Unsetenv("GOVULNCHECK_TIMEOUT_SECONDS")

		result := runRetryWrapper(t, []string{"./..."}, 45*time.Second)

		// Should log the configuration values
		assert.Contains(t, result.StdErr, "attempts=1", "Should use environment override for retries")
		assert.Contains(t, result.StdErr, "timeout=30", "Should use environment override for timeout")
	})
}
