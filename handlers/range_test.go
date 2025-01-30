// range_test.go
package handlers_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

func TestRange_ContextCanceled(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	values := []interface{}{"a", "b", "c"}

	ranger := handlers.NewRange("test", handler, values...)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tc := minds.NewThreadContext(ctx)
	_, err := ranger.HandleThread(tc, nil)

	is.True(err != nil)
	is.Equal(context.Canceled.Error(), err.Error())
	// is.Equal(1, handler.Started()) // it may or may not have started.
	is.Equal(0, handler.Completed())
}

func TestRange_Success(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	final := &mockHandler{name: "final"}
	values := []interface{}{"a", "b", "c"}

	ranger := handlers.NewRange("test", handler, values...)
	tc := minds.NewThreadContext(context.Background())

	_, err := ranger.HandleThread(tc, final)

	is.NoErr(err)
	is.Equal(3, handler.Completed()) // Handler should execute for each value
	is.Equal(1, final.Completed())   // Final handler executes once
}

func TestRange_WithMiddleware(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	middleware := &mockMiddleware{name: "test-middleware"}
	values := []interface{}{"a", "b", "c"}

	ranger := handlers.NewRange("test", handler, values...)
	ranger.Use(middleware)

	tc := minds.NewThreadContext(context.Background())
	_, err := ranger.HandleThread(tc, nil)

	is.NoErr(err)
	is.Equal(3, handler.Completed())   // Handler executes for each value
	is.Equal(3, middleware.applied)    // Middleware is applied 4 times: once for handler, once for each value
	is.Equal(3, middleware.executions) // Middleware executes for each value plus handler
}

func TestRange_WithError(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler", expectedErr: errHandlerFailed}
	final := &mockHandler{name: "final"}
	values := []interface{}{"a", "b", "c"}

	ranger := handlers.NewRange("test", handler, values...)
	tc := minds.NewThreadContext(context.Background())

	_, err := ranger.HandleThread(tc, final)

	is.True(err != nil)
	is.True(errors.Is(err, errHandlerFailed))
	is.Equal(1, handler.Started())   // Handler should start
	is.Equal(0, handler.Completed()) // Handler should not complete
	is.Equal(0, final.Completed())   // Final handler shouldn't execute after error
}

func TestRange_ValueInContext(t *testing.T) {
	is := is.New(t)
	values := []interface{}{"a", "b", "c"}
	var seenValues []interface{}

	handler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		seenValues = append(seenValues, tc.Metadata()["range_value"])
		return tc, nil
	})

	ranger := handlers.NewRange("test", handler, values...)
	tc := minds.NewThreadContext(context.Background())

	_, err := ranger.HandleThread(tc, nil)

	is.NoErr(err)
	is.Equal(len(values), len(seenValues))
	for i, v := range values {
		is.Equal(v, seenValues[i])
	}
}

func TestRange_MiddlewarePropagation(t *testing.T) {
	is := is.New(t)

	// Create handler and track execution order
	executionOrder := make([]string, 0)
	handler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		value := tc.Metadata()["range_value"]
		executionOrder = append(executionOrder, fmt.Sprintf("handler_%v", value))
		return tc, nil
	})

	// Create middleware that tracks execution per iteration
	mw := minds.MiddlewareFunc(func(next minds.ThreadHandler) minds.ThreadHandler {
		return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			value := tc.Metadata()["range_value"]
			executionOrder = append(executionOrder, fmt.Sprintf("middleware_start_%v", value))
			result, err := next.HandleThread(tc, nil)
			executionOrder = append(executionOrder, fmt.Sprintf("middleware_end_%v", value))
			return result, err
		})
	})

	// Create and configure range handler
	values := []any{"a", "b", "c"}
	ranger := handlers.NewRange("test", handler, values...)
	ranger.Use(mw)

	// Execute range
	tc := minds.NewThreadContext(context.Background())
	_, err := ranger.HandleThread(tc, nil)
	is.NoErr(err)

	// Verify execution order shows middleware wrapping both the Range handler itself
	// and each individual iteration
	expected := []string{
		"middleware_start_a", // Inner middleware starts (wrapping first iteration)
		"handler_a",          // First iteration executes
		"middleware_end_a",   // Inner middleware ends
		"middleware_start_b", // Inner middleware starts (wrapping second iteration)
		"handler_b",          // Second iteration executes
		"middleware_end_b",   // Inner middleware ends
		"middleware_start_c", // Inner middleware starts (wrapping third iteration)
		"handler_c",          // Third iteration executes
		"middleware_end_c",   // Inner middleware ends
	}
	is.Equal(executionOrder, expected)
}

func TestRange_NilHandler(t *testing.T) {
	is := is.New(t)

	defer func() {
		r := recover()
		is.True(r != nil)                                   // Should panic
		is.Equal(r.(string), "test: handler cannot be nil") // Should have correct panic message
	}()

	handlers.NewRange("test", nil, "a", "b", "c")
}

func TestRange_EmptyValues(t *testing.T) {
	is := is.New(t)

	handler := newMockHandler("handler")
	ranger := handlers.NewRange("test", handler)

	tc := minds.NewThreadContext(context.Background())
	_, err := ranger.HandleThread(tc, nil)
	is.NoErr(err)
	is.Equal(handler.Started(), 0)   // Handler should not execute
	is.Equal(handler.Completed(), 0) // Handler should not complete
}
