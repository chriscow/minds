package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/chriscow/minds"
	"github.com/matryer/is"
)

type mockProvider struct {
	response minds.Response
}

func (m *mockProvider) ModelName() string {
	return "mock-model"
}

func (m *mockProvider) GenerateContent(ctx context.Context, req minds.Request) (minds.Response, error) {
	return m.response, nil
}

func (m *mockProvider) Close() {
	// No-op
}

type mockHandler struct {
	name     string
	called   bool
	response minds.ThreadContext
	err      error
}

func (m *mockHandler) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	m.called = true
	if m.response != nil {
		return m.response, m.err
	}
	return tc, m.err
}

func (m *mockHandler) String() string {
	return m.name
}

type mockResponse struct {
	Content string
}

func (m mockResponse) String() string {
	return m.Content
}

func (m mockResponse) ToolCalls() []minds.ToolCall {
	return nil
}

func newMockTextResponse(content string) minds.Response {
	return mockResponse{
		Content: content,
	}
}

func newMockBoolResponse(content bool) minds.Response {
	resp := boolResp{Bool: content}
	data, _ := json.Marshal(resp)

	return mockResponse{
		Content: string(data),
	}
}

func TestSwitch(t *testing.T) {
	is := is.New(t)

	t.Run("matches first case", func(t *testing.T) {
		is := is.New(t)

		defaultHandler := &mockHandler{name: "default"}
		handler1 := &mockHandler{name: "handler1"}
		handler2 := &mockHandler{name: "handler2"}

		tc := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"type": "text"})

		sw := Switch("test",
			defaultHandler,
			SwitchCase{
				Condition: MetadataEquals{Key: "type", Value: "text"},
				Handler:   handler1,
			},
			SwitchCase{
				Condition: MetadataEquals{Key: "type", Value: "image"},
				Handler:   handler2,
			},
		)

		_, err := sw.HandleThread(tc, nil)
		is.NoErr(err)
		is.True(handler1.called)        // First handler should be called
		is.True(!handler2.called)       // Second handler should not be called
		is.True(!defaultHandler.called) // Default handler should not be called
	})

	t.Run("falls through to default", func(t *testing.T) {
		is := is.New(t)

		defaultHandler := &mockHandler{name: "default"}
		handler1 := &mockHandler{name: "handler1"}

		tc := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"type": "unknown"})

		sw := Switch("test",
			defaultHandler,
			SwitchCase{
				Condition: MetadataEquals{Key: "type", Value: "text"},
				Handler:   handler1,
			},
		)

		_, err := sw.HandleThread(tc, nil)
		is.NoErr(err)
		is.True(!handler1.called)      // Handler should not be called
		is.True(defaultHandler.called) // Default handler should be called
	})
}

func TestMetadataEquals(t *testing.T) {
	is := is.New(t)

	t.Run("matches existing key and value", func(t *testing.T) {
		is := is.New(t)

		cond := MetadataEquals{Key: "type", Value: "text"}
		tc := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"type": "text"})

		result, err := cond.Evaluate(tc)
		is.NoErr(err)
		is.True(result)
	})

	t.Run("non-matching value", func(t *testing.T) {
		is := is.New(t)

		cond := MetadataEquals{Key: "type", Value: "text"}
		tc := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"type": "image"})

		result, err := cond.Evaluate(tc)
		is.NoErr(err)
		is.True(!result)
	})

	t.Run("non-existing key", func(t *testing.T) {
		is := is.New(t)

		cond := MetadataEquals{Key: "missing", Value: "text"}
		tc := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"type": "text"})

		result, err := cond.Evaluate(tc)
		is.NoErr(err)
		is.True(!result)
	})
}

func TestLLMCondition(t *testing.T) {
	is := is.New(t)

	t.Run("positive response", func(t *testing.T) {
		is := is.New(t)

		provider := &mockProvider{
			response: newMockBoolResponse(true),
		}

		cond := LLMCondition{
			Generator: provider,
			Prompt:    "Is this positive?",
		}

		tc := minds.NewThreadContext(context.Background()).
			WithMessages(minds.Message{
				Role: minds.RoleUser, Content: "I love this!",
			})

		result, err := cond.Evaluate(tc)
		is.NoErr(err)
		is.True(result)
	})

	t.Run("negative response", func(t *testing.T) {
		is := is.New(t)

		provider := &mockProvider{
			response: newMockBoolResponse(false),
		}

		cond := LLMCondition{
			Generator: provider,
			Prompt:    "Is this positive?",
		}

		tc := minds.NewThreadContext(context.Background()).
			WithMessages(minds.Message{
				Role: minds.RoleUser, Content: "I don't like this.",
			})

		result, err := cond.Evaluate(tc)
		is.NoErr(err)
		is.True(!result)
	})
}
