package handlers_test

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

type mockMustHandler struct {
	name      string
	shouldErr bool
	executed  int32
}

func (m *mockMustHandler) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	atomic.AddInt32(&m.executed, 1) // Track execution count
	if m.shouldErr {
		return tc, errors.New(m.name + " encountered an error")
	}
	return tc, nil
}

func TestMust_AllHandlersSucceed(t *testing.T) {
	is := is.New(t)
	// Setup mock handlers
	h1 := &mockMustHandler{name: "Handler1"}
	h2 := &mockMustHandler{name: "Handler2"}
	h3 := &mockMustHandler{name: "Handler3"}

	// Create Must handler
	must := handlers.Must("AllSucceed", h1, h2, h3)

	tc := minds.NewThreadContext(context.Background())

	// Execute
	_, err := must.HandleThread(tc, minds.NoopThreadHandler{})

	// Assert
	is.NoErr(err)                   // "All handlers should succeed"
	is.Equal(int32(1), h1.executed) // "Handler1 should have executed once"
	is.Equal(int32(1), h2.executed) // "Handler2 should have executed once"
	is.Equal(int32(1), h3.executed) // "Handler3 should have executed once"
}

func TestMust_OneHandlerFails(t *testing.T) {
	is := is.New(t)

	// Setup mock handlers
	h1 := &mockMustHandler{name: "Handler1"}
	h2 := &mockMustHandler{name: "Handler2", shouldErr: true}
	h3 := &mockMustHandler{name: "Handler3"}

	// Create Must handler
	must := handlers.Must("OneFails", h1, h2, h3)
	tc := minds.NewThreadContext(context.Background())

	// Execute
	_, err := must.HandleThread(tc, nil)

	// Assert
	is.Equal(err.Error(), "Handler2 encountered an error")
	is.True(strings.HasPrefix(err.Error(), "Handler2 encountered an error"))
	is.Equal(int32(1), h1.executed) // "Handler1 should have executed once"
	is.Equal(int32(1), h2.executed) // "Handler2 should have executed once"
	is.Equal(int32(0), h3.executed) // "Handler3 should not have executed"
}

func TestMust_ContextCancellation(t *testing.T) {
	is := is.New(t)

	// Setup mock handlers
	h1 := &mockMustHandler{name: "Handler1"}
	h2 := &mockMustHandler{name: "Handler2"}
	h3 := &mockMustHandler{name: "Handler3"}

	// Create Must handler
	must := handlers.Must("ContextCancel", h1, h2, h3)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tc := minds.NewThreadContext(ctx)

	// Execute
	_, err := must.HandleThread(tc, nil)

	// Assert
	is.True(err != nil) // "Must should return an error when context is canceled"
	is.Equal(err.Error(), "context canceled")
}

func TestMust_NoHandlers(t *testing.T) {
	is := is.New(t)

	// Create Must handler with no handlers
	must := handlers.Must("Empty", []minds.ThreadHandler{}...)

	tc := minds.NewThreadContext(context.Background())

	// Execute
	_, err := must.HandleThread(tc, nil)

	// Assert
	is.NoErr(err) // "Must with no handlers should succeed"
}

func TestMust_NestedMust(t *testing.T) {
	is := is.New(t)

	// Setup mock handlers
	h1 := &mockMustHandler{name: "Handler1"}
	h2 := &mockMustHandler{name: "Handler2", shouldErr: true}
	h3 := &mockMustHandler{name: "Handler3"}

	nestedMust := handlers.Must("Nested", h2, h3)
	must := handlers.Must("Outer", h1, nestedMust)

	tc := minds.NewThreadContext(context.Background())

	// Execute
	_, err := must.HandleThread(tc, nil)

	// Assert
	is.Equal(err.Error(), "Handler2 encountered an error")
	is.True(strings.HasPrefix(err.Error(), "Handler2 encountered an error"))
}
