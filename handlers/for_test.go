package handlers_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

func TestFor_ContextCanceled(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler", sleep: 100 * time.Millisecond}
	loop := handlers.NewFor("test", 2, handler, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tc := minds.NewThreadContext(ctx)
	_, err := loop.HandleThread(tc, nil)

	is.True(err != nil) // For should return an error when context is canceled
	is.True(errors.Is(err, context.Canceled))
	is.Equal(0, handler.Started())
	is.Equal(0, handler.Completed())
}

func TestFor_FixedIterations(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	final := &mockHandler{name: "final"}
	iterations := 3

	loop := handlers.NewFor("test", iterations, handler, nil)
	tc := minds.NewThreadContext(context.Background())

	_, err := loop.HandleThread(tc, final)

	is.NoErr(err)
	is.Equal(iterations, handler.Completed())
	is.Equal(1, final.Completed())
}

func TestFor_WithError(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler", expectedErr: errHandlerFailed}
	final := &mockHandler{name: "final"}

	loop := handlers.NewFor("test", 3, handler, nil)
	tc := minds.NewThreadContext(context.Background())

	_, err := loop.HandleThread(tc, final)

	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "handler failed"))
	is.True(strings.Contains(err.Error(), " iteration 0 failed"))
	is.Equal(1, handler.Started())
	is.Equal(0, final.Completed())
}

func TestFor_MultipleHandlers(t *testing.T) {
	is := is.New(t)
	handler1 := &mockHandler{name: "handler1"}
	handler2 := &mockHandler{name: "handler2"}
	iterations := 2

	sequence := handlers.NewSequence("sequence", handler1, handler2)
	loop := handlers.NewFor("test", iterations, sequence, nil)
	tc := minds.NewThreadContext(context.Background())

	_, err := loop.HandleThread(tc, nil)

	is.NoErr(err)
	is.Equal(iterations, handler1.Completed())
	is.Equal(iterations, handler2.Completed())
}

func TestFor_InfiniteStop(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	loop := handlers.NewFor("test", 0, handler, nil)
	tc := minds.NewThreadContext(ctx)

	go func() {
		time.Sleep(100 * time.Millisecond) // Let it run a few iterations
		cancel()
		close(done)
	}()

	_, err := loop.HandleThread(tc, nil)
	<-done

	is.True(err != nil)
	is.True(errors.Is(err, context.Canceled))
	is.True(handler.Completed() > 0)
}

func TestFor_WithContinueFunction(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	final := &mockHandler{name: "final"}

	continueFn := func(tc minds.ThreadContext, iter int) bool {
		return iter < 2 // Stop after 2 iterations
	}

	loop := handlers.NewFor("test", 5, handler, continueFn)
	tc := minds.NewThreadContext(context.Background())

	_, err := loop.HandleThread(tc, final)

	is.NoErr(err)
	is.Equal(2, handler.Started()) // Should stop at 2 despite 5 iterations specified
	is.Equal(1, final.Completed()) // Should have completed 1 time
}

func TestFor_InfiniteWithContinueFunction(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}

	continueFn := func(tc minds.ThreadContext, i int) bool {
		return i < 3
	}

	loop := handlers.NewFor("test", 0, handler, continueFn)
	tc := minds.NewThreadContext(context.Background())

	_, err := loop.HandleThread(tc, nil)

	is.NoErr(err)
	is.Equal(3, handler.Completed())
}

func TestFor_WithMiddleware(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	m1 := &mockMiddleware{name: "Middleware1"}
	m2 := &mockMiddleware{name: "Middleware2"}

	loop := handlers.NewFor("test", 3, handler, nil)
	loop.Use(m1, m2)

	tc := minds.NewThreadContext(context.Background())
	_, err := loop.HandleThread(tc, nil)

	is.NoErr(err)
	is.Equal(3, m1.applied)    // Middleware should be applied 4 times, once for handler and 3 iterations
	is.Equal(3, m2.applied)    // Middleware should be applied 4 times, once for handler and 3 iterations
	is.Equal(3, m1.executions) // Middleware should execute 4 times, once for handler and 3 iterations
	is.Equal(3, m2.executions) // Middleware should execute 4 times, once for handler and 3 iterations
	is.Equal(3, handler.Completed())
}

func TestFor_MiddlewareError(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	m1 := &mockMiddleware{name: "Middleware1", expectedErr: errMiddlewareFailed}

	loop := handlers.NewFor("test", 3, handler, nil)
	loop.Use(m1)

	tc := minds.NewThreadContext(context.Background())
	_, err := loop.HandleThread(tc, nil)

	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "middleware failed"))
	is.Equal(0, handler.Completed()) // Handler should not execute due to middleware error
}

func TestFor_String(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	loop := handlers.NewFor("TestName", 5, handler, nil)

	str := loop.String()
	is.True(strings.Contains(str, "For(TestName"))
	is.True(strings.Contains(str, "5 iterations"))
}

func TestFor_MiddlewarePropagation(t *testing.T) {
	is := is.New(t)

	// Track execution order
	executionOrder := make([]string, 0)

	// Create handler that records its execution
	handler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		meta := tc.Metadata()
		iteration := meta["iteration"].(int)
		executionOrder = append(executionOrder, fmt.Sprintf("handler_%d", iteration))
		return tc, nil
	})

	// Create middleware that tracks execution with context
	mw := minds.MiddlewareFunc(func(next minds.ThreadHandler) minds.ThreadHandler {
		return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			// Determine operation type from metadata
			meta := tc.Metadata()
			var opType string
			if iteration, ok := meta["iteration"].(int); ok {
				opType = fmt.Sprintf("iter_%d", iteration)
			} else {
				opType = "outer"
			}

			executionOrder = append(executionOrder, fmt.Sprintf("middleware_start_%s", opType))
			result, err := next.HandleThread(tc, nil)
			if err == nil {
				executionOrder = append(executionOrder, fmt.Sprintf("middleware_end_%s", opType))
			}
			return result, err
		})
	})

	// Create For handler with middleware
	loop := handlers.NewFor("test", 3, handler, nil)
	loop.Use(mw)

	// Execute handler
	tc := minds.NewThreadContext(context.Background())
	_, err := loop.HandleThread(tc, nil)
	is.NoErr(err)

	// Verify execution order
	expected := []string{
		"middleware_start_iter_0", // First iteration middleware starts
		"handler_0",               // First iteration executes
		"middleware_end_iter_0",   // First iteration middleware ends
		"middleware_start_iter_1", // Second iteration middleware starts
		"handler_1",               // Second iteration executes
		"middleware_end_iter_1",   // Second iteration middleware ends
		"middleware_start_iter_2", // Third iteration middleware starts
		"handler_2",               // Third iteration executes
		"middleware_end_iter_2",   // Third iteration middleware ends
	}

	is.Equal(executionOrder, expected)
}

func TestFor_ThreadContextIsolation(t *testing.T) {
	is := is.New(t)

	metadata := make([]minds.Metadata, 0)
	mu := sync.Mutex{}

	// Create handler that modifies thread context
	handler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		tc.SetKeyValue("iteration", tc.Metadata()["iteration"].(int)+1)

		mu.Lock()
		metadata = append(metadata, tc.Metadata())
		mu.Unlock()

		return tc, nil
	})

	// Create and execute For handler
	tc := minds.NewThreadContext(context.Background())
	tc.SetKeyValue("iteration", 0)

	loop := handlers.NewFor("test", 3, handler, nil)
	_, err := loop.HandleThread(tc, nil)
	is.NoErr(err)

	// Verify thread context isolation
	mu.Lock()
	defer mu.Unlock()

	is.Equal(len(metadata), 3) // Should have metadata from each iteration

	// Each iteration should have its own isolated context
	for i, meta := range metadata {
		is.Equal(meta["iteration"], i+1)
	}
}

func TestFor_NilHandler(t *testing.T) {
	is := is.New(t)

	defer func() {
		r := recover()
		is.True(r != nil)                                   // Should panic
		is.Equal(r.(string), "test: handler cannot be nil") // Should have correct panic message
	}()

	handlers.NewFor("test", 3, nil, nil)
}

func TestFor_StringRepresentation(t *testing.T) {
	is := is.New(t)

	handler := newMockHandler("handler")

	// Test finite iterations
	loop := handlers.NewFor("test", 5, handler, nil)
	is.Equal(loop.String(), "For(test, 5 iterations)")

	// Test infinite iterations
	infiniteLoop := handlers.NewFor("test", 0, handler, nil)
	is.Equal(infiniteLoop.String(), "For(test, infinite iterations)")
}
