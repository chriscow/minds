package handlers_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

// You can define a sentinel error:
var errMustHandlerFailed = errors.New("handler failed")

// Or, if you need more detail, define a struct that implements `error`:
type MustHandlerError struct {
	Reason string
}

func (e *MustHandlerError) Error() string {
	return e.Reason
}

type mockMustHandler struct {
	name      string
	shouldErr bool
	sleep     time.Duration
	started   int32
	completed int32
}

func (m *mockMustHandler) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	atomic.AddInt32(&m.started, 1) // Track execution count

	if m.sleep <= 0 {
		if m.shouldErr {
			// Return an error immediately if shouldErr is true and sleep is zero or negative
			return tc, fmt.Errorf("%s: %w", m.name, errMustHandlerFailed)
		}

		// Complete immediately if no sleep is required
		atomic.AddInt32(&m.completed, 1)
		return tc, nil
	}

	// Simulate work that respects context cancellation
	ticker := time.NewTicker(2 * time.Millisecond)
	defer ticker.Stop()

	sleepTime := 0

	for {
		select {
		case <-ticker.C:
			sleepTime += 2
			// fmt.Printf("%s: slept %d ms\n", m.name, sleepTime)
			if sleepTime >= int(m.sleep/time.Millisecond) {
				if m.shouldErr {
					// fmt.Printf("%s: encountered an error\n", m.name)
					return tc, fmt.Errorf("%s: %w", m.name, errMustHandlerFailed)
				}

				// fmt.Printf("%s: completed\n", m.name)
				atomic.AddInt32(&m.completed, 1)
				return tc, nil
			}
		case <-tc.Context().Done():
			// fmt.Printf("%s: context canceled\n", m.name)
			return tc, tc.Context().Err() // Exit early on cancellation
		}
	}
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
	is.NoErr(err)                  // "All handlers should succeed"
	is.Equal(int32(1), h1.started) // "Handler1 should have executed once"
	is.Equal(int32(1), h2.started) // "Handler2 should have executed once"
	is.Equal(int32(1), h3.started) // "Handler3 should have executed once"
}

func TestMust_OneHandlerFails(t *testing.T) {
	is := is.New(t)

	for i := 0; i < 1000; i++ {
		t.Run(fmt.Sprintf("OneHandlerFails#%d", i), func(t *testing.T) {
			h1 := &mockMustHandler{name: "Handler1", sleep: 1000 * time.Millisecond}
			h2 := &mockMustHandler{name: "Handler2", shouldErr: true}
			h3 := &mockMustHandler{name: "Handler3", sleep: 1000 * time.Millisecond}

			must := handlers.Must("OneFails", h1, h2, h3)
			tc := minds.NewThreadContext(context.Background())
			_, err := must.HandleThread(tc, nil)

			// Assert: Allow both "Handler2 encountered an error" and "context canceled"
			is.True(err != nil) // "An error must occur"
			if !strings.Contains(err.Error(), "Handler2: handler failed") &&
				!errors.Is(err, context.Canceled) {
				t.Fatalf("unexpected error: %v", err)
			}

			// Handler1 may or may not have started and completed ¯\_(ツ)_/¯
			if h1.started > 0 {
				is.Equal(int32(0), h1.completed) // "Handler1 should not complete"
			}

			is.Equal(int32(1), h2.started)   // "Handler2 should have started"
			is.Equal(int32(0), h2.completed) // "Handler2 should not complete successfully"

			// Check Handler3's behavior: May or may not have started
			if h3.started > 0 {
				is.Equal(int32(0), h3.completed) // "Handler3 should not complete"
			}
		})
	}
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

	for i := 0; i < 1000; i++ {
		t.Run(fmt.Sprintf("OneHandlerFails#%d", i), func(t *testing.T) {

			// Setup mock handlers
			h1 := &mockMustHandler{name: "Handler1"}
			h2 := &mockMustHandler{name: "Handler2", shouldErr: true}
			h3 := &mockMustHandler{name: "Handler3"}

			// Create nested Must handler
			nestedMust := handlers.Must("Nested", h2, h3)
			must := handlers.Must("Outer", h1, nestedMust)

			tc := minds.NewThreadContext(context.Background())

			// Execute
			_, err := must.HandleThread(tc, nil)

			// Assert
			is.True(err != nil) // "An error must occur"
			// Ensure the error message contains the nested and outer context
			is.True(strings.Contains(err.Error(), "Nested"))
			is.True(strings.Contains(err.Error(), "Handler2: handler failed"))
		})
	}
}
