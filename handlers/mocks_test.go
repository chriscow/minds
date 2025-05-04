package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
)

var (
	errHandlerFailed    = errors.New("handler failed")
	errMiddlewareFailed = errors.New("middleware failed")
)

// Or, if you need more detail, define a struct that implements `error`:
type HandlerError struct {
	Reason string
}

func (e *HandlerError) Error() string {
	return e.Reason
}

type mockHandler struct {
	name          string
	expectedErr   error
	sleep         time.Duration
	started       int32
	completed     int32
	tcResult      minds.ThreadContext
	metadata      map[string]any
	mu            sync.Mutex
	customExecute func() // Added for flexible execution behavior
}

func newMockHandler(name string) *mockHandler {
	return &mockHandler{
		name:     name,
		metadata: make(map[string]any),
	}
}

func (m *mockHandler) String() string {
	return m.name
}

func (m *mockHandler) Called() bool {
	return m.Completed() > 0
}

func (m *mockHandler) NotCalled() bool {
	return m.Started() == 0
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

	if m.expectedErr != nil {
		return tc, m.expectedErr
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

// Mock middleware that counts applications
type mockMiddleware struct {
	name        string
	expectedErr error
	applied     int
	executions  int
}

func (m *mockMiddleware) Wrap(next minds.ThreadHandler) minds.ThreadHandler {
	m.applied++
	return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
		m.executions++
		if m.expectedErr != nil {
			return tc, m.expectedErr
		}
		return next.HandleThread(tc, nil)
	})
}

// Mock middleware handler that implements MiddlewareHandler interface
type mockMiddlewareHandler struct {
	*mockHandler
	middleware []minds.Middleware
	mu         sync.Mutex
}

func newMockMiddlewareHandler(name string) *mockMiddlewareHandler {
	return &mockMiddlewareHandler{
		mockHandler: &mockHandler{
			name:     name,
			metadata: make(map[string]any),
		},
	}
}

func (m *mockMiddlewareHandler) Use(middleware ...minds.Middleware) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.middleware = append(m.middleware, middleware...)
}

func (m *mockMiddlewareHandler) With(middleware ...minds.Middleware) minds.ThreadHandler {
	m.mu.Lock()
	defer m.mu.Unlock()

	newHandler := &mockMiddlewareHandler{
		mockHandler: m.mockHandler,
		middleware:  append([]minds.Middleware{}, m.middleware...),
	}
	newHandler.Use(middleware...)
	return newHandler
}

func (m *mockMiddlewareHandler) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	// Apply middleware in reverse order
	handler := minds.ThreadHandler(m)

	m.mu.Lock()
	middleware := append([]minds.Middleware{}, m.middleware...)
	m.mu.Unlock()

	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i].Wrap(handler)
	}

	return handler.HandleThread(tc, next)
}

type mockProvider struct {
	response minds.Response
}

func (m *mockProvider) ModelName() string {
	return "mock-model"
}

func (m *mockProvider) GenerateContent(ctx context.Context, req minds.Request) (minds.Response, error) {
	return m.response, nil
}

func (m *mockProvider) Close() {
	// No-op
}

type mockResponse struct {
	Content string
}

func (m mockResponse) String() string {
	return m.Content
}

func (m mockResponse) ToolCalls() []minds.ToolCall {
	return nil
}

func newMockTextResponse(content string) minds.Response {
	return mockResponse{
		Content: content,
	}
}

func newMockBoolResponse(content bool) minds.Response {
	resp := handlers.BoolResp{Bool: content}
	data, _ := json.Marshal(resp)

	return mockResponse{
		Content: string(data),
	}
}

// mockCondition implements SwitchCondition for testing error cases
type mockCondition struct {
	result bool
	err    error
}

func (m *mockCondition) Evaluate(tc minds.ThreadContext) (bool, error) {
	return m.result, m.err
}
