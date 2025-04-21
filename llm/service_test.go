package llm

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"glance/internal/mocks"
)

// We're using the LLMClient defined in internal/mocks package

func TestNewService(t *testing.T) {
	// Test with nil client
	t.Run("Nil client", func(t *testing.T) {
		service, err := NewService(nil)
		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "client cannot be nil")
	})

	// Test with valid client and default options
	t.Run("Valid client with default options", func(t *testing.T) {
		mockClient := new(mocks.LLMClient)
		service, err := NewService(mockClient)

		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, mockClient, service.client)
		assert.Equal(t, DefaultServiceConfig().MaxRetries, service.maxRetries)
		assert.Equal(t, DefaultServiceConfig().ModelName, service.modelName)
		assert.Equal(t, DefaultServiceConfig().Verbose, service.verbose)
	})

	// Test with valid client and custom options
	t.Run("Valid client with custom options", func(t *testing.T) {
		mockClient := new(mocks.LLMClient)
		customRetries := 10
		service, err := NewService(mockClient, WithServiceMaxRetries(customRetries))

		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, customRetries, service.maxRetries)
	})

	// Test with multiple options
	t.Run("Multiple options", func(t *testing.T) {
		mockClient := new(mocks.LLMClient)
		service, err := NewService(mockClient,
			WithServiceMaxRetries(5),
			WithServiceModelName("custom-model"),
			WithVerbose(true))

		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, 5, service.maxRetries)
		assert.Equal(t, "custom-model", service.modelName)
		assert.True(t, service.verbose)
	})
}

func TestGenerateGlanceMarkdown(t *testing.T) {
	mockClient := new(mocks.LLMClient)
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
		// Setup service with mock client and custom template
		customTemplate := "Custom template for test {{.Directory}}"
		service, err := NewService(mockClient, WithPromptTemplate(customTemplate))
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
		mockClient = new(mocks.LLMClient)

		// Setup service with mock client and 3 retries
		service, err := NewService(mockClient, WithServiceMaxRetries(3), WithPromptTemplate("test template"))
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
		mockClient = new(mocks.LLMClient)

		// Setup service with mock client and 2 retries
		service, err := NewService(mockClient, WithServiceMaxRetries(2), WithPromptTemplate("test template"))
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

	// Test with template generation error
	t.Run("Template generation error", func(t *testing.T) {
		// Create a service with a mock client
		mockClient = new(mocks.LLMClient)

		// Create an invalid template to cause an error in prompt generation
		invalidTemplate := "Invalid template with {{.MissingVar}}"
		service, err := NewService(mockClient, WithPromptTemplate(invalidTemplate))
		assert.NoError(t, err)

		// This should fail due to invalid template with .MissingVar
		result, err := service.GenerateGlanceMarkdown(ctx, dir, fileMap, subGlances)

		// Now we expect an error from template generation
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template")
		assert.Empty(t, result) // No result since template failed
	})

	// Test with prompt template from options
	t.Run("Use prompt template from options", func(t *testing.T) {
		// Reset mock
		mockClient = new(mocks.LLMClient)

		// Create a custom template
		customTemplate := "Custom template from options with {{.Directory}}"

		// Setup service with mock client and custom template option
		service, err := NewService(mockClient, WithPromptTemplate(customTemplate))
		assert.NoError(t, err)

		// Setup expectations for the mock
		// The mock would receive a prompt generated from the custom template
		mockClient.On("Generate", ctx, mock.AnythingOfType("string")).Run(func(args mock.Arguments) {
			// Verify that the prompt contains the custom template
			prompt := args.String(1)
			assert.Contains(t, prompt, "Custom template from options with")
			assert.Contains(t, prompt, dir) // Should contain the directory name
		}).Return(expectedResponse, nil).Once()

		mockClient.On("CountTokens", ctx, mock.AnythingOfType("string")).Return(100, nil).Maybe()

		// Call the method
		result, err := service.GenerateGlanceMarkdown(ctx, dir, fileMap, subGlances)

		// Verify results
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
		mockClient.AssertExpectations(t)
	})
}

func TestServiceConfig(t *testing.T) {
	// Test default config
	defaults := DefaultServiceConfig()
	assert.Greater(t, defaults.MaxRetries, 0)
	assert.NotEmpty(t, defaults.ModelName)
	assert.False(t, defaults.Verbose)

	// Test config option functions
	// Create test config instance
	testConfig := DefaultServiceConfig()

	// Test WithServiceMaxRetries
	retries := 5
	retriesOption := WithServiceMaxRetries(retries)
	retriesOption(&testConfig)
	assert.Equal(t, retries, testConfig.MaxRetries)

	// Test WithServiceModelName
	modelName := "custom-model"
	modelOption := WithServiceModelName(modelName)

	// Reset test config
	testConfig = DefaultServiceConfig()
	modelOption(&testConfig)
	assert.Equal(t, modelName, testConfig.ModelName)

	// Test WithVerbose
	verboseOption := WithVerbose(true)

	// Reset test config
	testConfig = DefaultServiceConfig()
	verboseOption(&testConfig)
	assert.True(t, testConfig.Verbose)

	// Test applying multiple options
	testConfig = DefaultServiceConfig()
	retriesOption(&testConfig)
	modelOption(&testConfig)
	verboseOption(&testConfig)

	assert.Equal(t, retries, testConfig.MaxRetries)
	assert.Equal(t, modelName, testConfig.ModelName)
	assert.True(t, testConfig.Verbose)
}

func TestServiceConfigFunctions(t *testing.T) {
	// Test the config functions directly
	// Create base config
	config := DefaultServiceConfig()

	// Apply WithServiceMaxRetries
	maxRetriesOption := WithServiceMaxRetries(7)
	maxRetriesOption(&config)
	assert.Equal(t, 7, config.MaxRetries)

	// Apply WithServiceModelName
	modelNameOption := WithServiceModelName("functional-option-model")
	modelNameOption(&config)
	assert.Equal(t, "functional-option-model", config.ModelName)

	// Apply WithVerbose
	verboseOption := WithVerbose(true)
	verboseOption(&config)
	assert.True(t, config.Verbose)

	// Apply WithPromptTemplate
	promptTemplate := "Custom prompt template"
	promptTemplateOption := WithPromptTemplate(promptTemplate)
	promptTemplateOption(&config)
	assert.Equal(t, promptTemplate, config.PromptTemplate)

	// Test invalid option values (should still work)
	negativeRetries := WithServiceMaxRetries(-1)
	negativeRetries(&config)
	assert.Equal(t, -1, config.MaxRetries) // Should allow negative values even if they're invalid

	emptyModel := WithServiceModelName("")
	emptyModel(&config)
	assert.Equal(t, "", config.ModelName) // Should allow empty string even if it's invalid
}

// Test the end-to-end workflow from prompt creation to generation
func TestEndToEndWorkflow(t *testing.T) {
	// Skip this test by default as it's an integration test
	t.Skip("Skipping end-to-end workflow test - requires mocking many dependencies")

	// Would normally include:
	// 1. Setting up a mock client with expected behaviors for all operations
	// 2. Creating a service with the client
	// 3. Preparing test data for directory, files, etc.
	// 4. Calling GenerateGlanceMarkdown
	// 5. Verifying the result
}
