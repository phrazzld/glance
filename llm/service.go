// Package llm provides abstractions and implementations for interacting with
// Large Language Model APIs in the glance application.
package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// Service provides high-level LLM operations for the Glance application.
// It encapsulates a Client and provides application-specific functionality
// for generating directory summaries.
type Service struct {
	client  Client
	options *ServiceOptions
}

// ServiceOptions configures the behavior of the LLM Service.
type ServiceOptions struct {
	// MaxRetries is the number of times to retry failed LLM operations
	MaxRetries int

	// ModelName is the name of the LLM model to use
	ModelName string

	// Verbose enables detailed logging for LLM operations
	Verbose bool
}

// WithMaxRetries returns a new ServiceOptions with the specified max retries value.
func (o *ServiceOptions) WithMaxRetries(maxRetries int) *ServiceOptions {
	newOpts := *o
	newOpts.MaxRetries = maxRetries
	return &newOpts
}

// WithModelName returns a new ServiceOptions with the specified model name.
func (o *ServiceOptions) WithModelName(modelName string) *ServiceOptions {
	newOpts := *o
	newOpts.ModelName = modelName
	return &newOpts
}

// WithVerbose returns a new ServiceOptions with the specified verbose setting.
func (o *ServiceOptions) WithVerbose(verbose bool) *ServiceOptions {
	newOpts := *o
	newOpts.Verbose = verbose
	return &newOpts
}

// DefaultServiceOptions returns a ServiceOptions instance with sensible defaults.
func DefaultServiceOptions() *ServiceOptions {
	return &ServiceOptions{
		MaxRetries: 3,
		ModelName:  "gemini-2.0-flash",
		Verbose:    false,
	}
}

// ServiceOption is a function type for applying options to a Service.
type ServiceOption func(*ServiceOptions)

// WithMaxRetries configures the maximum number of retries for the service.
func WithMaxRetries(maxRetries int) ServiceOption {
	return func(o *ServiceOptions) {
		o.MaxRetries = maxRetries
	}
}

// WithModelName configures the model name for the service.
func WithModelName(modelName string) ServiceOption {
	return func(o *ServiceOptions) {
		o.ModelName = modelName
	}
}

// WithVerbose configures verbose logging for the service.
func WithVerbose(verbose bool) ServiceOption {
	return func(o *ServiceOptions) {
		o.Verbose = verbose
	}
}

// NewService creates a new LLM Service with the specified client and options.
//
// Parameters:
//   - client: The LLM client to use for API interactions
//   - options: Optional functional options to configure the service
//
// Returns:
//   - A new Service instance
//   - An error if service creation fails
func NewService(client Client, options ...ServiceOption) (*Service, error) {
	if client == nil {
		return nil, fmt.Errorf("client cannot be nil")
	}

	// Start with default options
	serviceOptions := DefaultServiceOptions()

	// Apply any provided options
	for _, option := range options {
		option(serviceOptions)
	}

	return &Service{
		client:  client,
		options: serviceOptions,
	}, nil
}

// GenerateGlanceMarkdown generates a markdown summary for a directory using the LLM.
//
// Parameters:
//   - ctx: The context for the operation
//   - dir: The directory path being processed
//   - fileMap: A map of file names to their contents
//   - subGlances: The combined contents of subdirectory GLANCE.md files
//
// Returns:
//   - The generated markdown content
//   - An error if generation fails after all retries
func (s *Service) GenerateGlanceMarkdown(ctx context.Context, dir string, fileMap map[string]string, subGlances string) (string, error) {
	// Build prompt data
	promptData := BuildPromptData(dir, subGlances, fileMap)

	// Get template and generate prompt
	template, err := LoadTemplate("")
	if err != nil {
		return "", fmt.Errorf("failed to load template: %w", err)
	}

	prompt, err := GeneratePrompt(promptData, template)
	if err != nil {
		return "", fmt.Errorf("failed to generate prompt: %w", err)
	}

	// Optional token counting for debugging
	if s.options.Verbose {
		tokens, tokenErr := s.client.CountTokens(ctx, prompt)
		if tokenErr == nil {
			logrus.Debugf("🔤 Token count for %s: %d tokens in prompt", dir, tokens)
		} else {
			logrus.Debugf("⚠️ Couldn't count tokens for %s: %v", dir, tokenErr)
		}
	}

	// Attempt generation with retries
	var lastError error
	maxAttempts := s.options.MaxRetries + 1 // +1 for the initial attempt

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if s.options.Verbose {
			if attempt > 1 {
				logrus.Debugf("🔄 Retry #%d/%d for %s", attempt-1, s.options.MaxRetries, dir)
			} else {
				logrus.Debugf("🚀 Generating content for %s", dir)
			}
		}

		// Generate content
		result, err := s.client.Generate(ctx, prompt)
		if err == nil {
			// Success
			return result, nil
		}

		// Log error and retry
		lastError = err
		if s.options.Verbose {
			logrus.Debugf("❌ Attempt %d for %s failed: %v", attempt, dir, err)
		}

		// Simple backoff before retry
		if attempt < maxAttempts {
			backoffMs := 100 * attempt * attempt
			time.Sleep(time.Duration(backoffMs) * time.Millisecond)
		}
	}

	return "", fmt.Errorf("failed to generate content after %d attempts: %w",
		s.options.MaxRetries, lastError)
}
