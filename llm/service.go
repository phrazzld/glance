// Package llm provides abstractions and implementations for interacting with
// Large Language Model APIs in the glance application.
package llm

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// Service provides high-level LLM operations for the Glance application.
// It encapsulates a Client and provides application-specific functionality
// for generating directory summaries.
type Service struct {
	client         Client
	modelName      string
	promptTemplate string
}

// ServiceConfig contains configuration for creating a new Service.
// This simplifies the pattern for service creation while maintaining flexibility.
type ServiceConfig struct {
	// ModelName is the name of the LLM model to use
	ModelName string

	// PromptTemplate is the template string to use for generating prompts
	PromptTemplate string
}

// DefaultServiceConfig returns a ServiceConfig with sensible defaults.
// It uses the same default model as the client configuration.
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		ModelName:      "gemini-3-flash-preview", // Make sure this matches the client default
		PromptTemplate: "",
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
		modelName:      config.ModelName,
		promptTemplate: config.PromptTemplate,
	}, nil
}

// GenerateGlanceMarkdown generates a markdown summary for a directory using the LLM.
// It builds a prompt based on directory information, sends it to the LLM client,
// and returns the generated markdown.
//
// Parameters:
//   - ctx: The context for the operation
//   - dir: The directory path being processed
//   - fileMap: A map of file names to their contents
//   - subGlances: The combined contents of subdirectory glance.md files
//
// Returns:
//   - The generated markdown content
//   - An error if generation fails
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

	logrus.WithFields(logrus.Fields{
		"directory": dir,
		"model":     s.modelName,
		"operation": "generate_content",
	}).Debug("Generating content")

	result, err := s.client.Generate(ctx, prompt)
	if err == nil {
		logrus.WithFields(logrus.Fields{
			"directory": dir,
			"model":     s.modelName,
			"operation": "generate_content",
			"status":    "success",
		}).Debug("Content generation successful")
		return result, nil
	}

	logrus.WithFields(logrus.Fields{
		"directory": dir,
		"model":     s.modelName,
		"operation": "generate_content",
		"error":     err,
		"status":    "failed",
	}).Error("Content generation failed")

	return "", fmt.Errorf("failed to generate content: %w", err)
}
