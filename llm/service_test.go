package llm

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// We're using the MockClient defined in client_test.go

func TestNewService(t *testing.T) {
	// Test with nil client
	t.Run("Nil client", func(t *testing.T) {
		service, err := NewService(nil)
		assert.Error(t, err)
		assert.Nil(t, service)
	})

	// Test with valid client and default options
	t.Run("Valid client with default options", func(t *testing.T) {
		mockClient := new(MockClient)
		service, err := NewService(mockClient)
		
		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, DefaultServiceOptions().MaxRetries, service.options.MaxRetries)
	})
	
	// Test with valid client and custom options
	t.Run("Valid client with custom options", func(t *testing.T) {
		mockClient := new(MockClient)
		customRetries := 10
		service, err := NewService(mockClient, WithMaxRetries(customRetries))
		
		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, customRetries, service.options.MaxRetries)
	})
}

func TestGenerateGlanceMarkdown(t *testing.T) {
	mockClient := new(MockClient)
	ctx := context.Background()
	
	// Test data
	dir := "/test/dir"
	fileMap := map[string]string{
		"file1.txt": "Content 1",
		"file2.go":  "Content 2",
	}
	subGlances := "Sub glances content"
	expectedResponse := "Generated markdown content"
	
	// Test successful generation on first attempt
	t.Run("Successful generation", func(t *testing.T) {
		// Setup service with mock client
		service, err := NewService(mockClient)
		assert.NoError(t, err)
		
		// Setup expectations for the mock
		mockClient.On("Generate", ctx, mock.AnythingOfType("string")).Return(expectedResponse, nil).Once()
		mockClient.On("CountTokens", ctx, mock.AnythingOfType("string")).Return(100, nil).Maybe()
		
		// Call the method
		result, err := service.GenerateGlanceMarkdown(ctx, dir, fileMap, subGlances)
		
		// Verify results
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
		mockClient.AssertExpectations(t)
	})
	
	// Test with retries
	t.Run("Generation succeeds after retries", func(t *testing.T) {
		// Reset mock
		mockClient = new(MockClient)
		
		// Setup service with mock client and 3 retries
		service, err := NewService(mockClient, WithMaxRetries(3))
		assert.NoError(t, err)
		
		// Setup expectations - first 2 attempts fail, 3rd succeeds
		expectedError := errors.New("API error")
		mockClient.On("Generate", ctx, mock.AnythingOfType("string")).Return("", expectedError).Times(2)
		mockClient.On("Generate", ctx, mock.AnythingOfType("string")).Return(expectedResponse, nil).Once()
		mockClient.On("CountTokens", ctx, mock.AnythingOfType("string")).Return(100, nil).Maybe()
		
		// Call the method
		result, err := service.GenerateGlanceMarkdown(ctx, dir, fileMap, subGlances)
		
		// Verify results
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
		mockClient.AssertExpectations(t)
	})
	
	// Test with all retries failing
	t.Run("All retries fail", func(t *testing.T) {
		// Reset mock
		mockClient = new(MockClient)
		
		// Setup service with mock client and 2 retries
		service, err := NewService(mockClient, WithMaxRetries(2))
		assert.NoError(t, err)
		
		// Setup expectations - all attempts fail
		expectedError := errors.New("persistent API error")
		mockClient.On("Generate", ctx, mock.AnythingOfType("string")).Return("", expectedError).Times(3)
		mockClient.On("CountTokens", ctx, mock.AnythingOfType("string")).Return(100, nil).Maybe()
		
		// Call the method
		result, err := service.GenerateGlanceMarkdown(ctx, dir, fileMap, subGlances)
		
		// Verify results
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "persistent API error")
		assert.Empty(t, result)
		mockClient.AssertExpectations(t)
	})
}

func TestServiceOptions(t *testing.T) {
	// Test default options
	defaults := DefaultServiceOptions()
	assert.Greater(t, defaults.MaxRetries, 0)
	assert.NotEmpty(t, defaults.ModelName)
	
	// Test WithMaxRetries
	retries := 5
	options := DefaultServiceOptions().WithMaxRetries(retries)
	assert.Equal(t, retries, options.MaxRetries)
	
	// Test WithModelName
	modelName := "custom-model"
	options = DefaultServiceOptions().WithModelName(modelName)
	assert.Equal(t, modelName, options.ModelName)
	
	// Test WithVerbose
	options = DefaultServiceOptions().WithVerbose(true)
	assert.True(t, options.Verbose)
}