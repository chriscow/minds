package deepseek

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chriscow/minds"
	"github.com/matryer/is"
)

func newMockResponse() ChatCompletionResponse {
	return ChatCompletionResponse{
		ID:      "54c8ec3a-3364-4e6a-9a5d-7794f90fab5e",
		Object:  "chat.completion",
		Created: 1735834884,
		Model:   "deepseek-chat",
		Choices: []Choice{
			{
				Index: 0,
				Message: minds.Message{
					Role:    "assistant",
					Content: "Hello, world!",
				},
				Logprobs:     nil,
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:          9,
			CompletionTokens:      11,
			TotalTokens:           20,
			PromptCacheHitTokens:  0,
			PromptCacheMissTokens: 9,
		},
		SystemFingerprint: "fp_f1afce2943",
	}
}

func TestProvider_CreateChatCompletion(t *testing.T) {
	is := is.New(t)

	// Create a test server that returns mock responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(newMockResponse())
	}))
	defer server.Close()

	provider, err := NewProvider(WithBaseURL(server.URL), WithAPIKey("test"))
	is.NoErr(err) // NewProvider should not return an error

	ctx := context.Background()
	req := minds.NewRequest(minds.Messages{
		{Role: minds.RoleUser, Content: "Hello!"},
	})

	resp, err := provider.GenerateContent(ctx, req)
	is.NoErr(err)        // createChatCompletion should not return an error
	is.True(resp != nil) // Response should not be nil
	text, ok := resp.Text()
	is.Equal(ok, true)              // Should be able to extract text
	is.Equal(text, "Hello, world!") // Ensure the mock response matches

	is.Equal(resp.Type(), minds.ResponseTypeText) // Ensure it is a text response
	is.Equal(resp.String(), "Hello, world!")      // Ensure the mock response matches
	_, ok = resp.ToolCalls()
	is.Equal(ok, false) // Ensure there are no tool calls
}

func TestProvider_CreateChatCompletion_Retry(t *testing.T) {
	is := is.New(t)

	// Create a test server that returns 429 and 503 status codes before returning a mock response
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(newMockResponse())
	}))
	defer server.Close()

	provider, err := NewProvider(WithBaseURL(server.URL), WithAPIKey("test"))
	is.NoErr(err) // NewProvider should not return an error

	ctx := context.Background()
	req := minds.NewRequest(minds.Messages{
		{Role: minds.RoleUser, Content: "Hello!"},
	})

	start := time.Now()
	resp, err := provider.GenerateContent(ctx, req)
	duration := time.Since(start)

	is.NoErr(err)        // createChatCompletion should not return an error
	is.True(resp != nil) // Response should not be nil
	text, ok := resp.Text()
	is.Equal(ok, true)                 // Should be able to extract text
	is.Equal(text, "Hello, world!")    // Ensure the mock response matches
	is.True(attempts == 3)             // Ensure the request was retried twice before succeeding
	is.True(duration >= 3*time.Second) // Ensure the retries included exponential backoff
}
