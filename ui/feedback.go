// Package ui provides user interface functionality for the glance application.
package ui

import (
	"time"

	"github.com/briandowns/spinner"
	"github.com/schollz/progressbar/v3"
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
		WithSuffix("🔍 Scanning directories and loading .gitignore files..."),
		WithFinalMessage("🎉 Scan complete! Found all the good stuff!\n"),
	)
}

// NewGenerator creates a spinner specifically for content generation operations.
func NewGenerator() *Spinner {
	return NewCustomSpinner(
		WithSuffix("🧠 Generating content..."),
		WithFinalMessage("✅ Generation complete!\n"),
	)
}

// -----------------------------------------------------------------------------
// Progress Bar
// -----------------------------------------------------------------------------

// ProgressBarTheme defines the visual appearance of a progress bar.
type ProgressBarTheme struct {
	Saucer        string
	SaucerPadding string
	BarStart      string
	BarEnd        string
}

// DefaultTheme provides the default appearance for progress bars.
var DefaultTheme = ProgressBarTheme{
	Saucer:        "█",
	SaucerPadding: "░",
	BarStart:      "[",
	BarEnd:        "]",
}

// ProgressBar represents a terminal progress bar for visual feedback.
type ProgressBar struct {
	bar         *progressbar.ProgressBar
	total       int
	description string
	width       int
	theme       ProgressBarTheme
}

// Increment advances the progress bar by one step.
func (p *ProgressBar) Increment() error {
	return p.bar.Add(1)
}

// Set sets the progress bar to a specific value.
func (p *ProgressBar) Set(value int) error {
	return p.bar.Set(value)
}

// Finish completes the progress bar.
func (p *ProgressBar) Finish() error {
	return p.bar.Finish()
}

// ProgressBarOption is a function type that configures a ProgressBar.
type ProgressBarOption func(*ProgressBar)

// WithDescription sets the text displayed alongside the progress bar.
func WithDescription(description string) ProgressBarOption {
	return func(p *ProgressBar) {
		p.description = description
	}
}

// WithWidth sets the width of the progress bar.
func WithWidth(width int) ProgressBarOption {
	return func(p *ProgressBar) {
		p.width = width
	}
}

// WithTheme sets the visual theme of the progress bar.
func WithTheme(theme ProgressBarTheme) ProgressBarOption {
	return func(p *ProgressBar) {
		p.theme = theme
	}
}

// NewCustomProgressBar creates a new progress bar with custom options.
func NewCustomProgressBar(total int, options ...ProgressBarOption) *ProgressBar {
	// Default values
	p := &ProgressBar{
		total:       total,
		description: "Processing",
		width:       40,
		theme:       DefaultTheme,
	}
	
	// Apply custom options
	for _, option := range options {
		option(p)
	}
	
	// Create the underlying progress bar
	p.bar = progressbar.NewOptions(total,
		progressbar.OptionSetDescription(p.description),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(p.width),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        p.theme.Saucer,
			SaucerPadding: p.theme.SaucerPadding,
			BarStart:      p.theme.BarStart,
			BarEnd:        p.theme.BarEnd,
		}),
	)
	
	return p
}

// NewProcessor creates a progress bar for processing a known number of items.
func NewProcessor(total int) *ProgressBar {
	return NewCustomProgressBar(total,
		WithDescription("✍️ Creating GLANCE files"),
		WithWidth(40),
		WithTheme(DefaultTheme),
	)
}

// NewFileProcessor creates a progress bar specifically for file processing operations.
func NewFileProcessor(total int) *ProgressBar {
	return NewCustomProgressBar(total,
		WithDescription("📄 Processing files"),
		WithWidth(40),
		WithTheme(DefaultTheme),
	)
}

// -----------------------------------------------------------------------------
// Error Reporting
// -----------------------------------------------------------------------------

// ReportError logs an error and optionally displays it to the user.
func ReportError(err error, verbose bool, context string) {
	if err == nil {
		return
	}
	
	if verbose {
		logrus.Errorf("❌ %s: %v", context, err)
	} else {
		logrus.Errorf("❌ %s", context)
	}
}