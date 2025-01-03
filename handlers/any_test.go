package handlers_test

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

type mockAnyHandler struct {
	name      string
	shouldErr bool
	delay     time.Duration
	executed  int32
}

func (m *mockAnyHandler) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	atomic.AddInt32(&m.executed, 1)
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.shouldErr {
		return tc, errors.New(m.name + " encountered an error")
	}
	return tc, nil
}

func TestAny_FirstSuccessReturns(t *testing.T) {
	is := is.New(t)

	h1 := &mockAnyHandler{name: "Handler1", delay: 100 * time.Millisecond}
	h2 := &mockAnyHandler{name: "Handler2", shouldErr: true}
	h3 := &mockAnyHandler{name: "Handler3", delay: 50 * time.Millisecond}

	any := handlers.Any("FirstSuccess", h1, h2, h3)
	tc := minds.NewThreadContext(context.Background())

	_, err := any.HandleThread(tc, nil)

	is.NoErr(err)
	is.Equal(int32(1), h3.executed)
	time.Sleep(200 * time.Millisecond) // Wait for other handlers
	is.Equal(int32(1), h1.executed)
	is.Equal(int32(1), h2.executed)
}

func TestAny_AllHandlersFail(t *testing.T) {
	is := is.New(t)

	h1 := &mockAnyHandler{name: "Handler1", shouldErr: true}
	h2 := &mockAnyHandler{name: "Handler2", shouldErr: true}
	h3 := &mockAnyHandler{name: "Handler3", shouldErr: true}

	any := handlers.Any("AllFail", h1, h2, h3)
	tc := minds.NewThreadContext(context.Background())

	_, err := any.HandleThread(tc, nil)

	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "Handler1 encountered an error"))
	is.True(strings.Contains(err.Error(), "Handler2 encountered an error"))
	is.True(strings.Contains(err.Error(), "Handler3 encountered an error"))
}

func TestAny_ContextCancellation(t *testing.T) {
	is := is.New(t)

	h1 := &mockAnyHandler{name: "Handler1", delay: 200 * time.Millisecond}
	h2 := &mockAnyHandler{name: "Handler2", delay: 200 * time.Millisecond}

	any := handlers.Any("ContextCancel", h1, h2)

	ctx, cancel := context.WithCancel(context.Background())
	tc := minds.NewThreadContext(ctx)

	cancel() // Cancel immediately instead of with delay

	_, err := any.HandleThread(tc, nil)
	is.True(err != nil)
}

func TestAny_NoHandlers(t *testing.T) {
	is := is.New(t)

	any := handlers.Any("Empty")
	tc := minds.NewThreadContext(context.Background())

	_, err := any.HandleThread(tc, nil)
	is.NoErr(err)
}

func TestAny_NestedAny(t *testing.T) {
	is := is.New(t)

	h1 := &mockAnyHandler{name: "Handler1", shouldErr: true}
	h2 := &mockAnyHandler{name: "Handler2", delay: 50 * time.Millisecond}
	h3 := &mockAnyHandler{name: "Handler3", shouldErr: true}

	nestedAny := handlers.Any("Nested", h2, h3)
	any := handlers.Any("Outer", h1, nestedAny)

	tc := minds.NewThreadContext(context.Background())

	_, err := any.HandleThread(tc, nil)
	is.NoErr(err)
}
