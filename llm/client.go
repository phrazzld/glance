// Package llm provides abstractions and implementations for interacting with
// Large Language Model APIs in the glance application.
package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/genai"
)

// Client defines the interface for interacting with LLM services.
// This interface abstracts the underlying LLM provider, making it easier
// to switch providers or mock in tests.
type Client interface {
	// Generate takes a prompt and returns the generated text.
	// It handles all API interaction details and returns only the final result.
	Generate(ctx context.Context, prompt string) (string, error)

	// CountTokens counts the number of tokens in the provided prompt.
	// This is useful for understanding API usage and costs.
	CountTokens(ctx context.Context, prompt string) (int, error)

	// Close releases any resources used by the client.
	// It should be called when the client is no longer needed.
	Close()
}

// ClientOptions holds configuration options for LLM clients.
// It allows customizing client behavior while providing sensible defaults.
type ClientOptions struct {
	// ModelName is the name of the model to use (e.g., "gemini-2.0-flash")
	ModelName string

	// MaxRetries is the number of times to retry failed API calls
	MaxRetries int

	// Timeout is the maximum time in seconds to wait for API responses
	Timeout int
}

// DefaultClientOptions returns a ClientOptions instance with sensible defaults.
func DefaultClientOptions() ClientOptions {
	return ClientOptions{
		ModelName:  "gemini-2.5-flash-preview-04-17",
		MaxRetries: 3,
		Timeout:    60, // 60 seconds
	}
}

// ClientOption is a function type for applying options to ClientOptions.
type ClientOption func(*ClientOptions)

// WithModelName sets the model name for the client.
func WithModelName(modelName string) ClientOption {
	return func(o *ClientOptions) {
		o.ModelName = modelName
	}
}

// WithMaxRetries sets the maximum number of retries for the client.
func WithMaxRetries(maxRetries int) ClientOption {
	return func(o *ClientOptions) {
		o.MaxRetries = maxRetries
	}
}

// WithTimeout sets the timeout in seconds for the client.
func WithTimeout(timeout int) ClientOption {
	return func(o *ClientOptions) {
		o.Timeout = timeout
	}
}

// GeminiClient is a Client implementation that uses Google's Gemini API.
type GeminiClient struct {
	client  *genai.Client
	model   string
	options *ClientOptions
}

// NewGeminiClientFunc is a function type for creating LLM clients.
// This allows for replacing the implementation in tests without the full factory interface.
type NewGeminiClientFunc func(apiKey string, options ...ClientOption) (Client, error)

// The actual implementation function - can be swapped in tests
var createGeminiClient NewGeminiClientFunc = func(apiKey string, options ...ClientOption) (Client, error) {
	return newGeminiClient(apiKey, options...)
}

// NewGeminiClient creates a new client for the Google Gemini API.
// Tests can replace createGeminiClient to return mock implementations.
func NewGeminiClient(apiKey string, options ...ClientOption) (Client, error) {
	return createGeminiClient(apiKey, options...)
}

// newGeminiClient is the actual implementation for creating a new client for the Google Gemini API.
//
// Parameters:
//   - apiKey: The API key for authenticating with the Gemini API // #nosec G101 // pragma: allowlist secret
//   - options: Zero or more functional options to configure the client
//
// Returns:
//   - A new GeminiClient instance
//   - An error if client creation fails
func newGeminiClient(apiKey string, options ...ClientOption) (*GeminiClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Start with default options
	opts := DefaultClientOptions()

	// Apply any provided options
	for _, option := range options {
		option(&opts)
	}

	ctx := context.Background()
	// #nosec G101 -- API key is provided by the user and not hardcoded // pragma: allowlist secret
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey, // pragma: allowlist secret
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiClient{
		client:  client,
		model:   opts.ModelName,
		options: &opts,
	}, nil
}

// Generate implements the Client interface for GeminiClient.
// It sends the prompt to the Gemini API and processes the response.
// This uses the non-streaming API for better efficiency with simple requests.
func (c *GeminiClient) Generate(ctx context.Context, prompt string) (string, error) {
	if c.client == nil || c.model == "" {
		return "", fmt.Errorf("client is not properly initialized")
	}

	// Create a context with timeout if specified
	var genCtx context.Context
	var cancel context.CancelFunc
	if c.options.Timeout > 0 {
		genCtx, cancel = context.WithTimeout(ctx, time.Duration(c.options.Timeout)*time.Second)
		defer cancel()
	} else {
		genCtx = ctx
	}

	var lastError error

	// Prepare the content for the request
	contents := []*genai.Content{
		genai.NewContentFromText(prompt, "user"),
	}

	// Retry logic
	for attempt := 1; attempt <= c.options.MaxRetries; attempt++ {
		if attempt > 1 {
			logrus.Debugf("Retry attempt %d/%d for generating content", attempt, c.options.MaxRetries)
		}

		// Use non-streaming API for better efficiency
		resp, err := c.client.Models.GenerateContent(genCtx, c.model, contents, nil)
		if err != nil {
			lastError = fmt.Errorf("failed to generate content: %w", err)
			// Simple backoff before retry
			if attempt < c.options.MaxRetries {
				backoffMs := 100 * attempt * attempt // Exponential backoff
				time.Sleep(time.Duration(backoffMs) * time.Millisecond)
			}
			continue
		}

		// Check if we have valid candidates
		if resp == nil || len(resp.Candidates) == 0 {
			lastError = fmt.Errorf("received empty response")
			// Simple backoff before retry
			if attempt < c.options.MaxRetries {
				backoffMs := 100 * attempt * attempt // Exponential backoff
				time.Sleep(time.Duration(backoffMs) * time.Millisecond)
			}
			continue
		}

		// Check for finish reason issues
		if resp.Candidates[0].FinishReason != "FINISHED" {
			// Handle various non-success finish reasons
			reason := resp.Candidates[0].FinishReason
			if reason == "SAFETY" {
				lastError = fmt.Errorf("content blocked by safety settings")
			} else {
				lastError = fmt.Errorf("generation incomplete: %s", reason)
			}
			// Simple backoff before retry
			if attempt < c.options.MaxRetries {
				backoffMs := 100 * attempt * attempt // Exponential backoff
				time.Sleep(time.Duration(backoffMs) * time.Millisecond)
			}
			continue
		}

		// Extract text from the response
		var result strings.Builder
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				result.WriteString(part.Text)
			}
		}

		return result.String(), nil
	}

	return "", fmt.Errorf("failed to generate content after %d attempts: %w",
		c.options.MaxRetries, lastError)
}

// CountTokens implements the Client interface for GeminiClient.
// It counts the number of tokens in the provided prompt.
func (c *GeminiClient) CountTokens(ctx context.Context, prompt string) (int, error) {
	if c.client == nil || c.model == "" {
		return 0, fmt.Errorf("client is not properly initialized")
	}

	// Create a context with timeout if specified
	var tokenCtx context.Context
	var cancel context.CancelFunc
	if c.options.Timeout > 0 {
		tokenCtx, cancel = context.WithTimeout(ctx, time.Duration(c.options.Timeout)*time.Second)
		defer cancel()
	} else {
		tokenCtx = ctx
	}

	var lastError error

	// Prepare the content for the token count
	contents := []*genai.Content{
		genai.NewContentFromText(prompt, "user"),
	}

	// Retry logic
	for attempt := 1; attempt <= c.options.MaxRetries; attempt++ {
		if attempt > 1 {
			logrus.Debugf("Retry attempt %d/%d for counting tokens", attempt, c.options.MaxRetries)
		}

		response, err := c.client.Models.CountTokens(tokenCtx, c.model, contents, nil)
		if err == nil {
			// Convert int32 to int
			return int(response.TotalTokens), nil
		}

		lastError = err

		// Simple backoff before retry
		if attempt < c.options.MaxRetries {
			backoffMs := 100 * attempt * attempt // Exponential backoff
			time.Sleep(time.Duration(backoffMs) * time.Millisecond)
		}
	}

	return 0, fmt.Errorf("failed to count tokens after %d attempts: %w",
		c.options.MaxRetries, lastError)
}

// Close implements the Client interface for GeminiClient.
// It releases resources used by the client.
func (c *GeminiClient) Close() {
	if c.client != nil {
		// The new Google GenAI client doesn't require explicit closing
		c.client = nil
		c.model = ""
	}
}
