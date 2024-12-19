package openai

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

func TestHandleMessage(t *testing.T) {
	t.Run("returns updated thread", func(t *testing.T) {
		is := is.New(t)

		// Create a test server that returns mock responses
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(newMockResponse())
		}))
		defer server.Close()

		var provider minds.ContentGenerator
		provider, err := NewProvider(WithBaseURL(server.URL))
		is.NoErr(err) // NewProvider should not return an error

		handler, ok := provider.(minds.ThreadHandler)
		is.True(ok) // provider should implement the ThreadHandler interface

		thread := minds.NewThreadContext(context.Background()).WithMessages(minds.Messages{
			{Role: minds.RoleUser, Content: "Hi"},
		})

		result, err := handler.HandleThread(thread, nil)
		is.NoErr(err) // HandleMessage should not return an error
		messages := result.Messages()
		is.Equal(len(messages), 2)
		is.Equal(messages[1].Role, minds.RoleAssistant)
		is.Equal(messages[1].Content, "Hello, world!")
	})

	t.Run("returns error on failure", func(t *testing.T) {
		is := is.New(t)
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(newMockResponse())
		}))
		defer server.Close()

		var provider minds.ContentGenerator
		provider, err := NewProvider(WithBaseURL(server.URL))
		is.NoErr(err) // NewProvider should not return an error

		handler, ok := provider.(minds.ThreadHandler)
		is.True(ok) // provider should implement the ThreadHandler interface

		thread := minds.NewThreadContext(ctx).WithMessages(minds.Messages{
			{Role: minds.RoleUser, Content: "Hi"},
		})

		_, err = handler.HandleThread(thread, nil)
		is.True(err != nil) // HandleMessage should return an error
		is.Equal(err.Error(), context.DeadlineExceeded.Error())
		cancel()
	})
}
