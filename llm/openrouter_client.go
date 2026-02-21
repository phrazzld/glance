// Package llm provides abstractions and implementations for interacting with
// Large Language Model APIs in the glance application.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	customerrors "glance/errors"
)

const (
	openRouterBaseURL      = "https://openrouter.ai/api/v1"
	openRouterBodyLimit    = 8 * 1024 * 1024 // 8MB
	openRouterCodeBase     = "OPENROUTER"
	openRouterDefaultTitle = "failed to generate content"
)

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterChatRequest struct {
	Model       string              `json:"model"`
	Messages    []openRouterMessage `json:"messages"`
	MaxTokens   int32               `json:"max_tokens,omitempty"`
	Temperature *float32            `json:"temperature,omitempty"`
	TopP        *float32            `json:"top_p,omitempty"`
	TopK        *int32              `json:"top_k,omitempty"`
	Stop        []string            `json:"stop,omitempty"`
}

type openRouterError struct {
	Message string `json:"message"`
}

type openRouterChoiceMessage struct {
	Content json.RawMessage `json:"content"`
}

type openRouterChoice struct {
	Message openRouterChoiceMessage `json:"message"`
}

type openRouterChatResponse struct {
	Choices []openRouterChoice `json:"choices"`
	Error   *openRouterError   `json:"error"`
}

// OpenRouterClient is a Client implementation that uses OpenRouter's chat API.
type OpenRouterClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	model      string
	options    *ClientOptions
}

// NewOpenRouterClientFunc is a function type for creating OpenRouter clients.
// This enables mocking in tests.
type NewOpenRouterClientFunc func(apiKey string, options ...ClientOption) (Client, error)

// The actual implementation function - can be swapped in tests
var createOpenRouterClient NewOpenRouterClientFunc = func(apiKey string, options ...ClientOption) (Client, error) {
	return newOpenRouterClient(apiKey, options...)
}

// NewOpenRouterClient creates a new client for the OpenRouter API.
func NewOpenRouterClient(apiKey string, options ...ClientOption) (Client, error) {
	return createOpenRouterClient(apiKey, options...)
}

// newOpenRouterClient is the actual implementation for creating an OpenRouter client.
func newOpenRouterClient(apiKey string, options ...ClientOption) (*OpenRouterClient, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, customerrors.NewValidationError("OpenRouter API key is required", nil).
			WithCode(openRouterCodeBase + "-001").
			WithSuggestion("Set OPENROUTER_API_KEY in your environment")
	}

	opts := DefaultClientOptions()
	for _, option := range options {
		option(&opts)
	}

	if strings.TrimSpace(opts.ModelName) == "" {
		return nil, customerrors.NewValidationError("OpenRouter model name is required", nil).
			WithCode(openRouterCodeBase + "-002")
	}

	timeout := time.Duration(opts.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	return &OpenRouterClient{
		httpClient: &http.Client{Timeout: timeout},
		apiKey:     apiKey, // pragma: allowlist secret
		baseURL:    openRouterBaseURL,
		model:      opts.ModelName,
		options:    &opts,
	}, nil
}

// Generate sends the prompt to OpenRouter and returns the generated text.
func (c *OpenRouterClient) Generate(ctx context.Context, prompt string) (string, error) {
	if c.httpClient == nil || c.model == "" {
		return "", customerrors.NewValidationError("OpenRouter client is not properly initialized", nil).
			WithCode(openRouterCodeBase + "-003")
	}

	maxAttempts := c.options.MaxRetries + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		content, err := c.generateOnce(ctx, prompt)
		if err == nil {
			return content, nil
		}
		lastErr = err

		if attempt < maxAttempts {
			backoff := time.Duration(100*attempt*attempt) * time.Millisecond
			if sleepErr := sleepWithContext(ctx, backoff); sleepErr != nil {
				return "", sleepErr
			}
		}
	}

	return "", customerrors.WrapAPIError(lastErr, fmt.Sprintf("%s after %d attempts", openRouterDefaultTitle, maxAttempts)).
		WithCode(openRouterCodeBase + "-004")
}

func (c *OpenRouterClient) generateOnce(ctx context.Context, prompt string) (string, error) {
	reqBody := openRouterChatRequest{
		Model:    c.model,
		Messages: c.buildMessages(prompt),
	}

	if c.options.MaxOutputTokens > 0 {
		reqBody.MaxTokens = c.options.MaxOutputTokens
	}
	if c.options.Temperature > 0 {
		temp := c.options.Temperature
		reqBody.Temperature = &temp
	}
	if c.options.TopP > 0 {
		topP := c.options.TopP
		reqBody.TopP = &topP
	}
	if c.options.TopK > 0 {
		topK := int32(c.options.TopK)
		reqBody.TopK = &topK
	}
	if len(c.options.StopSequences) > 0 {
		reqBody.Stop = c.options.StopSequences
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", customerrors.WrapAPIError(err, "failed to encode OpenRouter request").
			WithCode(openRouterCodeBase + "-005")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/chat/completions",
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", customerrors.WrapAPIError(err, "failed to build OpenRouter request").
			WithCode(openRouterCodeBase + "-006")
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey) // pragma: allowlist secret
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", customerrors.WrapAPIError(err, "OpenRouter request failed").
			WithCode(openRouterCodeBase + "-007")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, openRouterBodyLimit))
	if err != nil {
		return "", customerrors.WrapAPIError(err, "failed reading OpenRouter response").
			WithCode(openRouterCodeBase + "-008")
	}

	var parsed openRouterChatResponse
	if len(bodyBytes) > 0 {
		_ = json.Unmarshal(bodyBytes, &parsed)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(parsed.errorMessage())
		if msg == "" {
			msg = strings.TrimSpace(string(bodyBytes))
		}
		if msg == "" {
			msg = "request failed with non-success status"
		}

		apiErr := customerrors.NewAPIError(
			fmt.Sprintf("OpenRouter returned status %d: %s", resp.StatusCode, msg),
			nil,
		).WithCode(openRouterCodeBase + "-009")

		if resp.StatusCode == http.StatusTooManyRequests {
			apiErr = apiErr.WithSuggestion("Rate limited by provider. Retry after backoff")
		}

		return "", apiErr
	}

	if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
		return "", customerrors.NewAPIError(parsed.Error.Message, nil).
			WithCode(openRouterCodeBase + "-010")
	}

	if len(parsed.Choices) == 0 {
		return "", customerrors.NewAPIError("OpenRouter response had no choices", nil).
			WithCode(openRouterCodeBase + "-011")
	}

	content := extractOpenRouterContent(parsed.Choices[0].Message.Content)
	if strings.TrimSpace(content) == "" {
		return "", customerrors.NewAPIError("OpenRouter response content was empty", nil).
			WithCode(openRouterCodeBase + "-012")
	}

	return content, nil
}

// CountTokens is not currently implemented for OpenRouter's generic client API.
func (c *OpenRouterClient) CountTokens(ctx context.Context, prompt string) (int, error) {
	_ = ctx
	_ = prompt
	return 0, customerrors.NewAPIError("token counting is not supported for OpenRouter client", nil).
		WithCode(openRouterCodeBase + "-013")
}

// GenerateStream uses non-streaming generation and returns one final chunk.
func (c *OpenRouterClient) GenerateStream(ctx context.Context, prompt string) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 2)
	go func() {
		defer close(ch)

		content, err := c.Generate(ctx, prompt)
		if err != nil {
			ch <- StreamChunk{Error: err, Done: true}
			return
		}

		ch <- StreamChunk{Text: content}
		ch <- StreamChunk{Done: true}
	}()

	return ch, nil
}

// Close is a no-op because OpenRouterClient currently has no persistent resources.
func (c *OpenRouterClient) Close() {}

func (c *OpenRouterClient) buildMessages(prompt string) []openRouterMessage {
	messages := make([]openRouterMessage, 0, 2)
	if strings.TrimSpace(c.options.SystemInstructions) != "" {
		messages = append(messages, openRouterMessage{
			Role:    "system",
			Content: c.options.SystemInstructions,
		})
	}
	messages = append(messages, openRouterMessage{
		Role:    "user",
		Content: prompt,
	})
	return messages
}

func (r *openRouterChatResponse) errorMessage() string {
	if r == nil || r.Error == nil {
		return ""
	}
	return strings.TrimSpace(r.Error.Message)
}

func extractOpenRouterContent(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	// Common case: simple string response.
	var contentString string
	if err := json.Unmarshal(raw, &contentString); err == nil {
		return contentString
	}

	// Alternate form: array of content parts.
	var contentParts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &contentParts); err == nil {
		var builder strings.Builder
		for _, part := range contentParts {
			if part.Text != "" {
				builder.WriteString(part.Text)
			}
		}
		return builder.String()
	}

	return ""
}
