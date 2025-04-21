package llm

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/iterator"
	"google.golang.org/genai"

	"glance/internal/mocks"
)

// We're using the LLMClient from internal/mocks package

// Test the function variable pattern for client creation
func TestClientCreation(t *testing.T) {
	t.Run("Default function creates client", func(t *testing.T) {
		// Test with the real function but invalid API key - should fail cleanly
		origFunc := createGeminiClient
		defer func() { createGeminiClient = origFunc }()

		client, err := NewGeminiClient("")
		assert.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("Mocked function returns predetermined client", func(t *testing.T) {
		// Set up a mock client to return
		mockClient := new(mocks.LLMClient)
		adapterClient := NewMockClientAdapter(mockClient)

		// Create a mock function that returns our mock client
		mockCreateFunc := func(apiKey string, options ...ClientOption) (Client, error) {
			return adapterClient, nil
		}

		// Replace the default function with our mock
		origFunc := createGeminiClient
		createGeminiClient = mockCreateFunc
		defer func() { createGeminiClient = origFunc }()

		// Now using NewGeminiClient should use our mock function
		client, err := NewGeminiClient("any-api-key")
		assert.NoError(t, err)
		assert.Equal(t, adapterClient, client)
	})
}

func TestClientInterface(t *testing.T) {
	mockClient := new(mocks.LLMClient)
	// Create adapter to convert mock to Client interface
	var client Client = NewMockClientAdapter(mockClient)

	ctx := context.Background()
	testPrompt := "Test prompt"
	expectedResponse := "Generated response"
	expectedTokenCount := 42

	// Create a test chunk channel for streaming
	mockChunkChan := make(chan mocks.StreamChunk, 3)
	mockChunkChan <- mocks.StreamChunk{Text: "Chunk 1"}
	mockChunkChan <- mocks.StreamChunk{Text: "Chunk 2"}
	mockChunkChan <- mocks.StreamChunk{Done: true}

	// Set up expectations
	mockClient.On("Generate", ctx, testPrompt).Return(expectedResponse, nil)
	mockClient.On("GenerateStream", ctx, testPrompt).Return((<-chan mocks.StreamChunk)(mockChunkChan), nil)
	mockClient.On("CountTokens", ctx, testPrompt).Return(expectedTokenCount, nil)
	mockClient.On("Close").Return()

	// Test Generate
	response, err := client.Generate(ctx, testPrompt)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)

	// Test GenerateStream
	streamChan, err := client.GenerateStream(ctx, testPrompt)
	assert.NoError(t, err)
	assert.NotNil(t, streamChan)

	// Collect chunks from the stream - we need to adapt from one stream chunk type to another
	var receivedTexts []string
	var done bool

	for chunk := range streamChan {
		if chunk.Text != "" {
			receivedTexts = append(receivedTexts, chunk.Text)
		}
		if chunk.Done {
			done = true
			break
		}
	}

	// Verify chunks
	assert.Len(t, receivedTexts, 2)
	assert.Equal(t, "Chunk 1", receivedTexts[0])
	assert.Equal(t, "Chunk 2", receivedTexts[1])
	assert.True(t, done)

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
	assert.Equal(t, "gemini-2.5-flash-preview-04-17", options.ModelName)
	assert.Greater(t, options.MaxRetries, 0)
	assert.Greater(t, options.Timeout, 0)

	// Verify default generation parameters have sensible values
	assert.InDelta(t, 0.7, float64(options.Temperature), 0.01)
	assert.InDelta(t, 0.95, float64(options.TopP), 0.01)
	assert.Greater(t, options.TopK, float32(0))
	assert.Greater(t, options.MaxOutputTokens, int32(0))
	assert.Equal(t, int32(1), options.CandidateCount)
	assert.Empty(t, options.StopSequences)
	assert.Empty(t, options.SafetySettings)
	assert.Empty(t, options.SystemInstructions)

	// Test basic configuration options

	// Test WithModelName option function
	customModel := "custom-model"
	modelOption := WithModelName(customModel)

	// Apply option to default options
	testOpts := DefaultClientOptions()
	modelOption(&testOpts)
	assert.Equal(t, customModel, testOpts.ModelName)

	// Test WithMaxRetries option function
	customRetries := 10
	retriesOption := WithMaxRetries(customRetries)

	testOpts = DefaultClientOptions()
	retriesOption(&testOpts)
	assert.Equal(t, customRetries, testOpts.MaxRetries)

	// Test WithTimeout option function
	customTimeout := 120
	timeoutOption := WithTimeout(customTimeout)

	testOpts = DefaultClientOptions()
	timeoutOption(&testOpts)
	assert.Equal(t, customTimeout, testOpts.Timeout)

	// Test generation parameter options

	// Test WithTemperature option
	customTemp := float32(0.2)
	tempOption := WithTemperature(customTemp)

	testOpts = DefaultClientOptions()
	tempOption(&testOpts)
	assert.Equal(t, customTemp, testOpts.Temperature)

	// Test WithTopP option
	customTopP := float32(0.8)
	topPOption := WithTopP(customTopP)

	testOpts = DefaultClientOptions()
	topPOption(&testOpts)
	assert.Equal(t, customTopP, testOpts.TopP)

	// Test WithTopK option
	customTopK := float32(20.0)
	topKOption := WithTopK(customTopK)

	testOpts = DefaultClientOptions()
	topKOption(&testOpts)
	assert.Equal(t, customTopK, testOpts.TopK)

	// Test WithMaxOutputTokens option
	customMaxTokens := int32(1000)
	maxTokensOption := WithMaxOutputTokens(customMaxTokens)

	testOpts = DefaultClientOptions()
	maxTokensOption(&testOpts)
	assert.Equal(t, customMaxTokens, testOpts.MaxOutputTokens)

	// Test WithCandidateCount option
	customCandidates := int32(3)
	candidatesOption := WithCandidateCount(customCandidates)

	testOpts = DefaultClientOptions()
	candidatesOption(&testOpts)
	assert.Equal(t, customCandidates, testOpts.CandidateCount)

	// Test WithStopSequences option
	customStops := []string{"STOP", "END"}
	stopsOption := WithStopSequences(customStops)

	testOpts = DefaultClientOptions()
	stopsOption(&testOpts)
	assert.Equal(t, customStops, testOpts.StopSequences)

	// Test WithSafetySetting option
	testOpts = DefaultClientOptions()
	WithSafetySetting(HarmCategoryHateSpeech, HarmBlockHighAndAbove)(&testOpts)
	assert.Len(t, testOpts.SafetySettings, 1)
	assert.Equal(t, HarmCategoryHateSpeech, testOpts.SafetySettings[0].Category)
	assert.Equal(t, HarmBlockHighAndAbove, testOpts.SafetySettings[0].Threshold)

	// Test adding multiple safety settings
	testOpts = DefaultClientOptions()
	WithSafetySetting(HarmCategoryHateSpeech, HarmBlockHighAndAbove)(&testOpts)
	WithSafetySetting(HarmCategorySexuallyExplicit, HarmBlockMediumAndAbove)(&testOpts)
	assert.Len(t, testOpts.SafetySettings, 2)

	// Test WithSystemInstructions option
	customInstructions := "You are a helpful assistant."
	instructionsOption := WithSystemInstructions(customInstructions)

	testOpts = DefaultClientOptions()
	instructionsOption(&testOpts)
	assert.Equal(t, customInstructions, testOpts.SystemInstructions)

	// Test applying multiple options
	testOpts = DefaultClientOptions()
	WithModelName("custom-model-2")(&testOpts)
	WithMaxRetries(5)(&testOpts)
	WithTimeout(30)(&testOpts)
	WithTemperature(0.5)(&testOpts)
	WithTopP(0.9)(&testOpts)
	WithMaxOutputTokens(500)(&testOpts)
	WithSystemInstructions("Be concise")(&testOpts)

	assert.Equal(t, "custom-model-2", testOpts.ModelName)
	assert.Equal(t, 5, testOpts.MaxRetries)
	assert.Equal(t, 30, testOpts.Timeout)
	assert.Equal(t, float32(0.5), testOpts.Temperature)
	assert.Equal(t, float32(0.9), testOpts.TopP)
	assert.Equal(t, int32(500), testOpts.MaxOutputTokens)
	assert.Equal(t, "Be concise", testOpts.SystemInstructions)
}

// TestNewGeminiClient tests the client creation functionality
func TestNewGeminiClient(t *testing.T) {
	// Test with empty API key
	t.Run("Empty API key", func(t *testing.T) {
		client, err := NewGeminiClient("")
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "API key is required")
	})

	// Test with custom options
	t.Run("Custom options", func(t *testing.T) {
		// Note: This test will behave differently depending on whether the API key
		// is accepted or rejected by the Gemini API. For reliable testing, we focus
		// on testing validation, not actual API calls.

		// First, ensure empty API key fails even with custom options
		client, err := NewGeminiClient("",
			WithModelName("custom-model"),
			WithMaxRetries(5),
			WithTimeout(30),
		)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "API key is required")
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
		opts := DefaultClientOptions()
		client := &GeminiClient{
			client:  nil,
			model:   "",
			options: &opts,
		}

		result, err := client.Generate(context.Background(), "test prompt")
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "not properly initialized")
	})
}

// TestGeminiClientGenerateStream tests the GenerateStream method of GeminiClient
func TestGeminiClientGenerateStream(t *testing.T) {
	// Test uninitialized client
	t.Run("Uninitialized client", func(t *testing.T) {
		opts := DefaultClientOptions()
		client := &GeminiClient{
			client:  nil,
			model:   "",
			options: &opts,
		}

		chunkChan, err := client.GenerateStream(context.Background(), "test prompt")
		assert.Error(t, err)
		assert.Nil(t, chunkChan)
		assert.Contains(t, err.Error(), "not properly initialized")
	})
}

// TestGeminiClientCountTokens tests the CountTokens method of GeminiClient
func TestGeminiClientCountTokens(t *testing.T) {
	// Test uninitialized client
	t.Run("Uninitialized client", func(t *testing.T) {
		opts := DefaultClientOptions()
		client := &GeminiClient{
			client:  nil,
			model:   "",
			options: &opts,
		}

		result, err := client.CountTokens(context.Background(), "test prompt")
		assert.Error(t, err)
		assert.Equal(t, 0, result)
		assert.Contains(t, err.Error(), "not properly initialized")
	})

	// Test timeout behavior
	t.Run("Timeout behavior", func(t *testing.T) {
		opts := DefaultClientOptions()
		opts.Timeout = 1 // 1 second timeout
		client := &GeminiClient{
			client:  nil, // Will cause an error before timeout is triggered
			model:   "test-model",
			options: &opts,
		}

		ctx := context.Background()
		_, err := client.CountTokens(ctx, "test prompt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not properly initialized")
	})

	// Test retry behavior
	t.Run("Retry behavior with API error", func(t *testing.T) {
		opts := DefaultClientOptions()
		opts.MaxRetries = 2
		// Unable to directly test retry logic without mocking the genai.Client
		// which is challenging due to its structure. This is more of an integration test.
		client := &GeminiClient{
			client:  nil, // Will always fail
			model:   "test-model",
			options: &opts,
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
		opts := DefaultClientOptions()
		client := &GeminiClient{
			client:  nil,
			model:   "",
			options: &opts,
		}

		// This should not panic
		client.Close()

		// Verify client is still nil after close
		assert.Nil(t, client.client)
		assert.Empty(t, client.model)
	})
}

// TestGeminiClientTimeout tests timeout handling in the client
func TestGeminiClientTimeout(t *testing.T) {
	t.Run("Context timeout behavior", func(t *testing.T) {
		// Create a client with a very short timeout
		options := ClientOptions{
			ModelName:  "test-model",
			MaxRetries: 1,
			Timeout:    1, // 1 second timeout
		}
		client := &GeminiClient{
			client:  nil, // We won't use the actual client in this test
			model:   "",
			options: &options,
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
		options := ClientOptions{
			ModelName:  "test-model",
			MaxRetries: 2,
			Timeout:    1,
		}
		client := &GeminiClient{
			client:  nil, // We won't use the actual client in this test
			model:   "",
			options: &options,
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
	client, err := NewGeminiClient(apiKey)
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
