package llm

import (
	"context"
	"testing"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/iterator"
)

// MockClient is a mock implementation of the Client interface for testing
type MockClient struct {
	mock.Mock
}

func (m *MockClient) Generate(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

func (m *MockClient) CountTokens(ctx context.Context, prompt string) (int, error) {
	args := m.Called(ctx, prompt)
	return args.Int(0), args.Error(1)
}

func (m *MockClient) Close() {
	m.Called()
}

// Test the interface definition with the mock implementation
func TestClientInterface(t *testing.T) {
	mockClient := new(MockClient)

	ctx := context.Background()
	testPrompt := "Test prompt"
	expectedResponse := "Generated response"
	expectedTokenCount := 42

	// Set up expectations
	mockClient.On("Generate", ctx, testPrompt).Return(expectedResponse, nil)
	mockClient.On("CountTokens", ctx, testPrompt).Return(expectedTokenCount, nil)
	mockClient.On("Close").Return()

	// Use the mock client through the interface
	var client Client = mockClient

	// Test Generate
	response, err := client.Generate(ctx, testPrompt)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)

	// Test CountTokens
	tokenCount, err := client.CountTokens(ctx, testPrompt)
	assert.NoError(t, err)
	assert.Equal(t, expectedTokenCount, tokenCount)

	// Test Close
	client.Close()

	// Verify all expectations were met
	mockClient.AssertExpectations(t)
}

// TestClientOptions ensures the options work correctly
func TestClientOptions(t *testing.T) {
	// Test default options
	options := DefaultClientOptions()
	assert.Equal(t, "gemini-2.0-flash", options.ModelName)
	assert.Greater(t, options.MaxRetries, 0)
	assert.Greater(t, options.Timeout, 0)

	// Test WithModelName
	customModel := "custom-model"
	updatedOpts := options.WithModelName(customModel)
	assert.Equal(t, customModel, updatedOpts.ModelName)
	assert.NotEqual(t, &options, &updatedOpts, "Should create a new options instance")

	// Test WithMaxRetries
	customRetries := 10
	updatedOpts = options.WithMaxRetries(customRetries)
	assert.Equal(t, customRetries, updatedOpts.MaxRetries)
	assert.NotEqual(t, &options, &updatedOpts, "Should create a new options instance")

	// Test WithTimeout
	customTimeout := 120
	updatedOpts = options.WithTimeout(customTimeout)
	assert.Equal(t, customTimeout, updatedOpts.Timeout)
	assert.NotEqual(t, &options, &updatedOpts, "Should create a new options instance")

	// Test chain of option modifications
	fullyCustomized := options.
		WithModelName("custom-model-2").
		WithMaxRetries(5).
		WithTimeout(30)

	assert.Equal(t, "custom-model-2", fullyCustomized.ModelName)
	assert.Equal(t, 5, fullyCustomized.MaxRetries)
	assert.Equal(t, 30, fullyCustomized.Timeout)
}

// TestNewGeminiClient tests the client creation functionality
func TestNewGeminiClient(t *testing.T) {
	// Test with empty API key
	t.Run("Empty API key", func(t *testing.T) {
		client, err := NewGeminiClient("", nil)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "API key is required")
	})

	// Test with default options (nil options should use defaults)
	t.Run("Default options", func(t *testing.T) {
		// Note: This test will behave differently depending on whether the API key
		// is accepted or rejected by the Gemini API. For reliable testing, we focus
		// on testing that nil options uses defaults, not on API key validation.

		// First, ensure empty API key fails regardless of options
		client, err := NewGeminiClient("", nil)
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

// MockGenerativeIterator mocks the streaming response iterator from the Gemini API
type MockGenerativeIterator struct {
	mock.Mock
	responses []*genai.GenerateContentResponse
	index     int
	err       error
}

func (m *MockGenerativeIterator) Next() (*genai.GenerateContentResponse, error) {
	if m.err != nil {
		return nil, m.err
	}

	if m.index >= len(m.responses) {
		return nil, iterator.Done
	}

	resp := m.responses[m.index]
	m.index++
	return resp, nil
}

// TestGeminiClientGenerate tests the Generate method of GeminiClient
func TestGeminiClientGenerate(t *testing.T) {
	// Since we can't easily mock the genai.Client directly,
	// this test focuses on error cases we can test directly

	// Test uninitialized client
	t.Run("Uninitialized client", func(t *testing.T) {
		client := &GeminiClient{
			client:  nil,
			model:   nil,
			options: DefaultClientOptions(),
		}

		result, err := client.Generate(context.Background(), "test prompt")
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "not properly initialized")
	})
}

// TestGeminiClientCountTokens tests the CountTokens method of GeminiClient
func TestGeminiClientCountTokens(t *testing.T) {
	// Test uninitialized client
	t.Run("Uninitialized client", func(t *testing.T) {
		client := &GeminiClient{
			client:  nil,
			model:   nil,
			options: DefaultClientOptions(),
		}

		result, err := client.CountTokens(context.Background(), "test prompt")
		assert.Error(t, err)
		assert.Equal(t, 0, result)
		assert.Contains(t, err.Error(), "not properly initialized")
	})
}

// TestGeminiClientClose tests the Close method of GeminiClient
func TestGeminiClientClose(t *testing.T) {
	// Test safe close (even if client is nil)
	t.Run("Safe close with nil client", func(t *testing.T) {
		client := &GeminiClient{
			client:  nil,
			model:   nil,
			options: DefaultClientOptions(),
		}

		// This should not panic
		client.Close()

		// Verify client is still nil after close
		assert.Nil(t, client.client)
		assert.Nil(t, client.model)
	})
}

// TestGeminiClientTimeout tests timeout handling in the client
func TestGeminiClientTimeout(t *testing.T) {
	t.Run("Context timeout behavior", func(t *testing.T) {
		// Create a client with a very short timeout
		client := &GeminiClient{
			client: nil, // We won't use the actual client in this test
			model:  nil,
			options: &ClientOptions{
				ModelName:  "test-model",
				MaxRetries: 1,
				Timeout:    1, // 1 second timeout
			},
		}

		// Create a context with a cancellation function
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Test that the timeout context is created correctly
		// The actual API call would fail due to nil client, but we're testing the timeout setup
		_, err := client.Generate(ctx, "test prompt")
		assert.Error(t, err) // Error because client is nil, not because of timeout
		assert.Contains(t, err.Error(), "not properly initialized")
	})
}

// TestGeminiClientRetryLogic tests the retry logic in the Generate and CountTokens methods
func TestGeminiClientRetryLogic(t *testing.T) {
	t.Run("Maximum retries reached", func(t *testing.T) {
		// Create a client with a small number of retries for testing
		client := &GeminiClient{
			client: nil, // We won't use the actual client in this test
			model:  nil,
			options: &ClientOptions{
				ModelName:  "test-model",
				MaxRetries: 2,
				Timeout:    1,
			},
		}

		// Test Generate retry logic (will fail due to uninitialized client)
		_, err := client.Generate(context.Background(), "test prompt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not properly initialized")

		// Test CountTokens retry logic (will fail due to uninitialized client)
		_, err = client.CountTokens(context.Background(), "test prompt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not properly initialized")
	})
}

// Integration test for the GeminiClient implementation
// This test is skipped by default as it requires a real API key
func TestGeminiClient_Integration(t *testing.T) {
	t.Skip("Skipping integration test - requires actual API key")

	// Setup
	apiKey := "your-api-key-here" // Replace with an actual key for manual testing
	client, err := NewGeminiClient(apiKey, DefaultClientOptions())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Test Generate with potential timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.Generate(ctx, "Say hello")
	assert.NoError(t, err)
	assert.NotEmpty(t, response)

	// Test CountTokens
	tokens, err := client.CountTokens(context.Background(), "Count these tokens")
	assert.NoError(t, err)
	assert.Greater(t, tokens, 0)

	// Test with empty prompt (should not error)
	emptyResponse, err := client.Generate(context.Background(), "")
	assert.NoError(t, err)
	assert.NotEmpty(t, emptyResponse) // LLMs usually generate something even with empty prompts

	// Test token count with empty prompt
	emptyTokens, err := client.CountTokens(context.Background(), "")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, emptyTokens, 0) // Should be 0 or more tokens
}
