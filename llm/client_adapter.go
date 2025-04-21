// Package llm provides abstractions and implementations for interacting with
// Large Language Model APIs in the glance application.
package llm

import (
	"context"

	"glance/internal/mocks"
)

// MockClientAdapter adapts a mocks.LLMClient to match the llm.Client interface.
// This is needed to work around the import cycle between the llm and mocks packages.
type MockClientAdapter struct {
	Mock *mocks.LLMClient
}

// Generate delegates to the mock client's Generate method.
func (a *MockClientAdapter) Generate(ctx context.Context, prompt string) (string, error) {
	return a.Mock.Generate(ctx, prompt)
}

// GenerateStream adapts the mock client's GenerateStream method to return the expected type.
func (a *MockClientAdapter) GenerateStream(ctx context.Context, prompt string) (<-chan StreamChunk, error) {
	mockChan, err := a.Mock.GenerateStream(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Create a channel of the correct type
	resultChan := make(chan StreamChunk)

	// Start a goroutine to convert from mock.StreamChunk to llm.StreamChunk
	go func() {
		defer close(resultChan)
		for mockChunk := range mockChan {
			// Convert mock chunk to llm chunk
			resultChan <- StreamChunk{
				Text:  mockChunk.Text,
				Error: mockChunk.Error,
				Done:  mockChunk.Done,
			}
		}
	}()

	return resultChan, nil
}

// CountTokens delegates to the mock client's CountTokens method.
func (a *MockClientAdapter) CountTokens(ctx context.Context, prompt string) (int, error) {
	return a.Mock.CountTokens(ctx, prompt)
}

// Close delegates to the mock client's Close method.
func (a *MockClientAdapter) Close() {
	a.Mock.Close()
}

// NewMockClientAdapter creates a new adapter for a mock client.
func NewMockClientAdapter(mockClient *mocks.LLMClient) Client {
	return &MockClientAdapter{
		Mock: mockClient,
	}
}
