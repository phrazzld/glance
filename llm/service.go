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
	promptTemplate string
}

// ServiceConfig contains configuration for creating a new Service.
// This simplifies the pattern for service creation while maintaining flexibility.
type ServiceConfig struct {
	// MaxRetries is the number of times to retry failed LLM operations
	MaxRetries int

	// ModelName is the name of the LLM model to use
	ModelName string

	// PromptTemplate is the template string to use for generating prompts
	PromptTemplate string
}

// DefaultServiceConfig returns a ServiceConfig with sensible defaults.
// It uses the same default model as the client configuration.
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		MaxRetries:     3,
		ModelName:      "gemini-2.0-flash", // Make sure this matches the client default
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

	// Log start of prompt generation with structured fields
	logrus.WithFields(logrus.Fields{
		"directory":  dir,
		"model":      s.modelName,
		"operation":  "generate_prompt",
		"file_count": len(fileMap),
	}).Debug("Generating prompt from template")

	// Use template from the service
	prompt, err := GeneratePrompt(promptData, s.promptTemplate)
	if err != nil {
		// Log prompt generation error with structured fields
		logrus.WithFields(logrus.Fields{
			"directory": dir,
			"model":     s.modelName,
			"operation": "generate_prompt",
			"error":     err,
			"status":    "failed",
		}).Error("Failed to generate prompt from template")
		return "", fmt.Errorf("failed to generate prompt: %w", err)
	}

	// Optional token counting for debugging
	tokens, tokenErr := s.client.CountTokens(ctx, prompt)
	if tokenErr == nil {
		logrus.WithFields(logrus.Fields{
			"directory":   dir,
			"token_count": tokens,
			"model":       s.modelName,
			"operation":   "count_tokens",
		}).Debug("Token count for prompt")
	} else {
		logrus.WithFields(logrus.Fields{
			"directory": dir,
			"model":     s.modelName,
			"operation": "count_tokens",
			"error":     tokenErr,
		}).Debug("Failed to count tokens")
	}

	// Attempt generation with retries
	var lastError error
	maxAttempts := s.maxRetries + 1 // +1 for the initial attempt

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			logrus.WithFields(logrus.Fields{
				"directory":    dir,
				"retry_number": attempt - 1,
				"max_retries":  s.maxRetries,
				"model":        s.modelName,
				"operation":    "generate_content",
			}).Debug("Retrying content generation")
		} else {
			logrus.WithFields(logrus.Fields{
				"directory": dir,
				"model":     s.modelName,
				"operation": "generate_content",
			}).Debug("Generating content")
		}

		// Generate content
		result, err := s.client.Generate(ctx, prompt)
		if err == nil {
			// Success
			logrus.WithFields(logrus.Fields{
				"directory": dir,
				"model":     s.modelName,
				"operation": "generate_content",
				"attempts":  attempt,
				"status":    "success",
			}).Debug("Content generation successful")
			return result, nil
		}

		// Log error and retry
		lastError = err
		logrus.WithFields(logrus.Fields{
			"directory": dir,
			"attempt":   attempt,
			"model":     s.modelName,
			"operation": "generate_content",
			"error":     err,
			"status":    "failed",
		}).Debug("Content generation attempt failed")

		// Simple backoff before retry
		if attempt < maxAttempts {
			backoffMs := 100 * attempt * attempt
			logrus.WithFields(logrus.Fields{
				"directory":  dir,
				"backoff_ms": backoffMs,
			}).Debug("Applying backoff before retry")
			time.Sleep(time.Duration(backoffMs) * time.Millisecond)
		}
	}

	// Log final error with structured fields
	logrus.WithFields(logrus.Fields{
		"directory":    dir,
		"max_attempts": maxAttempts,
		"model":        s.modelName,
		"operation":    "generate_content",
		"error":        lastError,
		"status":       "failed",
	}).Error("Content generation failed after all retry attempts")

	return "", fmt.Errorf("failed to generate content after %d attempts: %w",
		s.maxRetries, lastError)
}
