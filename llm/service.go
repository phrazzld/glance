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
	client         Client
	maxRetries     int
	modelName      string
	verbose        bool
	promptTemplate string
}

// ServiceConfig contains configuration for creating a new Service.
// This simplifies the pattern for service creation while maintaining flexibility.
type ServiceConfig struct {
	// MaxRetries is the number of times to retry failed LLM operations
	MaxRetries int

	// ModelName is the name of the LLM model to use
	ModelName string

	// Verbose enables detailed logging for LLM operations
	Verbose bool

	// PromptTemplate is the template string to use for generating prompts
	PromptTemplate string
}

// DefaultServiceConfig returns a ServiceConfig with sensible defaults.
// It uses the same default model as the client configuration.
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		MaxRetries:     3,
		ModelName:      "gemini-2.0-flash", // Make sure this matches the client default
		Verbose:        false,
		PromptTemplate: "",
	}
}

// WithServiceMaxRetries configures the maximum number of retries for the service.
func WithServiceMaxRetries(maxRetries int) func(*ServiceConfig) {
	return func(c *ServiceConfig) {
		c.MaxRetries = maxRetries
	}
}

// WithServiceModelName configures the model name for the service.
func WithServiceModelName(modelName string) func(*ServiceConfig) {
	return func(c *ServiceConfig) {
		c.ModelName = modelName
	}
}

// WithVerbose configures verbose logging for the service.
func WithVerbose(verbose bool) func(*ServiceConfig) {
	return func(c *ServiceConfig) {
		c.Verbose = verbose
	}
}

// WithPromptTemplate configures the prompt template for the service.
func WithPromptTemplate(template string) func(*ServiceConfig) {
	return func(c *ServiceConfig) {
		c.PromptTemplate = template
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
func NewService(client Client, options ...func(*ServiceConfig)) (*Service, error) {
	if client == nil {
		return nil, fmt.Errorf("client cannot be nil")
	}

	// Start with default config
	config := DefaultServiceConfig()

	// Apply any provided options
	for _, option := range options {
		option(&config)
	}

	return &Service{
		client:         client,
		maxRetries:     config.MaxRetries,
		modelName:      config.ModelName,
		verbose:        config.Verbose,
		promptTemplate: config.PromptTemplate,
	}, nil
}

// GenerateGlanceMarkdown generates a markdown summary for a directory using the LLM.
// It builds a prompt based on directory information, sends it to the LLM client,
// and handles retries with exponential backoff if necessary.
//
// Parameters:
//   - ctx: The context for the operation
//   - dir: The directory path being processed
//   - fileMap: A map of file names to their contents
//   - subGlances: The combined contents of subdirectory glance.md files
//
// Returns:
//   - The generated markdown content
//   - An error if generation fails after all retries
func (s *Service) GenerateGlanceMarkdown(ctx context.Context, dir string, fileMap map[string]string, subGlances string) (string, error) {
	// Build prompt data
	promptData := BuildPromptData(dir, subGlances, fileMap)

	// Use template from the service
	prompt, err := GeneratePrompt(promptData, s.promptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to generate prompt: %w", err)
	}

	// Optional token counting for debugging
	if s.verbose {
		tokens, tokenErr := s.client.CountTokens(ctx, prompt)
		if tokenErr == nil {
			logrus.Debugf("🔤 Token count for %s: %d tokens in prompt", dir, tokens)
		} else {
			logrus.Debugf("⚠️ Couldn't count tokens for %s: %v", dir, tokenErr)
		}
	}

	// Attempt generation with retries
	var lastError error
	maxAttempts := s.maxRetries + 1 // +1 for the initial attempt

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if s.verbose {
			if attempt > 1 {
				logrus.Debugf("🔄 Retry #%d/%d for %s", attempt-1, s.maxRetries, dir)
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
		if s.verbose {
			logrus.Debugf("❌ Attempt %d for %s failed: %v", attempt, dir, err)
		}

		// Simple backoff before retry
		if attempt < maxAttempts {
			backoffMs := 100 * attempt * attempt
			time.Sleep(time.Duration(backoffMs) * time.Millisecond)
		}
	}

	return "", fmt.Errorf("failed to generate content after %d attempts: %w",
		s.maxRetries, lastError)
}
