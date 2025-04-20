// Package mocks provides mock implementations for testing across the glance application.
package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockLLMServiceFactory is a mock implementation of a service factory.
// It returns predefined responses based on test expectations.
type MockLLMServiceFactory struct {
	mock.Mock
}

// CreateService mocks the service creation method.
func (m *MockLLMServiceFactory) CreateService(cfg interface{}) (interface{}, interface{}, error) {
	args := m.Called(cfg)
	return args.Get(0), args.Get(1), args.Error(2)
}
