package handlers_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

func TestIf(t *testing.T) {
	t.Run("condition true calls true handler", func(t *testing.T) {
		is := is.New(t)

		trueHandler := &mockHandler{name: "true-handler"}
		fallbackHandler := &mockHandler{name: "fallback"}

		tc := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"type": "text"})

		ih := handlers.NewIf("test",
			handlers.MetadataEquals{Key: "type", Value: "text"},
			trueHandler,
			fallbackHandler,
		)

		_, err := ih.HandleThread(tc, nil)
		is.NoErr(err)
		is.Equal(1, trueHandler.Completed())     // True handler should be called
		is.Equal(0, fallbackHandler.Completed()) // Fallback handler should not be called
	})

	t.Run("condition false calls fallback handler", func(t *testing.T) {
		is := is.New(t)

		trueHandler := &mockHandler{name: "true-handler"}
		fallbackHandler := &mockHandler{name: "fallback"}

		tc := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"type": "other"})

		ih := handlers.NewIf("test",
			handlers.MetadataEquals{Key: "type", Value: "text"},
			trueHandler,
			fallbackHandler,
		)

		_, err := ih.HandleThread(tc, nil)
		is.NoErr(err)
		is.Equal(0, trueHandler.Completed())     // True handler should not be called
		is.Equal(1, fallbackHandler.Completed()) // Fallback handler should be called
	})

	t.Run("condition false with nil fallback", func(t *testing.T) {
		is := is.New(t)

		trueHandler := &mockHandler{name: "true-handler"}

		tc := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"type": "other"})

		ih := handlers.NewIf("test",
			handlers.MetadataEquals{Key: "type", Value: "text"},
			trueHandler,
			nil,
		)

		result, err := ih.HandleThread(tc, nil)
		is.NoErr(err)
		is.Equal(0, trueHandler.Completed()) // True handler should not be called
		is.Equal(result, tc)                 // Should return unmodified context
	})

	t.Run("handles error from condition", func(t *testing.T) {
		is := is.New(t)

		trueHandler := &mockHandler{name: "true-handler"}
		fallbackHandler := &mockHandler{name: "fallback"}

		tc := minds.NewThreadContext(context.Background())

		errCondition := &mockCondition{err: context.DeadlineExceeded}

		ih := handlers.NewIf("test",
			errCondition,
			trueHandler,
			fallbackHandler,
		)

		_, err := ih.HandleThread(tc, nil)
		is.True(err != nil)                      // Should return error
		is.Equal(0, trueHandler.Completed())     // True handler should not be called
		is.Equal(0, fallbackHandler.Completed()) // Fallback handler should not be called
	})

	t.Run("propagates handler error", func(t *testing.T) {
		is := is.New(t)

		// expectedErr := context.DeadlineExceeded
		trueHandler := &mockHandler{name: "true-handler", expectedErr: context.DeadlineExceeded}
		fallbackHandler := &mockHandler{name: "fallback"}

		tc := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"type": "text"})

		ih := handlers.NewIf("test",
			handlers.MetadataEquals{Key: "type", Value: "text"},
			trueHandler,
			fallbackHandler,
		)

		_, err := ih.HandleThread(tc, nil)
		is.True(err != nil)                      // Should return handler error
		is.Equal(1, trueHandler.Started())       // True handler should be called
		is.Equal(0, fallbackHandler.Completed()) // Fallback handler should not be called
	})

	t.Run("with middleware", func(t *testing.T) {
		is := is.New(t)

		trueHandler := &mockHandler{name: "true-handler"}
		m1 := &mockMiddleware{name: "Middleware1"}
		m2 := &mockMiddleware{name: "Middleware2"}

		tc := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"type": "text"})

		ih := handlers.NewIf("test",
			handlers.MetadataEquals{Key: "type", Value: "text"},
			trueHandler,
			nil,
		)
		ih.Use(m1, m2)

		_, err := ih.HandleThread(tc, nil)
		is.NoErr(err)
		is.Equal(1, m1.applied)    // Middleware should be applied once
		is.Equal(1, m2.applied)    // Middleware should be applied once
		is.Equal(1, m1.executions) // Middleware should execute once
		is.Equal(1, m2.executions) // Middleware should execute once
		is.Equal(1, trueHandler.Completed())
	})

	t.Run("middleware error", func(t *testing.T) {
		is := is.New(t)

		trueHandler := &mockHandler{name: "true-handler"}
		m1 := &mockMiddleware{name: "Middleware1", expectedErr: errMiddlewareFailed}

		tc := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"type": "text"})

		ih := handlers.NewIf("test",
			handlers.MetadataEquals{Key: "type", Value: "text"},
			trueHandler,
			nil,
		)
		ih.Use(m1)

		_, err := ih.HandleThread(tc, nil)
		is.True(err != nil)
		is.True(strings.Contains(err.Error(), "middleware failed"))
		is.Equal(0, trueHandler.Completed()) // Handler should not execute due to middleware error
	})

	t.Run("with middleware chain", func(t *testing.T) {
		is := is.New(t)

		trueHandler := &mockHandler{name: "true-handler"}
		m1 := &mockMiddleware{name: "Middleware1"}
		m2 := &mockMiddleware{name: "Middleware2"}

		tc := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"type": "text"})

		ih := handlers.NewIf("test",
			handlers.MetadataEquals{Key: "type", Value: "text"},
			trueHandler,
			nil,
		)
		chainedIf := ih.With(m1, m2)

		_, err := chainedIf.HandleThread(tc, nil)
		is.NoErr(err)
		is.Equal(1, m1.applied)    // Middleware should be applied once
		is.Equal(1, m2.applied)    // Middleware should be applied once
		is.Equal(1, m1.executions) // Middleware should execute once
		is.Equal(1, m2.executions) // Middleware should execute once
		is.Equal(1, trueHandler.Completed())
	})

	t.Run("string representation", func(t *testing.T) {
		is := is.New(t)

		trueHandler := &mockHandler{name: "true-handler"}
		ih := handlers.NewIf("test-name",
			handlers.MetadataEquals{Key: "type", Value: "text"},
			trueHandler,
			nil,
		)

		str := ih.String()
		is.True(strings.Contains(str, "If(test-name)"))
	})
}

func TestIf_MiddlewarePropagation(t *testing.T) {
	is := is.New(t)

	// Track execution order
	executionOrder := make([]string, 0)

	// Create handlers that record their execution
	trueHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		executionOrder = append(executionOrder, "true_handler")
		return tc, nil
	})

	fallbackHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		executionOrder = append(executionOrder, "fallback_handler")
		return tc, nil
	})

	// Create middleware that tracks execution
	mw := minds.MiddlewareFunc(func(next minds.ThreadHandler) minds.ThreadHandler {
		return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			// Get condition result from metadata if available
			meta := tc.Metadata()
			var conditionPart string
			if meta["type"] == "text" {
				conditionPart = "true"
			} else if meta["type"] == "other" {
				conditionPart = "false"
			} else {
				conditionPart = "outer"
			}

			executionOrder = append(executionOrder, fmt.Sprintf("middleware_start_%s", conditionPart))
			result, err := next.HandleThread(tc, nil)
			executionOrder = append(executionOrder, fmt.Sprintf("middleware_end_%s", conditionPart))
			return result, err
		})
	})

	// Create If handler with middleware
	ih := handlers.NewIf("test",
		handlers.MetadataEquals{Key: "type", Value: "text"},
		trueHandler,
		fallbackHandler,
	)
	ih.Use(mw)

	// Test true path
	tc := minds.NewThreadContext(context.Background()).
		WithMetadata(minds.Metadata{"type": "text"})

	_, err := ih.HandleThread(tc, nil)
	is.NoErr(err)

	// Verify execution order for true path
	expectedTrue := []string{
		"middleware_start_true", // Inner middleware starts for true handler
		"true_handler",          // True handler executes
		"middleware_end_true",   // Inner middleware ends
	}
	is.Equal(executionOrder, expectedTrue)

	// Reset execution order and test false path
	executionOrder = nil
	tc = minds.NewThreadContext(context.Background()).
		WithMetadata(minds.Metadata{"type": "other"})

	_, err = ih.HandleThread(tc, nil)
	is.NoErr(err)

	// Verify execution order for false path
	expectedFalse := []string{
		"middleware_start_false", // Inner middleware starts for fallback
		"fallback_handler",       // Fallback handler executes
		"middleware_end_false",   // Inner middleware ends
	}
	is.Equal(executionOrder, expectedFalse)
}
