// Package llm provides abstractions and implementations for interacting with
// Large Language Model APIs in the glance application.
package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
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
func DefaultClientOptions() *ClientOptions {
	return &ClientOptions{
		ModelName:  "gemini-2.5-flash-preview-04-17",
		MaxRetries: 3,
		Timeout:    60, // 60 seconds
	}
}

// WithModelName returns a new ClientOptions with the specified model name.
func (o *ClientOptions) WithModelName(modelName string) *ClientOptions {
	newOpts := *o
	newOpts.ModelName = modelName
	return &newOpts
}

// WithMaxRetries returns a new ClientOptions with the specified max retries value.
func (o *ClientOptions) WithMaxRetries(maxRetries int) *ClientOptions {
	newOpts := *o
	newOpts.MaxRetries = maxRetries
	return &newOpts
}

// WithTimeout returns a new ClientOptions with the specified timeout in seconds.
func (o *ClientOptions) WithTimeout(timeout int) *ClientOptions {
	newOpts := *o
	newOpts.Timeout = timeout
	return &newOpts
}

// GeminiClient is a Client implementation that uses Google's Gemini API.
type GeminiClient struct {
	client  *genai.Client
	model   *genai.GenerativeModel
	options *ClientOptions
}

// NewGeminiClient creates a new client for the Google Gemini API.
//
// Parameters:
//   - apiKey: The API key for authenticating with the Gemini API  // pragma: allowlist secret
//   - options: Configuration options for the client
//
// Returns:
//   - A new GeminiClient instance
//   - An error if client creation fails
func NewGeminiClient(apiKey string, options *ClientOptions) (*GeminiClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	if options == nil {
		options = DefaultClientOptions()
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel(options.ModelName)

	return &GeminiClient{
		client:  client,
		model:   model,
		options: options,
	}, nil
}

// Generate implements the Client interface for GeminiClient.
// It sends the prompt to the Gemini API and processes the streaming response.
func (c *GeminiClient) Generate(ctx context.Context, prompt string) (string, error) {
	if c.client == nil || c.model == nil {
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

	var result strings.Builder
	var lastError error

	// Retry logic
	for attempt := 1; attempt <= c.options.MaxRetries; attempt++ {
		if attempt > 1 {
			logrus.Debugf("Retry attempt %d/%d for generating content", attempt, c.options.MaxRetries)
		}

		result.Reset()
		stream := c.model.GenerateContentStream(genCtx, genai.Text(prompt))

		// Process the streaming response
		success := true
		for {
			resp, err := stream.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				lastError = err
				logrus.Debugf("Error in stream response: %v", err)
				success = false
				break
			}

			// Process candidates from the response
			for _, candidate := range resp.Candidates {
				if candidate.Content == nil {
					continue
				}
				for _, part := range candidate.Content.Parts {
					if txt, ok := part.(genai.Text); ok {
						result.WriteString(string(txt))
					}
				}
			}
		}

		if success {
			return result.String(), nil
		}

		// Simple backoff before retry
		if attempt < c.options.MaxRetries {
			backoffMs := 100 * attempt * attempt // Exponential backoff
			time.Sleep(time.Duration(backoffMs) * time.Millisecond)
		}
	}

	return "", fmt.Errorf("failed to generate content after %d attempts: %w",
		c.options.MaxRetries, lastError)
}

// CountTokens implements the Client interface for GeminiClient.
// It counts the number of tokens in the provided prompt.
func (c *GeminiClient) CountTokens(ctx context.Context, prompt string) (int, error) {
	if c.client == nil || c.model == nil {
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

	// Retry logic
	for attempt := 1; attempt <= c.options.MaxRetries; attempt++ {
		if attempt > 1 {
			logrus.Debugf("Retry attempt %d/%d for counting tokens", attempt, c.options.MaxRetries)
		}

		tokenResp, err := c.model.CountTokens(tokenCtx, genai.Text(prompt))
		if err == nil {
			return int(tokenResp.TotalTokens), nil
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
		c.client.Close()
		c.client = nil
		c.model = nil
	}
}
