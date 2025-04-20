package llm

import (
	"github.com/stretchr/testify/mock"
)

// mockClientFactory implements ClientFactory for tests
type mockClientFactory struct {
	mock.Mock
}

// CreateClient implements ClientFactory.CreateClient for tests
func (m *mockClientFactory) CreateClient(apiKey string, options ...ClientOption) (Client, error) {
	args := m.Called(apiKey, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(Client), args.Error(1)
}

// Helper method to mock successful client creation
func newMockClientFactory(client Client, err error) *mockClientFactory {
	factory := new(mockClientFactory)
	factory.On("CreateClient", mock.Anything, mock.Anything).Return(client, err)
	return factory
}
