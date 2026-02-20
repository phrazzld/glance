package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewOpenRouterClientValidation(t *testing.T) {
	client, err := NewOpenRouterClient("")
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestOpenRouterClientGenerateSuccessStringContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": "openrouter generated text",
					},
				},
			},
		})
	}))
	defer server.Close()

	clientIface, err := NewOpenRouterClient(
		"test-key",
		WithModelName("x-ai/grok-4.1-fast"),
		WithMaxRetries(1),
	)
	assert.NoError(t, err)

	client, ok := clientIface.(*OpenRouterClient)
	assert.True(t, ok)
	client.baseURL = server.URL

	out, genErr := client.Generate(context.Background(), "test prompt")
	assert.NoError(t, genErr)
	assert.Equal(t, "openrouter generated text", out)
}

func TestOpenRouterClientGenerateSuccessArrayContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": []map[string]any{
							{"type": "text", "text": "hello "},
							{"type": "text", "text": "world"},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	clientIface, err := NewOpenRouterClient(
		"test-key",
		WithModelName("x-ai/grok-4.1-fast"),
		WithMaxRetries(1),
	)
	assert.NoError(t, err)

	client := clientIface.(*OpenRouterClient)
	client.baseURL = server.URL

	out, genErr := client.Generate(context.Background(), "test prompt")
	assert.NoError(t, genErr)
	assert.Equal(t, "hello world", out)
}

func TestOpenRouterClientGenerateHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "rate limited",
			},
		})
	}))
	defer server.Close()

	clientIface, err := NewOpenRouterClient(
		"test-key",
		WithModelName("x-ai/grok-4.1-fast"),
		WithMaxRetries(1),
	)
	assert.NoError(t, err)

	client := clientIface.(*OpenRouterClient)
	client.baseURL = server.URL

	out, genErr := client.Generate(context.Background(), "test prompt")
	assert.Error(t, genErr)
	assert.Empty(t, out)
	assert.Contains(t, genErr.Error(), "429")
}

func TestOpenRouterClientGenerateStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": "stream content",
					},
				},
			},
		})
	}))
	defer server.Close()

	clientIface, err := NewOpenRouterClient(
		"test-key",
		WithModelName("x-ai/grok-4.1-fast"),
		WithMaxRetries(1),
	)
	assert.NoError(t, err)

	client := clientIface.(*OpenRouterClient)
	client.baseURL = server.URL

	ch, streamErr := client.GenerateStream(context.Background(), "test prompt")
	assert.NoError(t, streamErr)

	var gotText string
	done := false
	for chunk := range ch {
		if chunk.Text != "" {
			gotText += chunk.Text
		}
		if chunk.Done {
			done = true
		}
	}

	assert.Equal(t, "stream content", gotText)
	assert.True(t, done)
}

func TestOpenRouterClientCountTokensUnsupported(t *testing.T) {
	clientIface, err := NewOpenRouterClient(
		"test-key",
		WithModelName("x-ai/grok-4.1-fast"),
	)
	assert.NoError(t, err)

	count, countErr := clientIface.CountTokens(context.Background(), "prompt")
	assert.Error(t, countErr)
	assert.Equal(t, 0, count)
}

func TestOpenRouterClientRespectsTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": "late response",
					},
				},
			},
		})
	}))
	defer server.Close()

	clientIface, err := NewOpenRouterClient(
		"test-key",
		WithModelName("x-ai/grok-4.1-fast"),
		WithTimeout(1), // one second client timeout, but request context below is shorter
		WithMaxRetries(1),
	)
	assert.NoError(t, err)

	client := clientIface.(*OpenRouterClient)
	client.baseURL = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, genErr := client.Generate(ctx, "test prompt")
	assert.Error(t, genErr)
}
