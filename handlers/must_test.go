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
var errHandlerFailed = errors.New("handler failed")

// Or, if you need more detail, define a struct that implements `error`:
type HandlerError struct {
	Reason string
}

func (e *HandlerError) Error() string {
	return e.Reason
}

type mockHandler struct {
	name      string
	shouldErr bool
	sleep     time.Duration
	started   int32
	completed int32
	tcResult  minds.ThreadContext
	metadata  map[string]interface{} // Add metadata for testing merging
}

func newMockHandler(name string) *mockHandler {
	return &mockHandler{
		name:     name,
		metadata: make(map[string]interface{}),
	}
}

func (m *mockHandler) String() string {
	return m.name
}

func (m *mockHandler) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	atomic.AddInt32(&m.started, 1)

	if m.tcResult == nil {
		m.tcResult = tc.Clone()
		m.tcResult.SetKeyValue("handler", m.name)
	}

	if m.sleep > 0 {
		select {
		case <-time.After(m.sleep):
		case <-tc.Context().Done():
			return tc, tc.Context().Err()
		}
	}

	if tc.Context().Err() != nil {
		return tc, tc.Context().Err()
	}

	if m.shouldErr {
		return tc, fmt.Errorf("%s: %w", m.name, errHandlerFailed)
	}

	atomic.AddInt32(&m.completed, 1)
	return m.tcResult, nil
}

func (m *mockHandler) Started() int {
	return int(atomic.LoadInt32(&m.started))
}

func (m *mockHandler) Completed() int {
	return int(atomic.LoadInt32(&m.completed))
}

func TestMust_AllHandlersSucceed(t *testing.T) {
	is := is.New(t)
	// Setup mock handlers
	h1 := &mockHandler{name: "Handler1"}
	h2 := &mockHandler{name: "Handler2"}
	h3 := &mockHandler{name: "Handler3"}

	// Create Must handler
	must := handlers.Must("AllSucceed", nil, h1, h2, h3)

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
			h1 := &mockHandler{name: "Handler1", sleep: 1000 * time.Millisecond}
			h2 := &mockHandler{name: "Handler2", shouldErr: true}
			h3 := &mockHandler{name: "Handler3", sleep: 1000 * time.Millisecond}

			must := handlers.Must("OneFails", nil, h1, h2, h3)
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
	h1 := &mockHandler{name: "Handler1"}
	h2 := &mockHandler{name: "Handler2"}
	h3 := &mockHandler{name: "Handler3"}

	// Create Must handler
	must := handlers.Must("ContextCancel", nil, h1, h2, h3)

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
	must := handlers.Must("Empty", nil)

	tc := minds.NewThreadContext(context.Background())
	_, err := must.HandleThread(tc, nil)

	// Without any handlers, there will be no results to aggregate
	is.Equal(err.Error(), "Empty aggregation: no results to aggregate") // "Must with no handlers return no aggregators error"
}

func TestMust_NestedMust(t *testing.T) {
	is := is.New(t)

	for i := 0; i < 1000; i++ {
		t.Run(fmt.Sprintf("OneHandlerFails#%d", i), func(t *testing.T) {

			// Setup mock handlers
			h1 := &mockHandler{name: "Handler1"}
			h2 := &mockHandler{name: "Handler2", shouldErr: true}
			h3 := &mockHandler{name: "Handler3"}

			// Create nested Must handler
			nestedMust := handlers.Must("Nested", nil, h2, h3)
			must := handlers.Must("Outer", nil, h1, nestedMust)

			tc := minds.NewThreadContext(context.Background())
			_, err := must.HandleThread(tc, nil)

			is.True(err != nil) // expected error to occur

			if strings.Contains(err.Error(), "context canceled") {
				return
			}

			// Ensure the error message contains the nested and outer context
			if !strings.Contains(err.Error(), "Handler2: handler failed") || !strings.Contains(err.Error(), "Nested") {
				t.Logf("Error: %s", err.Error())
			}

			is.True(strings.Contains(err.Error(), "Nested"))
			is.True(strings.Contains(err.Error(), "Handler2: handler failed"))
		})
	}
}

func TestMust_CustomAggregator(t *testing.T) {
	is := is.New(t)

	h1 := newMockHandler("Handler1")
	h2 := newMockHandler("Handler2")
	h3 := newMockHandler("Handler3")

	var aggregatorCalled bool
	customAggregator := func(results []handlers.HandlerResult) (minds.ThreadContext, error) {
		aggregatorCalled = true

		// Custom aggregation logic - only use results from Handler1 and Handler3
		var finalCtx minds.ThreadContext
		for _, r := range results {
			s := r.Handler.(fmt.Stringer)
			isHandler1 := strings.Contains(s.String(), "Handler1")
			isHandler3 := strings.Contains(s.String(), "Handler3")
			if r.Error == nil && (isHandler1 || isHandler3) {
				if finalCtx == nil {
					finalCtx = r.Context
				} else {
					finalMeta := finalCtx.Metadata()
					finalMeta["handler"] = finalMeta["handler"].(string) + "," + r.Context.Metadata()["handler"].(string)
					finalCtx = finalCtx.WithMetadata(finalMeta)
				}
			}
		}
		return finalCtx, nil
	}

	must := handlers.Must("CustomAggregator",
		customAggregator,
		h1, h2, h3,
	)

	tc := minds.NewThreadContext(context.Background())
	finalTC, err := must.HandleThread(tc, nil)

	is.NoErr(err)
	is.True(aggregatorCalled)

	// Verify that only Handler1 and Handler3 metadata is present
	metadata := finalTC.Metadata()
	is.True(metadata["handler"] == "Handler1,Handler3" || metadata["handler"] == "Handler3,Handler1")
	is.True(metadata["handler"] != "Handler2")
	is.True(!strings.Contains(metadata["handler"].(string), "Handler2"))
}

func TestMust_DefaultAggregator(t *testing.T) {
	is := is.New(t)

	h1 := newMockHandler("Handler1")
	h1.tcResult = minds.NewThreadContext(context.Background())
	h1.tcResult.SetKeyValue("key1", "value1")

	h2 := newMockHandler("Handler2")
	h2.tcResult = minds.NewThreadContext(context.Background())
	h2.tcResult.SetKeyValue("key2", "value2")

	must := handlers.Must("DefaultAggregator",
		handlers.DefaultAggregator,
		h1, h2,
	)

	tc := minds.NewThreadContext(context.Background())
	finalTC, err := must.HandleThread(tc, nil)

	is.NoErr(err)

	// Verify metadata was merged
	metadata := finalTC.Metadata()
	is.Equal(metadata["key1"], "value1")
	is.Equal(metadata["key2"], "value2")
}

func TestMust_AggregatorWithErrors(t *testing.T) {
	is := is.New(t)

	h1 := newMockHandler("Handler1")
	h2 := newMockHandler("Handler2")
	h2.shouldErr = true
	h3 := newMockHandler("Handler3")

	var aggregatorCalled bool
	aggregator := func(results []handlers.HandlerResult) (minds.ThreadContext, error) {
		aggregatorCalled = true
		return nil, fmt.Errorf("aggregator failed")
	}

	must := handlers.Must("AggregatorError",
		aggregator,
		h1, h2, h3,
	)

	tc := minds.NewThreadContext(context.Background())
	_, err := must.HandleThread(tc, nil)

	is.True(err != nil)
	is.True(!aggregatorCalled) // Aggregator shouldn't be called if a handler fails
	is.True(strings.Contains(err.Error(), "Handler2: handler failed"))
}
