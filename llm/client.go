// Package llm provides abstractions and implementations for interacting with
// Large Language Model APIs in the glance application.
package llm

import (
	"context"
	"errors" // For errors.Is
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/genai"

	customerrors "glance/errors" // Application's custom error package
)

// Client defines the interface for interacting with LLM services.
// This interface abstracts the underlying LLM provider, making it easier
// to switch providers or mock in tests.
type Client interface {
	// Generate takes a prompt and returns the generated text.
	// It handles all API interaction details and returns only the final result.
	Generate(ctx context.Context, prompt string) (string, error)

	// GenerateStream takes a prompt and returns a channel of generated text chunks.
	// It enables streaming responses from the LLM API for incremental processing.
	// Consumers must read from the channel until it's closed to avoid resource leaks.
	GenerateStream(ctx context.Context, prompt string) (<-chan StreamChunk, error)

	// CountTokens counts the number of tokens in the provided prompt.
	// This is useful for understanding API usage and costs.
	CountTokens(ctx context.Context, prompt string) (int, error)

	// Close releases any resources used by the client.
	// It should be called when the client is no longer needed.
	Close()
}

// StreamChunk represents a piece of content from a streaming LLM response.
// It contains either content text or an error encountered during streaming.
type StreamChunk struct {
	// Text contains the text content of this chunk, if any
	Text string

	// Error contains any error encountered during streaming
	Error error

	// Done indicates that this is the final chunk of the stream
	Done bool
}

// Safety thresholds for content filtering
const (
	// HarmBlockNone allows all content regardless of potential harm
	HarmBlockNone = "HARM_BLOCK_NONE"

	// HarmBlockLowAndAbove blocks content with low or higher likelihood of being harmful
	HarmBlockLowAndAbove = "HARM_BLOCK_LOW_AND_ABOVE"

	// HarmBlockMediumAndAbove blocks content with medium or higher likelihood of being harmful
	HarmBlockMediumAndAbove = "HARM_BLOCK_MEDIUM_AND_ABOVE"

	// HarmBlockHighAndAbove blocks only content with high likelihood of being harmful
	HarmBlockHighAndAbove = "HARM_BLOCK_HIGH_AND_ABOVE"

	// HarmBlockUnspecified uses the API's default blocking behavior
	HarmBlockUnspecified = "HARM_BLOCK_UNSPECIFIED"
)

// Safety categories for content filtering
const (
	// HarmCategoryHarassment represents content that harasses, intimidates, or bullies an individual or group
	HarmCategoryHarassment = "HARM_CATEGORY_HARASSMENT"

	// HarmCategoryHateSpeech represents content that expresses hatred toward identity attributes
	HarmCategoryHateSpeech = "HARM_CATEGORY_HATE_SPEECH"

	// HarmCategoryDangerousContent represents content that promotes dangerous or illegal activities
	HarmCategoryDangerousContent = "HARM_CATEGORY_DANGEROUS_CONTENT"

	// HarmCategorySexuallyExplicit represents content that contains sexual references
	HarmCategorySexuallyExplicit = "HARM_CATEGORY_SEXUALLY_EXPLICIT"

	// HarmCategoryDerogatory represents content that is rude, disrespectful, unreasonable or profane
	HarmCategoryDerogatory = "HARM_CATEGORY_DEROGATORY"
)

// SafetySetting represents a content filtering setting for a specific harm category
type SafetySetting struct {
	// Category is the harm category to filter
	Category string

	// Threshold is the blocking threshold to apply
	Threshold string
}

// ClientOptions holds configuration options for LLM clients.
// It allows customizing client behavior while providing sensible defaults.
type ClientOptions struct {
	// Basic client configuration
	// ModelName is the name of the model to use (e.g., "gemini-3-flash-preview")
	ModelName string

	// MaxRetries is the number of times to retry failed API calls
	MaxRetries int

	// Timeout is the maximum time in seconds to wait for API responses
	Timeout int

	// Generation parameters
	// Temperature controls the randomness of predictions (0.0 to 1.0)
	Temperature float32

	// TopP controls nucleus sampling (0.0 to 1.0)
	TopP float32

	// TopK controls the diversity of generated tokens (1.0 to max float32)
	TopK float32

	// MaxOutputTokens limits the length of the generated content
	MaxOutputTokens int32

	// CandidateCount is the number of response alternatives to generate
	CandidateCount int32

	// StopSequences are strings that stop generation if encountered
	StopSequences []string

	// SafetySettings are content filtering rules
	SafetySettings []*SafetySetting

	// SystemInstructions provide context or persona to the model
	SystemInstructions string
}

// DefaultClientOptions returns a ClientOptions instance with sensible defaults.
func DefaultClientOptions() ClientOptions {
	return ClientOptions{
		// Basic configuration
		ModelName:  "gemini-3-flash-preview",
		MaxRetries: 3,
		Timeout:    60, // 60 seconds

		// Generation parameters with reasonable defaults
		Temperature:     0.7,
		TopP:            0.95,
		TopK:            40,
		MaxOutputTokens: 2048,
		CandidateCount:  1,
		StopSequences:   []string{},
		SafetySettings:  []*SafetySetting{},
	}
}

// ClientOption is a function type for applying options to ClientOptions.
type ClientOption func(*ClientOptions)

// Basic configuration options

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

// Generation parameter options

// WithTemperature sets the temperature parameter for text generation.
// Values closer to 0 produce more predictable responses, while values
// closer to 1 produce more creative/varied responses. Valid range is 0.0 to 1.0.
func WithTemperature(temperature float32) ClientOption {
	return func(o *ClientOptions) {
		o.Temperature = temperature
	}
}

// WithTopP sets the nucleus sampling parameter for text generation.
// The model considers the smallest set of tokens whose cumulative probability
// exceeds topP. Valid range is 0.0 to 1.0.
func WithTopP(topP float32) ClientOption {
	return func(o *ClientOptions) {
		o.TopP = topP
	}
}

// WithTopK sets the top-k sampling parameter for text generation.
// The model considers only the top k most probable next tokens.
// Valid range is 1.0 to max float32 in the new API.
func WithTopK(topK float32) ClientOption {
	return func(o *ClientOptions) {
		o.TopK = topK
	}
}

// WithMaxOutputTokens sets the maximum number of tokens to generate.
// This limits the length of the response.
func WithMaxOutputTokens(maxOutputTokens int32) ClientOption {
	return func(o *ClientOptions) {
		o.MaxOutputTokens = maxOutputTokens
	}
}

// WithCandidateCount sets the number of candidate responses to generate.
// The API will return multiple alternative responses when this is > 1.
func WithCandidateCount(count int32) ClientOption {
	return func(o *ClientOptions) {
		o.CandidateCount = count
	}
}

// WithStopSequences sets sequences that will stop generation if encountered.
// These are strings that, if generated, will cause the model to stop.
func WithStopSequences(sequences []string) ClientOption {
	return func(o *ClientOptions) {
		o.StopSequences = sequences
	}
}

// WithSafetySetting adds a safety setting for content filtering.
// Each setting specifies a harm category and threshold for blocking content.
func WithSafetySetting(category, threshold string) ClientOption {
	return func(o *ClientOptions) {
		o.SafetySettings = append(o.SafetySettings, &SafetySetting{
			Category:  category,
			Threshold: threshold,
		})
	}
}

// WithSystemInstructions sets system instructions for the model.
// This provides context or persona guidance to the model.
func WithSystemInstructions(instructions string) ClientOption {
	return func(o *ClientOptions) {
		o.SystemInstructions = instructions
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
		return nil, customerrors.NewValidationError("API key is required", nil).
			WithCode("GENAI-001").
			WithSuggestion("Provide a valid API key either through environment variable or configuration")
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
		return nil, customerrors.WrapAPIError(err, "failed to create Gemini client").
			WithCode("GENAI-002").
			WithSuggestion("Check API key validity and network connectivity")
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
		return "", customerrors.NewValidationError("client is not properly initialized", nil).
			WithCode("GENAI-003").
			WithSuggestion("Ensure the client was created with a valid API key and model name")
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

	// Create generation config with our options
	genConfig := &genai.GenerateContentConfig{}

	// Apply generation parameters if they have non-zero values
	if c.options.Temperature > 0 {
		genConfig.Temperature = &c.options.Temperature
	}

	if c.options.TopP > 0 {
		genConfig.TopP = &c.options.TopP
	}

	if c.options.TopK > 0 {
		genConfig.TopK = &c.options.TopK
	}

	if c.options.MaxOutputTokens > 0 {
		genConfig.MaxOutputTokens = c.options.MaxOutputTokens
	}

	if c.options.CandidateCount > 0 {
		genConfig.CandidateCount = c.options.CandidateCount
	}

	if len(c.options.StopSequences) > 0 {
		genConfig.StopSequences = c.options.StopSequences
	}

	// Apply safety settings if any are defined
	if len(c.options.SafetySettings) > 0 {
		genConfig.SafetySettings = make([]*genai.SafetySetting, 0, len(c.options.SafetySettings))
		for _, ss := range c.options.SafetySettings {
			// Convert string category and threshold to genai enum types
			category := genai.HarmCategory(ss.Category)
			threshold := genai.HarmBlockThreshold(ss.Threshold)

			genConfig.SafetySettings = append(genConfig.SafetySettings, &genai.SafetySetting{
				Category:  category,
				Threshold: threshold,
			})
		}
	}

	// Prepare contents with system instructions if provided
	if c.options.SystemInstructions != "" {
		// Add system instructions as the first content item
		systemContent := genai.NewContentFromText(c.options.SystemInstructions, "system")
		contents = append([]*genai.Content{systemContent}, contents...)
	}

	// Retry logic
	for attempt := 1; attempt <= c.options.MaxRetries; attempt++ {
		if attempt > 1 {
			logrus.WithFields(logrus.Fields{
				"attempt":     attempt,
				"max_retries": c.options.MaxRetries,
			}).Debug("Retry attempt for generating content")
		}

		// Use non-streaming API with our configured generation options
		resp, err := c.client.Models.GenerateContent(genCtx, c.model, contents, genConfig)
		if err != nil {
			// Detect specific API error types
			lastError = customerrors.WrapAPIError(err, "failed to generate content").
				WithCode("GENAI-004")

			// Handle context deadline exceeded error specifically
			if errors.Is(err, context.DeadlineExceeded) {
				lastError = customerrors.WrapAPIError(err, "content generation timed out").
					WithCode("GENAI-005").
					WithSuggestion("Consider increasing the timeout value")
			}

			// Simple backoff before retry
			if attempt < c.options.MaxRetries {
				backoffMs := 100 * attempt * attempt // Exponential backoff
				time.Sleep(time.Duration(backoffMs) * time.Millisecond)
			}
			continue
		}

		// Check if we have valid candidates
		if resp == nil || len(resp.Candidates) == 0 {
			lastError = customerrors.NewAPIError("received empty response from API", nil).
				WithCode("GENAI-006").
				WithSuggestion("Check if the prompt contains content that may be filtered")

			// Simple backoff before retry
			if attempt < c.options.MaxRetries {
				backoffMs := 100 * attempt * attempt // Exponential backoff
				time.Sleep(time.Duration(backoffMs) * time.Millisecond)
			}
			continue
		}

		// Check for finish reason issues
		if resp.Candidates[0].FinishReason != "FINISHED" && resp.Candidates[0].FinishReason != "STOP" {
			// Handle various non-success finish reasons
			reason := resp.Candidates[0].FinishReason
			if reason == "SAFETY" {
				lastError = customerrors.NewAPIError("content blocked by safety settings", nil).
					WithCode("GENAI-007").
					WithSuggestion("Modify the prompt to avoid potentially harmful content")
			} else {
				lastError = customerrors.NewAPIError(fmt.Sprintf("generation incomplete: %s", reason), nil).
					WithCode("GENAI-008")
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

	// If we get here, all retries failed
	return "", customerrors.WrapAPIError(lastError, fmt.Sprintf("failed to generate content after %d attempts", c.options.MaxRetries)).
		WithCode("GENAI-009").
		WithSuggestion("Check internet connectivity, API key validity, and prompt content")
}

// CountTokens implements the Client interface for GeminiClient.
// It counts the number of tokens in the provided prompt using the google.golang.org/genai package.
//
// The method creates a genai.Content array from the prompt, then calls the CountTokens method
// on the genai.Models service. It includes retry logic with exponential backoff and honors
// the timeout set in ClientOptions.
//
// Returns:
//   - The total number of tokens in the prompt
//   - An error if the API call fails after all retries
func (c *GeminiClient) CountTokens(ctx context.Context, prompt string) (int, error) {
	if c.client == nil || c.model == "" {
		return 0, customerrors.NewValidationError("client is not properly initialized", nil).
			WithCode("GENAI-010").
			WithSuggestion("Ensure the client was created with a valid API key and model name")
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

	// Add system instructions to content if provided
	if c.options.SystemInstructions != "" {
		// Add system instructions as the first content item
		systemContent := genai.NewContentFromText(c.options.SystemInstructions, "system")
		contents = append([]*genai.Content{systemContent}, contents...)
	}

	// Create a config for counting tokens (has fewer options than generation)
	countConfig := &genai.CountTokensConfig{}

	// Retry logic
	for attempt := 1; attempt <= c.options.MaxRetries; attempt++ {
		if attempt > 1 {
			logrus.WithFields(logrus.Fields{
				"attempt":     attempt,
				"max_retries": c.options.MaxRetries,
			}).Debug("Retry attempt for counting tokens")
		}

		// Call the CountTokens API with the model name, contents, and our configuration
		response, err := c.client.Models.CountTokens(tokenCtx, c.model, contents, countConfig)
		if err == nil && response != nil {
			// Convert int32 to int and return the token count
			return int(response.TotalTokens), nil
		}

		// Handle specific error cases
		if err == nil && response == nil {
			lastError = customerrors.NewAPIError("received nil response from CountTokens API", nil).
				WithCode("GENAI-011").
				WithSuggestion("This may be a temporary API issue, retry later")
		} else if errors.Is(err, context.DeadlineExceeded) {
			lastError = customerrors.WrapAPIError(err, "token counting timed out").
				WithCode("GENAI-012").
				WithSuggestion("Consider increasing the timeout value")
		} else {
			lastError = customerrors.WrapAPIError(err, "failed to count tokens").
				WithCode("GENAI-013")
		}

		// Simple backoff before retry
		if attempt < c.options.MaxRetries {
			backoffMs := 100 * attempt * attempt // Exponential backoff
			time.Sleep(time.Duration(backoffMs) * time.Millisecond)
		}
	}

	// If we get here, all retries failed
	return 0, customerrors.WrapAPIError(lastError, fmt.Sprintf("failed to count tokens after %d attempts", c.options.MaxRetries)).
		WithCode("GENAI-014").
		WithSuggestion("Check internet connectivity and API key validity")
}

// GenerateStream implements the Client interface for GeminiClient.
// It sends the prompt to the Gemini API and processes the streaming response.
// This method returns a channel that will receive text chunks as they are generated.
func (c *GeminiClient) GenerateStream(ctx context.Context, prompt string) (<-chan StreamChunk, error) {
	if c.client == nil || c.model == "" {
		return nil, customerrors.NewValidationError("client is not properly initialized", nil).
			WithCode("GENAI-015").
			WithSuggestion("Ensure the client was created with a valid API key and model name")
	}

	// Create a context with timeout if specified
	var genCtx context.Context
	var cancel context.CancelFunc
	if c.options.Timeout > 0 {
		genCtx, cancel = context.WithTimeout(ctx, time.Duration(c.options.Timeout)*time.Second)
	} else {
		genCtx = ctx
		cancel = func() {} // No-op cancel function
	}

	// Create a channel for streaming content back to the caller
	chunkChan := make(chan StreamChunk)

	// Prepare the content for the request
	contents := []*genai.Content{
		genai.NewContentFromText(prompt, "user"),
	}

	// Create generation config with our options
	genConfig := &genai.GenerateContentConfig{}

	// Apply generation parameters if they have non-zero values
	if c.options.Temperature > 0 {
		genConfig.Temperature = &c.options.Temperature
	}

	if c.options.TopP > 0 {
		genConfig.TopP = &c.options.TopP
	}

	if c.options.TopK > 0 {
		genConfig.TopK = &c.options.TopK
	}

	if c.options.MaxOutputTokens > 0 {
		genConfig.MaxOutputTokens = c.options.MaxOutputTokens
	}

	// Candidate count doesn't make sense for streaming, so we omit it

	if len(c.options.StopSequences) > 0 {
		genConfig.StopSequences = c.options.StopSequences
	}

	// Apply safety settings if any are defined
	if len(c.options.SafetySettings) > 0 {
		genConfig.SafetySettings = make([]*genai.SafetySetting, 0, len(c.options.SafetySettings))
		for _, ss := range c.options.SafetySettings {
			// Convert string category and threshold to genai enum types
			category := genai.HarmCategory(ss.Category)
			threshold := genai.HarmBlockThreshold(ss.Threshold)

			genConfig.SafetySettings = append(genConfig.SafetySettings, &genai.SafetySetting{
				Category:  category,
				Threshold: threshold,
			})
		}
	}

	// Prepare contents with system instructions if provided
	if c.options.SystemInstructions != "" {
		// Add system instructions as the first content item
		systemContent := genai.NewContentFromText(c.options.SystemInstructions, "system")
		contents = append([]*genai.Content{systemContent}, contents...)
	}

	// Start a goroutine to handle the streaming response
	go func() {
		defer close(chunkChan)
		defer cancel()

		var lastError error
		success := false

		// Retry logic
		for attempt := 1; attempt <= c.options.MaxRetries; attempt++ {
			if attempt > 1 {
				logrus.WithFields(logrus.Fields{
					"attempt":     attempt,
					"max_retries": c.options.MaxRetries,
				}).Debug("Retry attempt for streaming content")
			}

			// Create a stream for the response using our configuration
			streamChan := c.client.Models.GenerateContentStream(genCtx, c.model, contents, genConfig)

			// Process the streaming response
			chunkReceived := false
			responseFinished := false

			for resp := range streamChan {
				// Check for context canceled or deadline exceeded
				if errors.Is(genCtx.Err(), context.Canceled) {
					lastError = customerrors.WrapAPIError(genCtx.Err(), "context was canceled during stream").
						WithCode("GENAI-016")
					responseFinished = true
					break
				}

				if errors.Is(genCtx.Err(), context.DeadlineExceeded) {
					lastError = customerrors.WrapAPIError(genCtx.Err(), "streaming content generation timed out").
						WithCode("GENAI-017").
						WithSuggestion("Consider increasing the timeout value")
					responseFinished = true
					break
				}

				// If the response has an error, break and retry
				if resp == nil {
					lastError = customerrors.NewAPIError("received nil response", nil).
						WithCode("GENAI-018").
						WithSuggestion("This may be a temporary API issue, retry later")
					break
				}

				// Extract text from the response
				if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
					candidate := resp.Candidates[0]

					// Check for finish reason issues
					if candidate.FinishReason != "" && candidate.FinishReason != "FINISHED" && candidate.FinishReason != "STOP" {
						reason := candidate.FinishReason
						if reason == "SAFETY" {
							lastError = customerrors.NewAPIError("content blocked by safety settings", nil).
								WithCode("GENAI-019").
								WithSuggestion("Modify the prompt to avoid potentially harmful content")
						} else {
							lastError = customerrors.NewAPIError(fmt.Sprintf("generation incomplete: %s", reason), nil).
								WithCode("GENAI-020")
						}
						responseFinished = true
						break
					}

					// Send each part to the channel
					for _, part := range candidate.Content.Parts {
						if part.Text != "" {
							chunkChan <- StreamChunk{
								Text: part.Text,
							}
							chunkReceived = true
						}
					}
				}
			}

			// Check if we got any chunks back
			if chunkReceived && !responseFinished {
				// Success - we got content and it finished normally
				success = true
				break
			}

			// Simple backoff before retry
			if attempt < c.options.MaxRetries {
				backoffMs := 100 * attempt * attempt // Exponential backoff
				time.Sleep(time.Duration(backoffMs) * time.Millisecond)
			}
		}

		// Send final chunk with error if we failed
		if !success {
			var streamError error
			if lastError != nil {
				streamError = customerrors.WrapAPIError(lastError, fmt.Sprintf(
					"failed to generate streaming content after %d attempts",
					c.options.MaxRetries)).
					WithCode("GENAI-021").
					WithSuggestion("Check internet connectivity, API key validity, and prompt content")
			} else {
				streamError = customerrors.NewAPIError(
					fmt.Sprintf("failed to generate streaming content after %d attempts", c.options.MaxRetries),
					nil).
					WithCode("GENAI-022").
					WithSuggestion("Check your prompt content and try again")
			}

			chunkChan <- StreamChunk{
				Error: streamError,
				Done:  true,
			}
		} else {
			// Send a final chunk to signal completion
			chunkChan <- StreamChunk{
				Done: true,
			}
		}
	}()

	return chunkChan, nil
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
