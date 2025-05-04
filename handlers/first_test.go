package handlers_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

func TestFirst_FirstSuccessReturns(t *testing.T) {
	is := is.New(t)

	h1 := &mockHandler{name: "Handler1", sleep: 100 * time.Millisecond}
	h2 := &mockHandler{name: "Handler2", expectedErr: errHandlerFailed}
	h3 := &mockHandler{name: "Handler3", sleep: 50 * time.Millisecond}

	first := handlers.NewFirst("FirstSuccess", h1, h2, h3)
	tc := minds.NewThreadContext(context.Background())

	_, err := first.HandleThread(tc, nil)

	is.NoErr(err)
	is.Equal(1, h3.Completed())        // h3 should complete
	time.Sleep(200 * time.Millisecond) // Wait for other handlers
	is.Equal(1, h1.Started())          // h1 should start
	is.Equal(1, h2.Started())          // h2 should start
}

func TestFirst_AllHandlersFail(t *testing.T) {
	is := is.New(t)

	h1 := &mockHandler{name: "Handler1", expectedErr: errHandlerFailed}
	h2 := &mockHandler{name: "Handler2", expectedErr: errHandlerFailed}
	h3 := &mockHandler{name: "Handler3", expectedErr: errHandlerFailed}

	first := handlers.NewFirst("AllFail", h1, h2, h3)
	tc := minds.NewThreadContext(context.Background())

	_, err := first.HandleThread(tc, nil)

	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "all handlers failed"))
}

func TestFirst_ContextCancellation(t *testing.T) {
	is := is.New(t)

	h1 := &mockHandler{name: "Handler1", sleep: 200 * time.Millisecond}
	h2 := &mockHandler{name: "Handler2", sleep: 200 * time.Millisecond}

	first := handlers.NewFirst("ContextCancel", h1, h2)

	ctx, cancel := context.WithCancel(context.Background())
	tc := minds.NewThreadContext(ctx)

	cancel() // Cancel immediately instead of with delay

	_, err := first.HandleThread(tc, nil)
	is.True(err != nil)
}

func TestFirst_NoHandlers(t *testing.T) {
	is := is.New(t)

	first := handlers.NewFirst("Empty")
	tc := minds.NewThreadContext(context.Background())

	_, err := first.HandleThread(tc, nil)
	is.NoErr(err)
}

func TestFirst_NestedFirst(t *testing.T) {
	is := is.New(t)

	h1 := &mockHandler{name: "Handler1", expectedErr: errHandlerFailed}
	h2 := &mockHandler{name: "Handler2", sleep: 50 * time.Millisecond}
	h3 := &mockHandler{name: "Handler3", expectedErr: errHandlerFailed}

	nestedFirst := handlers.NewFirst("Nested", h2, h3)
	first := handlers.NewFirst("Outer", h1, nestedFirst)

	tc := minds.NewThreadContext(context.Background())

	_, err := first.HandleThread(tc, nil)
	is.NoErr(err)
}

func TestFirst_WithMiddleware(t *testing.T) {
	is := is.New(t)

	h1 := &mockHandler{name: "Handler1", sleep: 50 * time.Millisecond}
	h2 := &mockHandler{name: "Handler2", expectedErr: errHandlerFailed}

	m1 := &mockMiddleware{name: "Middleware1"}
	m2 := &mockMiddleware{name: "Middleware2"}

	first := handlers.NewFirst("WithMiddleware", h1, h2)
	first.Use(m1, m2)

	tc := minds.NewThreadContext(context.Background())
	_, err := first.HandleThread(tc, nil)

	is.NoErr(err)
	is.True(m1.executions == 2) // First middleware should execute
	is.True(m2.executions == 2) // Second middleware should execute
}

func TestFirst_MiddlewareError(t *testing.T) {
	is := is.New(t)

	h1 := &mockHandler{name: "Handler1"}
	m1 := &mockMiddleware{name: "Middleware1", expectedErr: errHandlerFailed}

	first := handlers.NewFirst("MiddlewareError", h1)
	first.Use(m1)

	tc := minds.NewThreadContext(context.Background())
	_, err := first.HandleThread(tc, nil)

	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "all handlers failed"))
}

func TestFirst_WithMiddlewareChain(t *testing.T) {
	is := is.New(t)

	h1 := &mockHandler{name: "Handler1"}
	m1 := &mockMiddleware{name: "Middleware1"}
	m2 := &mockMiddleware{name: "Middleware2"}

	first := handlers.NewFirst("WithMiddlewareChain", h1)
	chainedFirst := first.With(m1, m2)

	tc := minds.NewThreadContext(context.Background())
	_, err := chainedFirst.HandleThread(tc, nil)

	is.NoErr(err)
	is.True(m1.executions == 1) // First middleware should execute
	is.True(m2.executions == 1) // Second middleware should execute
}

func TestFirst_String(t *testing.T) {
	is := is.New(t)

	h1 := &mockHandler{name: "Handler1"}
	h2 := &mockHandler{name: "Handler2"}

	first := handlers.NewFirst("TestName", h1, h2)
	str := first.String()

	is.True(strings.Contains(str, "TestName"))
}

func TestFirst_MiddlewarePropagation(t *testing.T) {
	is := is.New(t)

	// Track execution order
	executionOrder := make([]string, 0)
	var mu sync.Mutex

	addExecution := func(s string) {
		mu.Lock()
		executionOrder = append(executionOrder, s)
		mu.Unlock()
	}

	// Create handlers that record their execution
	h1 := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		addExecution("handler_1")
		time.Sleep(50 * time.Millisecond)
		return tc, errHandlerFailed // First handler fails
	})

	h2 := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		addExecution("handler_2")
		time.Sleep(25 * time.Millisecond)
		return tc, nil // Second handler succeeds
	})

	// Create middleware that tracks execution with operation context
	mw := minds.MiddlewareFunc(func(next minds.ThreadHandler) minds.ThreadHandler {
		return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			meta := tc.Metadata()
			opType := "outer"
			if handlerName, ok := meta["handler_name"].(string); ok {
				opType = handlerName
			}

			addExecution(fmt.Sprintf("middleware_start_%s", opType))
			result, err := next.HandleThread(tc, nil)
			addExecution(fmt.Sprintf("middleware_end_%s", opType))

			return result, err
		})
	})

	// Create First handler with middleware
	first := handlers.NewFirst("test", h1, h2)
	first.Use(mw)

	// Execute handler
	tc := minds.NewThreadContext(context.Background())
	_, err := first.HandleThread(tc, nil)
	is.NoErr(err)

	// Allow time for all goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Get final execution order
	mu.Lock()
	finalOrder := make([]string, len(executionOrder))
	copy(finalOrder, executionOrder)
	mu.Unlock()

	// Verify execution order
	// We should see:
	// 1. Middleware starts
	// 2. Both handlers start with their own middleware (order may vary)
	// 3. Handler 2 completes with its middleware
	// 4. All middleware completes
	is.True(contains(finalOrder, "middleware_start_h1"))
	is.True(contains(finalOrder, "handler_1"))
	is.True(contains(finalOrder, "middleware_end_h1"))
	is.True(contains(finalOrder, "middleware_start_h2"))
	is.True(contains(finalOrder, "handler_2"))
	is.True(contains(finalOrder, "middleware_end_h2"))
}

// Helper functions for verifying execution order
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func TestFirst_ThreadContextIsolation(t *testing.T) {
	is := is.New(t)

	var mu sync.Mutex
	metadata := make(map[string]any)

	// Create handlers that modify thread context
	h1 := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		tc.SetKeyValue("handler", "h1")
		time.Sleep(50 * time.Millisecond)

		mu.Lock()
		metadata["h1"] = tc.Metadata()
		mu.Unlock()

		return tc, nil
	})

	h2 := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		tc.SetKeyValue("handler", "h2")

		mu.Lock()
		metadata["h2"] = tc.Metadata()
		mu.Unlock()

		return tc, nil
	})

	// Create and execute First handler
	first := handlers.NewFirst("test", h1, h2)
	tc := minds.NewThreadContext(context.Background())

	_, err := first.HandleThread(tc, nil)
	is.NoErr(err)

	// Allow time for all goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Verify thread context isolation
	mu.Lock()
	defer mu.Unlock()

	h1Meta := metadata["h1"].(minds.Metadata)
	h2Meta := metadata["h2"].(minds.Metadata)

	is.Equal(h1Meta["handler"], "h1")
	is.Equal(h2Meta["handler"], "h2")
}
