package main

import (
	"github.com/stretchr/testify/mock"

	"glance/config"
	"glance/llm"
)

// mockLLMServiceFactory implements LLMServiceFactory for tests
type mockLLMServiceFactory struct {
	mock.Mock
}

// CreateService implements LLMServiceFactory.CreateService for tests
func (m *mockLLMServiceFactory) CreateService(cfg *config.Config) (llm.Client, *llm.Service, error) {
	args := m.Called(cfg)

	var client llm.Client
	if args.Get(0) != nil {
		client = args.Get(0).(llm.Client)
	}

	var service *llm.Service
	if args.Get(1) != nil {
		service = args.Get(1).(*llm.Service)
	}

	return client, service, args.Error(2)
}

// Helper method to create a mock factory with predefined returns
func newMockLLMServiceFactory(client llm.Client, service *llm.Service, err error) *mockLLMServiceFactory {
	factory := new(mockLLMServiceFactory)
	factory.On("CreateService", mock.Anything).Return(client, service, err)
	return factory
}
