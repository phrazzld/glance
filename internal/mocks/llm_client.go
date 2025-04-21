// Package mocks provides mock implementations for testing across the glance application.
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// StreamChunk represents a piece of content from a streaming LLM response in mocks.
// This is a duplicate of llm.StreamChunk to avoid import cycles.
type StreamChunk struct {
	// Text contains the text content of this chunk, if any
	Text string

	// Error contains any error encountered during streaming
	Error error

	// Done indicates that this is the final chunk of the stream
	Done bool
}

// LLMClient is a mock implementation of a client for LLM services.
// It uses github.com/stretchr/testify/mock to provide a flexible mocking system.
// It implements the llm.Client interface for testing purposes.
type LLMClient struct {
	mock.Mock
}

// Generate mocks the Generate method that takes a prompt and returns generated text.
func (m *LLMClient) Generate(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

// GenerateStream mocks the GenerateStream method that takes a prompt and returns a channel of text chunks.
func (m *LLMClient) GenerateStream(ctx context.Context, prompt string) (<-chan StreamChunk, error) {
	args := m.Called(ctx, prompt)

	// Get first argument as interface{} and convert to the expected channel type
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(<-chan StreamChunk), args.Error(1)
}

// CountTokens mocks the method that counts tokens in a prompt.
func (m *LLMClient) CountTokens(ctx context.Context, prompt string) (int, error) {
	args := m.Called(ctx, prompt)
	return args.Int(0), args.Error(1)
}

// Close mocks the method that releases resources.
func (m *LLMClient) Close() {
	m.Called()
}
