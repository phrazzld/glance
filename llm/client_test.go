package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	
	// Test WithMaxRetries
	customRetries := 10
	updatedOpts = options.WithMaxRetries(customRetries)
	assert.Equal(t, customRetries, updatedOpts.MaxRetries)
	
	// Test WithTimeout
	customTimeout := 120
	updatedOpts = options.WithTimeout(customTimeout)
	assert.Equal(t, customTimeout, updatedOpts.Timeout)
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
	
	// Test Generate
	response, err := client.Generate(context.Background(), "Say hello")
	assert.NoError(t, err)
	assert.NotEmpty(t, response)
	
	// Test CountTokens
	tokens, err := client.CountTokens(context.Background(), "Count these tokens")
	assert.NoError(t, err)
	assert.Greater(t, tokens, 0)
}