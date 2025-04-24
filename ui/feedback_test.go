package ui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/briandowns/spinner"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Since spinners and progress bars don't consistently write to stdout in tests,
// we'll focus on testing the configuration rather than the actual output

// Helper function to capture logrus output during tests
func captureLogOutput(fn func()) string {
	var buf strings.Builder
	originalOutput := logrus.StandardLogger().Out
	logrus.SetOutput(&buf)
	defer logrus.SetOutput(originalOutput)
	fn()
	return buf.String()
}

// -----------------------------------------------------------------------------
// Spinner Tests
// -----------------------------------------------------------------------------

func TestNewScanner(t *testing.T) {
	// Create a spinner
	spinner := NewScanner()

	// Verify spinner is properly configured
	assert.NotNil(t, spinner, "Spinner should not be nil")
	assert.Contains(t, spinner.suffix, "Scanning")
	assert.Contains(t, spinner.finalMsg, "Scan complete")
	assert.NotNil(t, spinner.spinner, "Underlying spinner should be initialized")

	// Check default values are properly set
	assert.Equal(t, 120*time.Millisecond, spinner.speed)
}

func TestNewGenerator(t *testing.T) {
	// Create a generator spinner
	spinner := NewGenerator()

	// Verify spinner is properly configured
	assert.NotNil(t, spinner, "Generator spinner should not be nil")
	assert.Contains(t, spinner.suffix, "Generating")
	assert.Contains(t, spinner.finalMsg, "Generation complete")
	assert.NotNil(t, spinner.spinner, "Underlying spinner should be initialized")
}

func TestNewCustomSpinner(t *testing.T) {
	// Test creation with no options (default values)
	t.Run("Default values", func(t *testing.T) {
		s := NewCustomSpinner()
		assert.NotNil(t, s, "Spinner should not be nil")
		assert.Equal(t, "Processing...", s.suffix)
		assert.Equal(t, "Done!\n", s.finalMsg)
		assert.Equal(t, 120*time.Millisecond, s.speed)
		assert.NotNil(t, s.spinner, "Underlying spinner should be initialized")
	})

	// Test with custom options
	t.Run("Custom options", func(t *testing.T) {
		s := NewCustomSpinner(
			WithSuffix("Test suffix"),
			WithFinalMessage("Test final message"),
			WithCharset(9),
			WithSpeed(50*time.Millisecond),
		)

		assert.Equal(t, "Test suffix", s.suffix)
		assert.Equal(t, "Test final message", s.finalMsg)
		assert.Equal(t, 50*time.Millisecond, s.speed)
	})

	// Test with invalid charset
	t.Run("Invalid charset", func(t *testing.T) {
		// Test with negative charset index (should use default)
		s1 := NewCustomSpinner(WithCharset(-1))
		assert.NotNil(t, s1.spinner, "Spinner should be initialized even with invalid charset")

		// Test with too large charset index (should use default)
		s2 := NewCustomSpinner(WithCharset(1000))
		assert.NotNil(t, s2.spinner, "Spinner should be initialized even with invalid charset")
	})
}

func TestSpinnerStartStop(t *testing.T) {
	// Create a spinner
	s := NewCustomSpinner(WithSuffix("Testing"), WithFinalMessage("Done testing"))

	// Just test that Start and Stop don't panic
	s.Start()
	time.Sleep(10 * time.Millisecond) // Give it time to spin
	s.Stop()

	// Success if we get here without panicking
	assert.True(t, true)
}

func TestSpinnerUpdateMessage(t *testing.T) {
	// Create a spinner
	s := NewCustomSpinner(WithSuffix("Initial message"))
	assert.Equal(t, "Initial message", s.suffix)

	// Update the message
	s.UpdateMessage("Updated message")
	assert.Equal(t, " Updated message", s.spinner.Suffix, "Spinner message should be updated")
}

func TestCustomSpinnerOptions(t *testing.T) {
	// Individual tests for each option function

	t.Run("WithSuffix", func(t *testing.T) {
		opt := WithSuffix("Test suffix")
		s := &Spinner{
			spinner: spinner.New(spinner.CharSets[0], 100*time.Millisecond),
		}
		opt(s)
		assert.Equal(t, "Test suffix", s.suffix)
		assert.Equal(t, " Test suffix", s.spinner.Suffix)
	})

	t.Run("WithFinalMessage", func(t *testing.T) {
		opt := WithFinalMessage("Test final message")
		s := &Spinner{}
		opt(s)
		assert.Equal(t, "Test final message", s.finalMsg)
	})

	t.Run("WithCharset", func(t *testing.T) {
		// We can't easily test the charset change since it's not directly accessible
		// Instead we'll verify that the function is called without error
		s := &Spinner{
			spinner: spinner.New(spinner.CharSets[0], 100*time.Millisecond),
		}

		// With valid charset
		WithCharset(5)(s)
		assert.NotNil(t, s.spinner, "Spinner should still be valid after changing charset")

		// With invalid charset (negative)
		WithCharset(-5)(s)
		assert.NotNil(t, s.spinner, "Spinner should still be valid after invalid charset")

		// With invalid charset (too large)
		WithCharset(len(spinner.CharSets) + 10)(s)
		assert.NotNil(t, s.spinner, "Spinner should still be valid after invalid charset")
	})

	t.Run("WithSpeed", func(t *testing.T) {
		opt := WithSpeed(75 * time.Millisecond)
		s := &Spinner{
			spinner: spinner.New(spinner.CharSets[0], 100*time.Millisecond),
		}
		opt(s)
		assert.Equal(t, 75*time.Millisecond, s.speed)
	})

	// Test combining multiple options
	t.Run("Multiple options", func(t *testing.T) {
		// Test that options properly modify the spinner
		s := NewCustomSpinner(
			WithSuffix("Custom suffix"),
			WithFinalMessage("Custom final message"),
			WithCharset(1), // Different charset
			WithSpeed(100*time.Millisecond),
		)

		assert.Equal(t, "Custom suffix", s.suffix)
		assert.Equal(t, "Custom final message", s.finalMsg)
		assert.Equal(t, 100*time.Millisecond, s.speed)
	})
}

// Progress bar tests have been removed as part of the simplification process.
// Progress bars are now used directly in glance.go with the progressbar library.

// -----------------------------------------------------------------------------
// Error Reporting Tests
// -----------------------------------------------------------------------------

func TestReportError(t *testing.T) {
	// Save original log output and level
	originalOutput := logrus.StandardLogger().Out
	originalLevel := logrus.GetLevel()
	defer func() {
		logrus.SetOutput(originalOutput)
		logrus.SetLevel(originalLevel)
	}()

	// Set log level to enable error logs
	logrus.SetLevel(logrus.ErrorLevel)

	// Test with nil error (should not log anything)
	t.Run("Nil error", func(t *testing.T) {
		output := captureLogOutput(func() {
			ReportError(nil, "Test context")
		})
		assert.Empty(t, output, "No output should be logged for nil error")
	})

	// Test with non-nil error
	t.Run("Error logging", func(t *testing.T) {
		testErr := errors.New("test error")
		output := captureLogOutput(func() {
			ReportError(testErr, "Test context")
		})
		assert.Contains(t, output, "Test context")
		assert.Contains(t, output, "test error")
		assert.NotContains(t, output, "❌")
	})

	// Test with different context values
	t.Run("Different context values", func(t *testing.T) {
		testErr := errors.New("test error")

		output1 := captureLogOutput(func() {
			ReportError(testErr, "Context 1")
		})
		assert.Contains(t, output1, "Context 1")

		output2 := captureLogOutput(func() {
			ReportError(testErr, "Context 2")
		})
		assert.Contains(t, output2, "Context 2")
	})
}

// -----------------------------------------------------------------------------
// Integration Tests
// -----------------------------------------------------------------------------

func TestSpinnerOnly(t *testing.T) {
	// This test shows basic spinner functionality

	// Create a spinner for initialization
	spinner := NewScanner()
	spinner.Start()

	// Simulate some initialization work
	time.Sleep(50 * time.Millisecond)

	// Stop the spinner
	spinner.Stop()

	// If we got here without panicking, the test passes
	assert.True(t, true)
}
