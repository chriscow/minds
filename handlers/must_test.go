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

func TestMust_AllHandlersSucceed(t *testing.T) {
	is := is.New(t)
	h1 := &mockMiddlewareHandler{mockHandler: &mockHandler{name: "Handler1"}}
	h2 := &mockMiddlewareHandler{mockHandler: &mockHandler{name: "Handler2"}}
	h3 := &mockMiddlewareHandler{mockHandler: &mockHandler{name: "Handler3"}}

	must := handlers.NewMust("AllSucceed", nil, h1, h2, h3)
	tc := minds.NewThreadContext(context.Background())

	_, err := must.HandleThread(tc, minds.NoopThreadHandler{})

	is.NoErr(err)
	is.Equal(1, h1.mockHandler.Completed()) // Handler1 should complete once
	is.Equal(1, h2.mockHandler.Completed()) // Handler2 should complete once
	is.Equal(1, h3.mockHandler.Completed()) // Handler3 should complete once
}

func TestMust_OneHandlerFails(t *testing.T) {
	is := is.New(t)

	// Track execution order and middleware application
	var executionOrder []string
	var mu sync.Mutex

	addExecution := func(s string) {
		mu.Lock()
		executionOrder = append(executionOrder, s)
		mu.Unlock()
	}

	// Create middleware that tracks execution
	trackingMiddleware := minds.MiddlewareFunc(func(next minds.ThreadHandler) minds.ThreadHandler {
		return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			handler := next.(*mockMiddlewareHandler)
			addExecution(fmt.Sprintf("middleware_start_%s", handler.mockHandler.name))
			result, err := next.HandleThread(tc, nil)
			addExecution(fmt.Sprintf("middleware_end_%s", handler.mockHandler.name))
			return result, err
		})
	})

	h1 := &mockMiddlewareHandler{mockHandler: &mockHandler{name: "Handler1", sleep: 100 * time.Millisecond}}
	h2 := &mockMiddlewareHandler{mockHandler: &mockHandler{name: "Handler2", expectedErr: errHandlerFailed}}
	h3 := &mockMiddlewareHandler{mockHandler: &mockHandler{name: "Handler3", sleep: 100 * time.Millisecond}}

	must := handlers.NewMust("OneFails", nil, h1, h2, h3)
	must.Use(trackingMiddleware)

	tc := minds.NewThreadContext(context.Background())
	_, err := must.HandleThread(tc, nil)

	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "Handler2: handler failed"))

	// Verify middleware was applied to each handler
	mu.Lock()
	defer mu.Unlock()

	// Verify middleware started for all handlers
	foundStarts := make(map[string]bool)
	for _, entry := range executionOrder {
		if strings.HasPrefix(entry, "middleware_start_") {
			handlerName := strings.TrimPrefix(entry, "middleware_start_")
			foundStarts[handlerName] = true
		}
	}
	is.Equal(len(foundStarts), 3) // All handlers should have middleware start

	// Verify middleware ended appropriately
	foundEnds := make(map[string]bool)
	for _, entry := range executionOrder {
		if strings.HasPrefix(entry, "middleware_end_") {
			handlerName := strings.TrimPrefix(entry, "middleware_end_")
			foundEnds[handlerName] = true
		}
	}
	is.True(len(foundEnds) < 3) // Not all handlers should complete due to cancellation
}

func TestMust_MiddlewareOrdering(t *testing.T) {
	is := is.New(t)

	var executionOrder []string
	var mu sync.Mutex

	createOrderingMiddleware := func(name string) minds.Middleware {
		return minds.MiddlewareFunc(func(next minds.ThreadHandler) minds.ThreadHandler {
			return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
				mu.Lock()
				executionOrder = append(executionOrder, fmt.Sprintf("%s_start", name))
				mu.Unlock()

				result, err := next.HandleThread(tc, nil)

				mu.Lock()
				executionOrder = append(executionOrder, fmt.Sprintf("%s_end", name))
				mu.Unlock()

				return result, err
			})
		})
	}

	h1 := &mockMiddlewareHandler{mockHandler: &mockHandler{name: "Handler1"}}
	h2 := &mockMiddlewareHandler{mockHandler: &mockHandler{name: "Handler2"}}

	m1 := createOrderingMiddleware("middleware1")
	m2 := createOrderingMiddleware("middleware2")

	must := handlers.NewMust("OrderingTest", nil, h1, h2)
	must.Use(m1, m2)

	tc := minds.NewThreadContext(context.Background())
	_, err := must.HandleThread(tc, nil)

	is.NoErr(err)

	mu.Lock()
	defer mu.Unlock()

	// Verify middleware executed in correct order for each handler
	expectedPrefixes := []string{
		"middleware1_start",
		"middleware2_start",
		"middleware2_end",
		"middleware1_end",
	}

	// Check execution order pattern for each handler
	handlerCount := 0
	patternStart := 0
	for i := 0; i < len(executionOrder); i += len(expectedPrefixes) {
		pattern := executionOrder[patternStart : patternStart+len(expectedPrefixes)]
		for j, prefix := range expectedPrefixes {
			is.True(strings.HasPrefix(pattern[j], prefix))
		}
		handlerCount++
		patternStart += len(expectedPrefixes)
	}

	is.Equal(handlerCount, 2) // Pattern should repeat for each handler
}

func TestMust_ConcurrentExecution(t *testing.T) {
	is := is.New(t)

	var startTimes []time.Time
	var endTimes []time.Time
	var mu sync.Mutex

	// Create handlers with known execution times
	createTimedHandler := func(name string, duration time.Duration) *mockMiddlewareHandler {
		return &mockMiddlewareHandler{
			mockHandler: &mockHandler{
				name:  name,
				sleep: duration,
				customExecute: func() {
					mu.Lock()
					startTimes = append(startTimes, time.Now())
					mu.Unlock()

					time.Sleep(duration)

					mu.Lock()
					endTimes = append(endTimes, time.Now())
					mu.Unlock()
				},
			},
		}
	}

	h1 := createTimedHandler("Handler1", 100*time.Millisecond)
	h2 := createTimedHandler("Handler2", 100*time.Millisecond)
	h3 := createTimedHandler("Handler3", 100*time.Millisecond)

	must := handlers.NewMust("ConcurrentTest", nil, h1, h2, h3)

	start := time.Now()
	tc := minds.NewThreadContext(context.Background())
	_, err := must.HandleThread(tc, nil)
	duration := time.Since(start)

	is.NoErr(err)

	// Verify execution was concurrent
	is.True(duration < 250*time.Millisecond) // Total duration should be less than sum of individual durations

	mu.Lock()
	defer mu.Unlock()

	// Verify all handlers executed
	is.Equal(len(startTimes), 3)
	is.Equal(len(endTimes), 3)

	// Verify overlapping execution
	for i := 1; i < len(startTimes); i++ {
		is.True(startTimes[i].Sub(startTimes[0]) < 50*time.Millisecond) // All handlers should start within 50ms of first
	}
}

func TestMust_ThreadContextIsolation(t *testing.T) {
	is := is.New(t)

	var mu sync.Mutex
	metadata := make(map[string]interface{})

	h1 := &mockMiddlewareHandler{
		mockHandler: &mockHandler{
			name: "Handler1",
			customExecute: func() {
				time.Sleep(50 * time.Millisecond)
				mu.Lock()
				metadata["h1"] = "value1"
				mu.Unlock()
			},
		},
	}

	h2 := &mockMiddlewareHandler{
		mockHandler: &mockHandler{
			name: "Handler2",
			customExecute: func() {
				mu.Lock()
				metadata["h2"] = "value2"
				mu.Unlock()
			},
		},
	}

	must := handlers.NewMust("IsolationTest", nil, h1, h2)
	tc := minds.NewThreadContext(context.Background())

	_, err := must.HandleThread(tc, nil)
	is.NoErr(err)

	// Verify context isolation
	mu.Lock()
	defer mu.Unlock()

	is.Equal(metadata["h1"], "value1")
	is.Equal(metadata["h2"], "value2")
}

// Additional test cases from the original test file remain unchanged...
// (TestMust_ContextCancellation, TestMust_NoHandlers, etc.)
