package handlers_test

import (
	"context"
	"testing"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

func TestRangeHandler_ContextCanceled(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	values := []interface{}{"a", "b", "c"}

	ranger := handlers.Range("test", handler, values...)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tc := minds.NewThreadContext(ctx)
	_, err := ranger.HandleThread(tc, nil)

	is.True(err != nil)
	is.Equal("context canceled", err.Error())
	is.Equal(1, handler.executed)
}

func TestRangeHandler_Success(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	final := &mockHandler{name: "final"}
	values := []interface{}{"a", "b", "c"}

	ranger := handlers.Range("test", handler, values...)
	tc := minds.NewThreadContext(context.Background())

	_, err := ranger.HandleThread(tc, final)

	is.NoErr(err)
	is.Equal(3, handler.executed) // Handler should execute for each value
	is.Equal(1, final.executed)   // Final handler executes once
}

func TestRangeHandler_WithError(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler", shouldErr: true}
	final := &mockHandler{name: "final"}
	values := []interface{}{"a", "b", "c"}

	ranger := handlers.Range("test", handler, values...)
	tc := minds.NewThreadContext(context.Background())

	_, err := ranger.HandleThread(tc, final)

	is.True(err != nil)
	is.Equal("handler encountered an error", err.Error())
	is.Equal(1, handler.executed) // Should stop at first error
	is.Equal(0, final.executed)   // Final handler shouldn't execute after error
}

func TestRangeHandler_WithMiddleware(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	middleware := &mockHandler{name: "middleware"}
	final := &mockHandler{name: "final"}
	values := []interface{}{"a", "b", "c"}

	ranger := handlers.Range("test", handler, values...)
	ranger.Use(middleware)

	tc := minds.NewThreadContext(context.Background())
	_, err := ranger.HandleThread(tc, final)

	is.NoErr(err)
	is.Equal(3, handler.executed)    // Handler executes for each value
	is.Equal(3, middleware.executed) // Middleware executes for each value
	is.Equal(1, final.executed)      // Final handler executes once
}

func TestRangeHandler_ValueInContext(t *testing.T) {
	is := is.New(t)
	values := []interface{}{"a", "b", "c"}
	var seenValues []interface{}

	handler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		seenValues = append(seenValues, tc.Metadata()["range_value"])
		return tc, nil
	})

	ranger := handlers.Range("test", handler, values...)
	tc := minds.NewThreadContext(context.Background())

	_, err := ranger.HandleThread(tc, nil)

	is.NoErr(err)
	is.Equal(len(values), len(seenValues))
	for i, v := range values {
		is.Equal(v, seenValues[i])
	}
}
