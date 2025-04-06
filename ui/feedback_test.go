package ui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Since spinners and progress bars don't consistently write to stdout in tests,
// we'll focus on testing the configuration rather than the actual output

func TestNewScanner(t *testing.T) {
	// Create a spinner
	spinner := NewScanner()
	
	// Verify spinner is properly configured
	assert.NotNil(t, spinner, "Spinner should not be nil")
	assert.Contains(t, spinner.suffix, "Scanning")
	assert.Contains(t, spinner.finalMsg, "Scan complete")
	assert.NotNil(t, spinner.spinner, "Underlying spinner should be initialized")
}

func TestNewProcessor(t *testing.T) {
	// Create a progress bar for 10 items
	bar := NewProcessor(10)
	
	// Verify progress bar is properly configured
	assert.NotNil(t, bar, "Progress bar should not be nil")
	assert.Equal(t, 10, bar.total)
	assert.NotNil(t, bar.bar, "Underlying progress bar should be initialized")
}

func TestSpinnerStartStop(t *testing.T) {
	// Create a spinner
	spinner := NewCustomSpinner(WithSuffix("Testing"), WithFinalMessage("Done testing"))
	
	// Just test that Start and Stop don't panic
	spinner.Start()
	time.Sleep(10 * time.Millisecond) // Give it time to spin
	spinner.Stop()
	
	// Success if we get here without panicking
	assert.True(t, true)
}

func TestProgressBarIncrement(t *testing.T) {
	// Create a progress bar with 3 items
	bar := NewProcessor(3)
	
	// Just test that increments don't panic
	bar.Increment()
	bar.Increment()
	bar.Increment()
	bar.Finish()
	
	// Success if we get here without panicking
	assert.True(t, true)
}

func TestCustomSpinnerOptions(t *testing.T) {
	// Test that options properly modify the spinner
	spinner := NewCustomSpinner(
		WithSuffix("Custom suffix"),
		WithFinalMessage("Custom final message"),
		WithCharset(1), // Different charset
		WithSpeed(100 * time.Millisecond),
	)
	
	assert.Equal(t, "Custom suffix", spinner.suffix)
	assert.Equal(t, "Custom final message", spinner.finalMsg)
	assert.Equal(t, 100*time.Millisecond, spinner.speed)
}

func TestCustomProgressBarOptions(t *testing.T) {
	// Test that options properly modify the progress bar
	bar := NewCustomProgressBar(5,
		WithDescription("Custom description"),
		WithWidth(50),
		WithTheme(ProgressBarTheme{
			Saucer:        "X",
			SaucerPadding: ".",
			BarStart:      "(",
			BarEnd:        ")",
		}),
	)
	
	assert.Equal(t, 5, bar.total)
	assert.Equal(t, "Custom description", bar.description)
	assert.Equal(t, 50, bar.width)
	assert.Equal(t, "X", bar.theme.Saucer)
}