package llm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"glance/internal/mocks"
)

func TestNewFallbackClientValidation(t *testing.T) {
	mockClient := new(mocks.LLMClient)
	adapter := NewMockClientAdapter(mockClient)

	t.Run("rejects empty tier list", func(t *testing.T) {
		client, err := NewFallbackClient(nil, 1)
		assert.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("rejects negative retries", func(t *testing.T) {
		client, err := NewFallbackClient([]FallbackTier{{Name: "t1", Client: adapter}}, -1)
		assert.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("rejects nil tier client", func(t *testing.T) {
		client, err := NewFallbackClient([]FallbackTier{{Name: "t1", Client: nil}}, 1)
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestFallbackClientGenerate(t *testing.T) {
	ctx := context.Background()
	prompt := "test prompt"

	t.Run("uses primary tier on success", func(t *testing.T) {
		primaryMock := new(mocks.LLMClient)
		secondaryMock := new(mocks.LLMClient)

		primary := NewMockClientAdapter(primaryMock)
		secondary := NewMockClientAdapter(secondaryMock)

		primaryMock.On("Generate", ctx, prompt).Return("ok-primary", nil).Once()
		primaryMock.On("Close").Return().Once()
		secondaryMock.On("Close").Return().Once()

		client, err := NewFallbackClientWithBackoff(
			[]FallbackTier{
				{Name: "primary", Client: primary},
				{Name: "secondary", Client: secondary},
			},
			1,
			time.Millisecond,
			time.Millisecond,
		)
		assert.NoError(t, err)

		out, genErr := client.Generate(ctx, prompt)
		assert.NoError(t, genErr)
		assert.Equal(t, "ok-primary", out)

		client.Close()
		primaryMock.AssertExpectations(t)
		secondaryMock.AssertExpectations(t)
	})

	t.Run("retries inside same tier before success", func(t *testing.T) {
		primaryMock := new(mocks.LLMClient)
		secondaryMock := new(mocks.LLMClient)

		primary := NewMockClientAdapter(primaryMock)
		secondary := NewMockClientAdapter(secondaryMock)

		primaryMock.
			On("Generate", ctx, prompt).
			Return("", errors.New("temporary error")).
			Once()
		primaryMock.
			On("Generate", ctx, prompt).
			Return("ok-after-retry", nil).
			Once()
		primaryMock.On("Close").Return().Once()
		secondaryMock.On("Close").Return().Once()

		client, err := NewFallbackClientWithBackoff(
			[]FallbackTier{
				{Name: "primary", Client: primary},
				{Name: "secondary", Client: secondary},
			},
			1,
			time.Millisecond,
			time.Millisecond,
		)
		assert.NoError(t, err)

		out, genErr := client.Generate(ctx, prompt)
		assert.NoError(t, genErr)
		assert.Equal(t, "ok-after-retry", out)

		client.Close()
		primaryMock.AssertExpectations(t)
		secondaryMock.AssertExpectations(t)
	})

	t.Run("falls back when primary tier exhausts", func(t *testing.T) {
		primaryMock := new(mocks.LLMClient)
		secondaryMock := new(mocks.LLMClient)

		primary := NewMockClientAdapter(primaryMock)
		secondary := NewMockClientAdapter(secondaryMock)

		primaryMock.
			On("Generate", ctx, prompt).
			Return("", errors.New("primary down")).
			Times(2)
		secondaryMock.
			On("Generate", ctx, prompt).
			Return("ok-secondary", nil).
			Once()
		primaryMock.On("Close").Return().Once()
		secondaryMock.On("Close").Return().Once()

		client, err := NewFallbackClientWithBackoff(
			[]FallbackTier{
				{Name: "primary", Client: primary},
				{Name: "secondary", Client: secondary},
			},
			1, // 1 retry => 2 attempts per tier
			time.Millisecond,
			time.Millisecond,
		)
		assert.NoError(t, err)

		out, genErr := client.Generate(ctx, prompt)
		assert.NoError(t, genErr)
		assert.Equal(t, "ok-secondary", out)

		client.Close()
		primaryMock.AssertExpectations(t)
		secondaryMock.AssertExpectations(t)
	})

	t.Run("returns error when all tiers fail", func(t *testing.T) {
		primaryMock := new(mocks.LLMClient)
		secondaryMock := new(mocks.LLMClient)

		primary := NewMockClientAdapter(primaryMock)
		secondary := NewMockClientAdapter(secondaryMock)

		primaryMock.
			On("Generate", ctx, prompt).
			Return("", errors.New("primary down")).
			Times(2)
		secondaryMock.
			On("Generate", ctx, prompt).
			Return("", errors.New("secondary down")).
			Times(2)
		primaryMock.On("Close").Return().Once()
		secondaryMock.On("Close").Return().Once()

		client, err := NewFallbackClientWithBackoff(
			[]FallbackTier{
				{Name: "primary", Client: primary},
				{Name: "secondary", Client: secondary},
			},
			1, // 1 retry => 2 attempts per tier
			time.Millisecond,
			time.Millisecond,
		)
		assert.NoError(t, err)

		out, genErr := client.Generate(ctx, prompt)
		assert.Error(t, genErr)
		assert.Empty(t, out)

		client.Close()
		primaryMock.AssertExpectations(t)
		secondaryMock.AssertExpectations(t)
	})
}

func TestFallbackClientCountTokens(t *testing.T) {
	ctx := context.Background()
	prompt := "token prompt"

	primaryMock := new(mocks.LLMClient)
	secondaryMock := new(mocks.LLMClient)
	primary := NewMockClientAdapter(primaryMock)
	secondary := NewMockClientAdapter(secondaryMock)

	primaryMock.
		On("CountTokens", ctx, prompt).
		Return(0, errors.New("count failed")).
		Once()
	secondaryMock.
		On("CountTokens", ctx, prompt).
		Return(77, nil).
		Once()
	primaryMock.On("Close").Return().Once()
	secondaryMock.On("Close").Return().Once()

	client, err := NewFallbackClientWithBackoff(
		[]FallbackTier{
			{Name: "primary", Client: primary},
			{Name: "secondary", Client: secondary},
		},
		1,
		time.Millisecond,
		time.Millisecond,
	)
	assert.NoError(t, err)

	count, countErr := client.CountTokens(ctx, prompt)
	assert.NoError(t, countErr)
	assert.Equal(t, 77, count)

	client.Close()
	primaryMock.AssertExpectations(t)
	secondaryMock.AssertExpectations(t)
}

func TestFallbackClientGenerateStream(t *testing.T) {
	ctx := context.Background()
	prompt := "stream prompt"

	primaryMock := new(mocks.LLMClient)
	secondaryMock := new(mocks.LLMClient)
	primary := NewMockClientAdapter(primaryMock)
	secondary := NewMockClientAdapter(secondaryMock)

	primaryMock.
		On("GenerateStream", ctx, prompt).
		Return((<-chan mocks.StreamChunk)(nil), errors.New("stream init failed")).
		Once()

	stream := make(chan mocks.StreamChunk, 2)
	stream <- mocks.StreamChunk{Text: "hello"}
	stream <- mocks.StreamChunk{Done: true}
	close(stream)

	secondaryMock.
		On("GenerateStream", ctx, prompt).
		Return((<-chan mocks.StreamChunk)(stream), nil).
		Once()
	primaryMock.On("Close").Return().Once()
	secondaryMock.On("Close").Return().Once()

	client, err := NewFallbackClientWithBackoff(
		[]FallbackTier{
			{Name: "primary", Client: primary},
			{Name: "secondary", Client: secondary},
		},
		1,
		time.Millisecond,
		time.Millisecond,
	)
	assert.NoError(t, err)

	ch, streamErr := client.GenerateStream(ctx, prompt)
	assert.NoError(t, streamErr)

	var got []string
	for chunk := range ch {
		if chunk.Text != "" {
			got = append(got, chunk.Text)
		}
		if chunk.Done {
			break
		}
	}
	assert.Equal(t, []string{"hello"}, got)

	client.Close()
	primaryMock.AssertExpectations(t)
	secondaryMock.AssertExpectations(t)
}

func TestSleepWithContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sleepWithContext(ctx, 10*time.Millisecond)
	assert.Error(t, err)
}

func TestFallbackClientContextCancelDuringRetry(t *testing.T) {
	prompt := "test prompt"

	primaryMock := new(mocks.LLMClient)
	primary := NewMockClientAdapter(primaryMock)

	// First attempt fails, triggering a backoff sleep
	primaryMock.
		On("Generate", mock.Anything, prompt).
		Return("", errors.New("transient")).
		Once()
	primaryMock.On("Close").Return().Once()

	client, err := NewFallbackClientWithBackoff(
		[]FallbackTier{{Name: "primary", Client: primary}},
		2,           // 2 retries = 3 attempts
		time.Second, // long backoff so cancel fires during sleep
		5*time.Second,
	)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay — during the backoff sleep
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, genErr := client.Generate(ctx, prompt)
	assert.Error(t, genErr)
	assert.ErrorIs(t, genErr, context.Canceled)

	client.Close()
	primaryMock.AssertExpectations(t)
}

func TestFallbackClientBackoffCap(t *testing.T) {
	clientIface, err := NewFallbackClientWithBackoff(
		[]FallbackTier{{Name: "tier", Client: NewMockClientAdapter(new(mocks.LLMClient))}},
		5,
		2*time.Millisecond,
		3*time.Millisecond,
	)
	assert.NoError(t, err)

	client, ok := clientIface.(*FallbackClient)
	assert.True(t, ok)

	attemptOne := ExponentialBackoff(1, client.baseBackoff, client.maxBackoff)
	assert.GreaterOrEqual(t, attemptOne, 1600*time.Microsecond)
	assert.LessOrEqual(t, attemptOne, 2400*time.Microsecond)

	capped := ExponentialBackoff(3, client.baseBackoff, client.maxBackoff)
	assert.GreaterOrEqual(t, capped, 2400*time.Microsecond)
	assert.LessOrEqual(t, capped, 3*time.Millisecond)
}
