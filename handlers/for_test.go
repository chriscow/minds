package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

func TestCycleHandler_ContextCanceled(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler", sleep: 100 * time.Millisecond}
	cycler := handlers.For("test", 2, handler, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tc := minds.NewThreadContext(ctx)
	_, err := cycler.HandleThread(tc, nil)

	is.True(err != nil) // Cycle should return an error when context is canceled
	is.Equal("context canceled", err.Error())
	is.Equal(1, handler.Started())
	is.Equal(0, handler.Completed())
}

func TestCycleHandler_FixedIterations(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	final := &mockHandler{name: "final"}
	iterations := 3

	cycler := handlers.For("test", iterations, handler, nil)
	tc := minds.NewThreadContext(context.Background())

	_, err := cycler.HandleThread(tc, final)

	is.NoErr(err)
	is.Equal(iterations, handler.Completed())
	is.Equal(1, final.Completed())
}

func TestCycleHandler_WithError(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler", shouldErr: true}
	final := &mockHandler{name: "final"}

	cycler := handlers.For("test", 3, handler, nil)
	tc := minds.NewThreadContext(context.Background())

	_, err := cycler.HandleThread(tc, final)

	is.True(err != nil)
	is.Equal("handler: handler failed", err.Error())
	is.Equal(1, handler.Started())
	is.Equal(0, final.Completed())
}

func TestCycleHandler_MultipleHandlers(t *testing.T) {
	is := is.New(t)
	handler1 := &mockHandler{name: "handler1"}
	handler2 := &mockHandler{name: "handler2"}
	iterations := 2

	sequence := handlers.Sequential("sequence", handler1, handler2)

	cycler := handlers.For("test", iterations, sequence, nil)
	tc := minds.NewThreadContext(context.Background())

	_, err := cycler.HandleThread(tc, nil)

	is.NoErr(err)
	is.Equal(iterations, handler1.Completed())
	is.Equal(iterations, handler2.Completed())
}

func TestCycleHandler_InfiniteStop(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	cycler := handlers.For("test", 0, handler, nil)
	tc := minds.NewThreadContext(ctx)

	go func() {
		time.Sleep(100 * time.Millisecond) // Let it run a few iterations
		cancel()
		close(done)
	}()

	_, err := cycler.HandleThread(tc, nil)
	<-done

	is.True(err != nil)
	is.Equal("context canceled", err.Error())
	is.True(handler.Completed() > 0)
}

func TestCycleHandler_WithContinueFunction(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	final := &mockHandler{name: "final"}

	continueFn := func(tc minds.ThreadContext, iter int) bool {
		return iter < 2 // Stop after 2 iterations
	}

	cycler := handlers.For("test", 5, handler, continueFn)
	tc := minds.NewThreadContext(context.Background())

	_, err := cycler.HandleThread(tc, final)

	is.NoErr(err)
	is.Equal(2, handler.Started()) // Should stop at 2 despite 5 iterations specified
	is.Equal(1, final.Completed()) // Should have completed 1 time
}

func TestCycleHandler_InfiniteWithContinueFunction(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}

	continueFn := func(tc minds.ThreadContext, i int) bool {
		return i < 3
	}

	cycler := handlers.For("test", 0, handler, continueFn)
	tc := minds.NewThreadContext(context.Background())

	_, err := cycler.HandleThread(tc, nil)

	is.NoErr(err)
	is.Equal(3, handler.Completed())
}
