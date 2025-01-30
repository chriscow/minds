package handlers_test

import (
	"context"
	"testing"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

func TestSwitch(t *testing.T) {
	is := is.New(t)

	t.Run("matches first case", func(t *testing.T) {
		is := is.New(t)

		defaultHandler := &mockHandler{name: "default"}
		handler1 := &mockHandler{name: "handler1"}
		handler2 := &mockHandler{name: "handler2"}

		tc := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"type": "text"})

		sw := handlers.NewSwitch("test",
			defaultHandler,
			handlers.SwitchCase{
				Condition: handlers.MetadataEquals{Key: "type", Value: "text"},
				Handler:   handler1,
			},
			handlers.SwitchCase{
				Condition: handlers.MetadataEquals{Key: "type", Value: "image"},
				Handler:   handler2,
			},
		)

		_, err := sw.HandleThread(tc, nil)
		is.NoErr(err)
		is.True(1 == handler1.Completed())     // First handler should be called
		is.True(0 == handler2.Started())       // Second handler should not be called
		is.True(0 == defaultHandler.Started()) // Default handler should not be called
	})

	t.Run("falls through to default", func(t *testing.T) {
		is := is.New(t)

		defaultHandler := &mockHandler{name: "default"}
		handler1 := &mockHandler{name: "handler1"}

		tc := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"type": "unknown"})

		sw := handlers.NewSwitch("test",
			defaultHandler,
			handlers.SwitchCase{
				Condition: handlers.MetadataEquals{Key: "type", Value: "text"},
				Handler:   handler1,
			},
		)

		_, err := sw.HandleThread(tc, nil)
		is.NoErr(err)
		is.True(0 == handler1.Started())         // Handler should not be called
		is.True(1 == defaultHandler.Completed()) // Default handler should be called
	})
}

func Testhandlers_MetadataEquals(t *testing.T) {
	is := is.New(t)

	t.Run("matches existing key and value", func(t *testing.T) {
		is := is.New(t)

		cond := handlers.MetadataEquals{Key: "type", Value: "text"}
		tc := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"type": "text"})

		result, err := cond.Evaluate(tc)
		is.NoErr(err)
		is.True(result)
	})

	t.Run("non-matching value", func(t *testing.T) {
		is := is.New(t)

		cond := handlers.MetadataEquals{Key: "type", Value: "text"}
		tc := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"type": "image"})

		result, err := cond.Evaluate(tc)
		is.NoErr(err)
		is.True(!result)
	})

	t.Run("non-existing key", func(t *testing.T) {
		is := is.New(t)

		cond := handlers.MetadataEquals{Key: "missing", Value: "text"}
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

		cond := handlers.LLMCondition{
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

		cond := handlers.LLMCondition{
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

// switch_test.go

func TestSwitch_MiddlewarePropagation(t *testing.T) {
	is := is.New(t)

	// Create handlers that support middleware
	handler1 := &mockMiddlewareHandler{mockHandler: newMockHandler("handler1")}
	handler2 := &mockMiddlewareHandler{mockHandler: newMockHandler("handler2")}
	defaultHandler := &mockMiddlewareHandler{mockHandler: newMockHandler("default")}

	// Create middleware that tracks execution
	executionOrder := make([]string, 0)
	mw := minds.MiddlewareFunc(func(next minds.ThreadHandler) minds.ThreadHandler {
		return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			executionOrder = append(executionOrder, "middleware_"+next.(*mockMiddlewareHandler).name)
			result, err := next.HandleThread(tc, nil)
			if err == nil {
				executionOrder = append(executionOrder, "middleware_end_"+next.(*mockMiddlewareHandler).name)
			}
			return result, err
		})
	})

	// Create switch with cases
	sw := handlers.NewSwitch("test",
		defaultHandler,
		handlers.SwitchCase{
			Condition: handlers.MetadataEquals{Key: "type", Value: "case1"},
			Handler:   handler1,
		},
		handlers.SwitchCase{
			Condition: handlers.MetadataEquals{Key: "type", Value: "case2"},
			Handler:   handler2,
		},
	)
	sw.Use(mw)

	// Test case 1 handler
	tc := minds.NewThreadContext(context.Background()).
		WithMetadata(minds.Metadata{"type": "case1"})

	_, err := sw.HandleThread(tc, nil)
	is.NoErr(err)

	// Verify middleware was applied
	is.Equal(len(handler1.middleware), 1)

	// Verify execution order
	expected := []string{
		"middleware_handler1",
		"middleware_end_handler1",
	}
	is.Equal(executionOrder, expected)

	// Test default handler
	executionOrder = nil
	tc = minds.NewThreadContext(context.Background()).
		WithMetadata(minds.Metadata{"type": "unknown"})

	_, err = sw.HandleThread(tc, nil)
	is.NoErr(err)

	// Verify middleware was applied to default handler
	is.Equal(len(defaultHandler.middleware), 1)

	// Verify execution order
	expected = []string{
		"middleware_default",
		"middleware_end_default",
	}
	is.Equal(executionOrder, expected)
}
