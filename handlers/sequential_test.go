package handlers_test

import (
	"context"
	"testing"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

// type mockHandler struct {
// 	name      string
// 	shouldErr bool
// 	executed  int
// }

// func (m *mockHandler) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
// 	m.Completed()++
// 	ctx := tc.Context()
// 	if ctx.Err() != nil {
// 		return tc, ctx.Err()
// 	}
// 	if m.shouldErr {
// 		return tc, errors.New(m.name + " encountered an error")
// 	}
// 	if next != nil {
// 		return next.HandleThread(tc, nil)
// 	}
// 	return tc, nil
// }

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
	is.True(this.Started() == 1)              // "this handler should not have started"
	is.True(that.Completed() == 0)            // "that handler should not have completed"
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
	is.True(this.Completed() == 1)  // "this handler should have completed"
	is.True(that.Completed() == 1)  // "that handler should have completed"
	is.True(final.Completed() == 1) // "final handler should have completed"
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
	is.Equal(that.name+": "+errHandlerFailed.Error(), err.Error())
	is.True(this.Started() == 1)    // "this handler should have started"
	is.True(that.Started() == 1)    // "that handler should have started"
	is.True(final.Completed() == 0) // "final handler should not have completed after error"
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
	is.True(before.Completed() == 2) // "before handler should have completed"
	is.True(this.Completed() == 1)   // "this handler should have completed"
	is.True(that.Completed() == 1)   // "that handler should have completed"
	is.True(final.Completed() == 1)  // "final handler should have completed"
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
	is.True(middleware.Completed() == 2) // "after handler should have completed"
	is.True(this.Completed() == 1)       // "this handler should have completed"
	is.True(that.Completed() == 1)       // "that handler should have completed"
	is.True(final.Completed() == 1)      // "final handler should have completed"
}
