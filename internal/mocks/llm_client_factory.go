// Package mocks provides mock implementations for testing across the glance application.
package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockLLMClientFactory is a mock implementation of a client factory.
// It returns predefined responses based on test expectations.
type MockLLMClientFactory struct {
	mock.Mock
}

// CreateClient mocks the client creation method.
func (m *MockLLMClientFactory) CreateClient(apiKey string, options interface{}) (interface{}, error) {
	args := m.Called(apiKey, options)
	return args.Get(0), args.Error(1)
}
