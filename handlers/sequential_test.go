package handlers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

type mockHandler struct {
	name      string
	shouldErr bool
	executed  int
}

func (m *mockHandler) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	m.executed++
	ctx := tc.Context()
	if ctx.Err() != nil {
		return tc, ctx.Err()
	}
	if m.shouldErr {
		return tc, errors.New(m.name + " encountered an error")
	}
	if next != nil {
		return next.HandleThread(tc, nil)
	}
	return tc, nil
}

func TestSequentialHandler_ContextCanceled(t *testing.T) {
	is := is.New(t)
	// Setup mock handlers
	this := &mockHandler{name: "this"}
	that := &mockHandler{name: "that"}

	// Create Sequential handler
	seq := handlers.Sequential("test", this, that)

	// Create a context that is already canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tc := minds.NewThreadContext(ctx)

	// Execute
	_, err := seq.HandleThread(tc, nil)

	// Assert
	is.True(err != nil)                       // expect an error because the context is already canceled
	is.Equal("context canceled", err.Error()) // "expect context canceled error"
	is.True(this.executed == 1)               // "this handler should not have executed"
	is.True(that.executed == 0)               // "that handler should not have executed"
}

func TestSequentialHandler_Success(t *testing.T) {
	is := is.New(t)
	// Setup mock handlers
	this := &mockHandler{name: "this"}
	that := &mockHandler{name: "that"}
	final := &mockHandler{name: "final"}

	// Create Sequential handler
	seq := handlers.Sequential("test", this, that)
	tc := minds.NewThreadContext(context.Background())

	// Execute
	_, err := seq.HandleThread(tc, final)

	// Assert
	is.NoErr(err)
	is.True(this.executed == 1)  // "this handler should have executed"
	is.True(that.executed == 1)  // "that handler should have executed"
	is.True(final.executed == 1) // "final handler should have executed"
}

func TestSequentialHandler_WithError(t *testing.T) {
	is := is.New(t)
	// Setup mock handlers
	this := &mockHandler{name: "this"}
	that := &mockHandler{name: "that", shouldErr: true}
	final := &mockHandler{name: "final"}

	// Create Sequential handler
	seq := handlers.Sequential("test", this, that)
	tc := minds.NewThreadContext(context.Background())

	// Execute
	_, err := seq.HandleThread(tc, final)

	is.True(err != nil)
	is.Equal("that encountered an error", err.Error())
	is.True(this.executed == 1)  // "this handler should have executed"
	is.True(that.executed == 1)  // "that handler should have executed"
	is.True(final.executed == 0) // "final handler should not have executed after error"
}

func TestSequentialHandler_BeforeEach(t *testing.T) {
	is := is.New(t)

	// Setup mock handlers
	this := &mockHandler{name: "this"}
	that := &mockHandler{name: "that"}
	before := &mockHandler{name: "before"}
	final := &mockHandler{name: "final"}

	// Create Sequential handler
	seq := handlers.Sequential("test", this, that)
	seq.Use(before)

	tc := minds.NewThreadContext(context.Background())

	// Execute
	_, err := seq.HandleThread(tc, final)

	// Assert
	is.NoErr(err)
	is.True(before.executed == 2) // "before handler should have executed"
	is.True(this.executed == 1)   // "this handler should have executed"
	is.True(that.executed == 1)   // "that handler should have executed"
	is.True(final.executed == 1)  // "final handler should have executed"
}

func TestSequentialHandler_AfterEach(t *testing.T) {
	is := is.New(t)

	// Setup mock handlers
	this := &mockHandler{name: "this"}
	that := &mockHandler{name: "that"}
	middleware := &mockHandler{name: "middleware"}
	final := &mockHandler{name: "final"}

	// Create Sequential handler
	seq := handlers.Sequential("test", this, that)
	seq.Use(middleware)

	tc := minds.NewThreadContext(context.Background())

	// Execute
	_, err := seq.HandleThread(tc, final)

	// Assert
	is.NoErr(err)
	is.True(middleware.executed == 2) // "after handler should have executed"
	is.True(this.executed == 1)       // "this handler should have executed"
	is.True(that.executed == 1)       // "that handler should have executed"
	is.True(final.executed == 1)      // "final handler should have executed"
}
