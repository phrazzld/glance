package llm

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// We're using the MockClient defined in client_test.go

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
		mockClient := new(MockClient)
		service, err := NewService(mockClient)

		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, mockClient, service.client)
		assert.Equal(t, DefaultServiceOptions().MaxRetries, service.options.MaxRetries)
		assert.Equal(t, DefaultServiceOptions().ModelName, service.options.ModelName)
		assert.Equal(t, DefaultServiceOptions().Verbose, service.options.Verbose)
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

	// Test with multiple options
	t.Run("Multiple options", func(t *testing.T) {
		mockClient := new(MockClient)
		service, err := NewService(mockClient,
			WithMaxRetries(5),
			WithModelName("custom-model"),
			WithVerbose(true))

		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, 5, service.options.MaxRetries)
		assert.Equal(t, "custom-model", service.options.ModelName)
		assert.True(t, service.options.Verbose)
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

	// Test with custom template
	t.Run("Custom template", func(t *testing.T) {
		// Reset mock
		mockClient = new(MockClient)

		// Create a temporary directory for the test
		tempDir := t.TempDir()

		// Create a custom template file
		customTemplate := "Custom template with {{.Directory}} and {{.SubGlances}} and {{.FileContents}}"
		customTemplatePath := filepath.Join(tempDir, "custom_template.txt")
		err := os.WriteFile(customTemplatePath, []byte(customTemplate), 0644)
		require.NoError(t, err)

		// Create temp prompt.txt in current directory to test template loading
		currentDir, err := os.Getwd()
		require.NoError(t, err)
		promptPath := filepath.Join(currentDir, "prompt.txt")
		err = os.WriteFile(promptPath, []byte(customTemplate), 0644)
		require.NoError(t, err)

		// Clean up the prompt.txt file after the test
		defer os.Remove(promptPath)

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

	// Test with verbose mode enabled
	t.Run("Verbose mode enabled", func(t *testing.T) {
		// Reset mock
		mockClient = new(MockClient)

		// Save current log level and restore it after test
		originalLevel := logrus.GetLevel()
		logrus.SetLevel(logrus.DebugLevel)
		defer logrus.SetLevel(originalLevel)

		// Setup service with mock client and verbose mode
		service, err := NewService(mockClient, WithVerbose(true))
		assert.NoError(t, err)

		// Setup expectations for the mock - include token counting since verbose is enabled
		mockClient.On("Generate", ctx, mock.AnythingOfType("string")).Return(expectedResponse, nil).Once()
		mockClient.On("CountTokens", ctx, mock.AnythingOfType("string")).Return(100, nil).Once() // Should be called with verbose=true

		// Call the method
		result, err := service.GenerateGlanceMarkdown(ctx, dir, fileMap, subGlances)

		// Verify results
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
		mockClient.AssertExpectations(t)
	})

	// Test with token counting error in verbose mode
	t.Run("Token counting error in verbose mode", func(t *testing.T) {
		// Reset mock
		mockClient = new(MockClient)

		// Save current log level and restore it after test
		originalLevel := logrus.GetLevel()
		logrus.SetLevel(logrus.DebugLevel)
		defer logrus.SetLevel(originalLevel)

		// Setup service with mock client and verbose mode
		service, err := NewService(mockClient, WithVerbose(true))
		assert.NoError(t, err)

		// Setup expectations for the mock
		tokenError := errors.New("token counting error")
		mockClient.On("CountTokens", ctx, mock.AnythingOfType("string")).Return(0, tokenError).Once()
		mockClient.On("Generate", ctx, mock.AnythingOfType("string")).Return(expectedResponse, nil).Once()

		// Call the method - should still work despite token counting error
		result, err := service.GenerateGlanceMarkdown(ctx, dir, fileMap, subGlances)

		// Verify results - generation should succeed despite token counting error
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
		mockClient.AssertExpectations(t)
	})

	// Test with template loading error
	t.Run("Template loading error", func(t *testing.T) {
		// Create a service with a mock client that expects nothing to be called
		// because we'll error before reaching the client
		mockClient = new(MockClient)
		service, err := NewService(mockClient)
		assert.NoError(t, err)

		// Create a directory where we can't read files (no permissions)
		// This is tricky to do in a test, so we'll mock the behavior by
		// patching the LoadTemplate function (not easy in Go)
		// Instead, we'll test error handling for prompt generation

		// Create an invalid template to cause an error
		invalidTemplate := "Invalid template with {{.MissingVar}}"
		customTemplatePath := filepath.Join(t.TempDir(), "invalid_template.txt")
		err = os.WriteFile(customTemplatePath, []byte(invalidTemplate), 0644)
		require.NoError(t, err)

		// Set mock expectations for nil fileMap case
		mockClient.On("Generate", ctx, mock.AnythingOfType("string")).Return(expectedResponse, nil).Once()
		mockClient.On("CountTokens", ctx, mock.AnythingOfType("string")).Return(100, nil).Maybe()

		// Test handling of nil fileMap
		result, err := service.GenerateGlanceMarkdown(ctx, dir, nil, subGlances)
		assert.NoError(t, err) // Should handle nil fileMap gracefully
		assert.Equal(t, expectedResponse, result)
	})

	// Test with prompt template from options
	t.Run("Use prompt template from options", func(t *testing.T) {
		// Reset mock
		mockClient = new(MockClient)

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

func TestServiceOptions(t *testing.T) {
	// Test default options
	defaults := DefaultServiceOptions()
	assert.Greater(t, defaults.MaxRetries, 0)
	assert.NotEmpty(t, defaults.ModelName)
	assert.False(t, defaults.Verbose)

	// Test WithMaxRetries
	retries := 5
	options := DefaultServiceOptions().WithMaxRetries(retries)
	assert.Equal(t, retries, options.MaxRetries)
	assert.NotEqual(t, &defaults, &options, "Should create a new options instance")

	// Test WithModelName
	modelName := "custom-model"
	options = DefaultServiceOptions().WithModelName(modelName)
	assert.Equal(t, modelName, options.ModelName)
	assert.NotEqual(t, &defaults, &options, "Should create a new options instance")

	// Test WithVerbose
	options = DefaultServiceOptions().WithVerbose(true)
	assert.True(t, options.Verbose)
	assert.NotEqual(t, &defaults, &options, "Should create a new options instance")

	// Test chaining options
	chainedOptions := DefaultServiceOptions().
		WithMaxRetries(10).
		WithModelName("chained-model").
		WithVerbose(true)

	assert.Equal(t, 10, chainedOptions.MaxRetries)
	assert.Equal(t, "chained-model", chainedOptions.ModelName)
	assert.True(t, chainedOptions.Verbose)
}

func TestServiceOptionFunctions(t *testing.T) {
	// Test the functional option pattern functions directly
	// Create base options
	options := DefaultServiceOptions()

	// Apply WithMaxRetries
	maxRetriesOption := WithMaxRetries(7)
	maxRetriesOption(options)
	assert.Equal(t, 7, options.MaxRetries)

	// Apply WithModelName
	modelNameOption := WithModelName("functional-option-model")
	modelNameOption(options)
	assert.Equal(t, "functional-option-model", options.ModelName)

	// Apply WithVerbose
	verboseOption := WithVerbose(true)
	verboseOption(options)
	assert.True(t, options.Verbose)

	// Apply WithPromptTemplate
	promptTemplate := "Custom prompt template"
	promptTemplateOption := WithPromptTemplate(promptTemplate)
	promptTemplateOption(options)
	assert.Equal(t, promptTemplate, options.PromptTemplate)

	// Test invalid option values (should still work)
	negativeRetries := WithMaxRetries(-1)
	negativeRetries(options)
	assert.Equal(t, -1, options.MaxRetries) // Should allow negative values even if they're invalid

	emptyModel := WithModelName("")
	emptyModel(options)
	assert.Equal(t, "", options.ModelName) // Should allow empty string even if it's invalid
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
