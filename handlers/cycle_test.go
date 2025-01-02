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
	handler := &mockHandler{name: "handler"}
	cycler := handlers.Cycle("test", 2, handler)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tc := minds.NewThreadContext(ctx)
	_, err := cycler.HandleThread(tc, nil)

	is.True(err != nil)
	is.Equal("context canceled", err.Error())
	is.Equal(1, handler.executed)
}

func TestCycleHandler_FixedIterations(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	final := &mockHandler{name: "final"}
	iterations := 3

	cycler := handlers.Cycle("test", iterations, handler)
	tc := minds.NewThreadContext(context.Background())

	_, err := cycler.HandleThread(tc, final)

	is.NoErr(err)
	is.Equal(iterations, handler.executed)
	is.Equal(1, final.executed)
}

func TestCycleHandler_WithError(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler", shouldErr: true}
	final := &mockHandler{name: "final"}

	cycler := handlers.Cycle("test", 3, handler)
	tc := minds.NewThreadContext(context.Background())

	_, err := cycler.HandleThread(tc, final)

	is.True(err != nil)
	is.Equal("handler encountered an error", err.Error())
	is.Equal(1, handler.executed)
	is.Equal(0, final.executed)
}

func TestCycleHandler_WithMiddleware(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	middleware := &mockHandler{name: "middleware"}
	final := &mockHandler{name: "final"}
	iterations := 2

	cycler := handlers.Cycle("test", iterations, handler)
	cycler.Use(middleware)

	tc := minds.NewThreadContext(context.Background())
	_, err := cycler.HandleThread(tc, final)

	is.NoErr(err)
	is.Equal(iterations, handler.executed)
	is.Equal(iterations, middleware.executed)
	is.Equal(1, final.executed)
}

func TestCycleHandler_MultipleHandlers(t *testing.T) {
	is := is.New(t)
	handler1 := &mockHandler{name: "handler1"}
	handler2 := &mockHandler{name: "handler2"}
	iterations := 2

	cycler := handlers.Cycle("test", iterations, handler1, handler2)
	tc := minds.NewThreadContext(context.Background())

	_, err := cycler.HandleThread(tc, nil)

	is.NoErr(err)
	is.Equal(iterations, handler1.executed)
	is.Equal(iterations, handler2.executed)
}

func TestCycleHandler_InfiniteStop(t *testing.T) {
	is := is.New(t)
	handler := &mockHandler{name: "handler"}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	cycler := handlers.Cycle("test", 0, handler)
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
	is.True(handler.executed > 0)
}
