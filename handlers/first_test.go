package handlers_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

func TestFirst_FirstSuccessReturns(t *testing.T) {
	is := is.New(t)

	h1 := &mockHandler{name: "Handler1", sleep: 100 * time.Millisecond}
	h2 := &mockHandler{name: "Handler2", shouldErr: true}
	h3 := &mockHandler{name: "Handler3", sleep: 50 * time.Millisecond}

	any := handlers.First("FirstSuccess", h1, h2, h3)
	tc := minds.NewThreadContext(context.Background())

	_, err := any.HandleThread(tc, nil)

	is.NoErr(err)
	is.Equal(1, h3.Completed())        // h3 should complete
	time.Sleep(200 * time.Millisecond) // Wait for other handlers
	is.Equal(1, h1.Started())          // h1 should start
	is.Equal(1, h2.Started())          // h2 should start
}

func TestFirst_AllHandlersFail(t *testing.T) {
	is := is.New(t)

	h1 := &mockHandler{name: "Handler1", shouldErr: true}
	h2 := &mockHandler{name: "Handler2", shouldErr: true}
	h3 := &mockHandler{name: "Handler3", shouldErr: true}

	any := handlers.First("AllFail", h1, h2, h3)
	tc := minds.NewThreadContext(context.Background())

	_, err := any.HandleThread(tc, nil)

	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "Handler1: "+errHandlerFailed.Error()))
	is.True(strings.Contains(err.Error(), "Handler2: "+errHandlerFailed.Error()))
	is.True(strings.Contains(err.Error(), "Handler3: "+errHandlerFailed.Error()))
}

func TestFirst_ContextCancellation(t *testing.T) {
	is := is.New(t)

	h1 := &mockHandler{name: "Handler1", sleep: 200 * time.Millisecond}
	h2 := &mockHandler{name: "Handler2", sleep: 200 * time.Millisecond}

	any := handlers.First("ContextCancel", h1, h2)

	ctx, cancel := context.WithCancel(context.Background())
	tc := minds.NewThreadContext(ctx)

	cancel() // Cancel immediately instead of with delay

	_, err := any.HandleThread(tc, nil)
	is.True(err != nil)
}

func TestFirst_NoHandlers(t *testing.T) {
	is := is.New(t)

	any := handlers.First("Empty")
	tc := minds.NewThreadContext(context.Background())

	_, err := any.HandleThread(tc, nil)
	is.NoErr(err)
}

func TestFirst_NestedAny(t *testing.T) {
	is := is.New(t)

	h1 := &mockHandler{name: "Handler1", shouldErr: true}
	h2 := &mockHandler{name: "Handler2", sleep: 50 * time.Millisecond}
	h3 := &mockHandler{name: "Handler3", shouldErr: true}

	nestedAny := handlers.First("Nested", h2, h3)
	any := handlers.First("Outer", h1, nestedAny)

	tc := minds.NewThreadContext(context.Background())

	_, err := any.HandleThread(tc, nil)
	is.NoErr(err)
}
