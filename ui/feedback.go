// Package ui provides user interface functionality for the glance application.
package ui

import (
	"time"

	"github.com/briandowns/spinner"
	"github.com/sirupsen/logrus"
)

// -----------------------------------------------------------------------------
// Spinner
// -----------------------------------------------------------------------------

// Spinner represents a terminal spinner for visual feedback during operations
// that don't have easily quantifiable progress.
type Spinner struct {
	spinner  *spinner.Spinner
	suffix   string
	finalMsg string
	speed    time.Duration
}

// Start activates the spinner animation.
func (s *Spinner) Start() {
	s.spinner.Start()
}

// Stop halts the spinner animation and displays the final message.
func (s *Spinner) Stop() {
	s.spinner.FinalMSG = s.finalMsg
	s.spinner.Stop()
}

// UpdateMessage changes the message displayed alongside the spinner.
func (s *Spinner) UpdateMessage(message string) {
	s.spinner.Suffix = " " + message
}

// SpinnerOption is a function type that configures a Spinner.
type SpinnerOption func(*Spinner)

// WithSuffix sets the text displayed after the spinner.
func WithSuffix(suffix string) SpinnerOption {
	return func(s *Spinner) {
		s.suffix = suffix
		s.spinner.Suffix = " " + suffix
	}
}

// WithFinalMessage sets the text displayed when the spinner stops.
func WithFinalMessage(message string) SpinnerOption {
	return func(s *Spinner) {
		s.finalMsg = message
	}
}

// WithCharset sets the spinner's animation character set.
func WithCharset(charset int) SpinnerOption {
	return func(s *Spinner) {
		if charset >= 0 && charset < len(spinner.CharSets) {
			s.spinner.UpdateCharSet(spinner.CharSets[charset])
		}
	}
}

// WithSpeed sets the speed of the spinner animation.
func WithSpeed(speed time.Duration) SpinnerOption {
	return func(s *Spinner) {
		s.speed = speed
		s.spinner.UpdateSpeed(speed)
	}
}

// NewCustomSpinner creates a new spinner with custom options.
func NewCustomSpinner(options ...SpinnerOption) *Spinner {
	// Default values
	s := &Spinner{
		spinner:  spinner.New(spinner.CharSets[14], 120*time.Millisecond),
		suffix:   "Processing...",
		finalMsg: "Done!\n",
		speed:    120 * time.Millisecond,
	}

	// Apply suffix to the spinner
	s.spinner.Suffix = " " + s.suffix

	// Apply custom options
	for _, option := range options {
		option(s)
	}

	return s
}

// NewScanner creates a spinner specifically for directory scanning operations.
func NewScanner() *Spinner {
	return NewCustomSpinner(
		WithSuffix("Scanning directories and loading .gitignore files..."),
		WithFinalMessage("Scan complete!\n"),
	)
}

// NewGenerator creates a spinner specifically for content generation operations.
func NewGenerator() *Spinner {
	return NewCustomSpinner(
		WithSuffix("Generating content..."),
		WithFinalMessage("Generation complete!\n"),
	)
}

// -----------------------------------------------------------------------------
// Error Reporting
// -----------------------------------------------------------------------------

// ReportError logs an error and optionally displays it to the user.
func ReportError(err error, context string) {
	if err == nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"context": context,
		"error":   err,
	}).Error("Operation failed")
}
