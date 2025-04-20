// Package mocks provides mock implementations for testing across the glance application.
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// LLMClient is a mock implementation of a client for LLM services.
// It uses github.com/stretchr/testify/mock to provide a flexible mocking system.
type LLMClient struct {
	mock.Mock
}

// Generate mocks the Generate method that takes a prompt and returns generated text.
func (m *LLMClient) Generate(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
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
