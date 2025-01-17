package handlers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

// Mock handlers and middleware for testing
func newRecordingHandler(name string, values *[]string) minds.ThreadHandlerFunc {
	return func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
		*values = append(*values, name)
		return tc, nil
	}
}

func newErrorHandler(name string) minds.ThreadHandlerFunc {
	return func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
		return tc, errors.New(name + " error")
	}
}

func newErrorMiddleware(name string) handlers.Middleware {
	return handlers.MiddlewareFunc(func(next minds.ThreadHandler) minds.ThreadHandler {
		return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			return tc, errors.New(name + " error")
		})
	})
}

type recordingMiddleware struct {
	name   string
	values *[]string
}

func newRecordingMiddleware(name string, values *[]string) handlers.Middleware {
	return &recordingMiddleware{name: name, values: values}
}

func (m *recordingMiddleware) Wrap(next minds.ThreadHandler) minds.ThreadHandler {
	return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
		*m.values = append(*m.values, "before "+m.name)

		result, err := next.HandleThread(tc, nil)
		if err != nil {
			return result, err
		}

		*m.values = append(*m.values, "after "+m.name)
		return result, nil
	})
}

func TestThreadFlow_Basic(t *testing.T) {
	t.Run("single handler", func(t *testing.T) {
		is := is.New(t)

		var order []string
		flow := handlers.NewThreadFlow("test")
		flow.Handle(newRecordingHandler("handler1", &order))

		ctx := minds.NewThreadContext(context.Background())
		_, err := flow.HandleThread(ctx, nil)

		is.NoErr(err)
		is.Equal([]string{"handler1"}, order)
	})

	t.Run("multiple handlers", func(t *testing.T) {
		is := is.New(t)
		var order []string
		flow := handlers.NewThreadFlow("test")
		flow.Handle(newRecordingHandler("handler1", &order))
		flow.Handle(newRecordingHandler("handler2", &order))

		ctx := minds.NewThreadContext(context.Background())
		_, err := flow.HandleThread(ctx, nil)

		is.NoErr(err)
		is.Equal([]string{"handler1", "handler2"}, order)
	})
}

func TestThreadFlow_Middleware(t *testing.T) {
	t.Run("single middleware", func(t *testing.T) {
		is := is.New(t)

		var order []string
		flow := handlers.NewThreadFlow("test")
		flow.Use(newRecordingMiddleware("mid1", &order))
		flow.Handle(newRecordingHandler("handler", &order))

		ctx := minds.NewThreadContext(context.Background())
		_, err := flow.HandleThread(ctx, nil)

		is.NoErr(err)
		is.Equal([]string{
			"before mid1",
			"handler",
			"after mid1",
		}, order)
	})

	t.Run("multiple middleware", func(t *testing.T) {
		is := is.New(t)
		var order []string
		flow := handlers.NewThreadFlow("test")
		flow.Use(newRecordingMiddleware("mid1", &order))
		flow.Use(newRecordingMiddleware("mid2", &order))
		flow.Handle(newRecordingHandler("handler", &order))

		ctx := minds.NewThreadContext(context.Background())
		_, err := flow.HandleThread(ctx, nil)

		is.NoErr(err)
		is.Equal([]string{
			"before mid1",
			"before mid2",
			"handler",
			"after mid2",
			"after mid1",
		}, order)
	})
}

func TestThreadFlow_Groups(t *testing.T) {
	t.Run("middleware scoping", func(t *testing.T) {
		is := is.New(t)
		var order []string
		flow := handlers.NewThreadFlow("test")

		// Global middleware
		flow.Use(newRecordingMiddleware("global", &order))

		// Base handler
		flow.Handle(newRecordingHandler("base", &order))

		// Group with additional middleware
		flow.Group(func(f *handlers.ThreadFlow) {
			f.Use(newRecordingMiddleware("group1", &order))
			f.Handle(newRecordingHandler("handler1", &order))
			f.Handle(newRecordingHandler("handler2", &order))
		})

		// Another group
		flow.Group(func(f *handlers.ThreadFlow) {
			f.Use(newRecordingMiddleware("group2", &order))
			f.Handle(newRecordingHandler("handler3", &order))
		})

		ctx := minds.NewThreadContext(context.Background())
		_, err := flow.HandleThread(ctx, nil)

		is.NoErr(err)
		is.Equal([]string{
			// Base handler with global middleware
			"before global",
			"base",
			"after global",

			// First group handlers
			"before global",
			"before group1",
			"handler1",
			"after group1",
			"after global",

			"before global",
			"before group1",
			"handler2",
			"after group1",
			"after global",

			// Second group handler
			"before global",
			"before group2",
			"handler3",
			"after group2",
			"after global",
		}, order)
	})
}

func TestThreadFlow_ErrorHandling(t *testing.T) {
	t.Run("handler error", func(t *testing.T) {
		is := is.New(t)

		var order []string
		flow := handlers.NewThreadFlow("test")
		flow.Handle(newRecordingHandler("handler1", &order))
		flow.Handle(newErrorHandler("error"))
		flow.Handle(newRecordingHandler("handler2", &order))

		ctx := minds.NewThreadContext(context.Background())
		_, err := flow.HandleThread(ctx, nil)

		is.True(err != nil)
		is.Equal("error error", err.Error())
		is.Equal([]string{"handler1"}, order)
	})

	t.Run("middleware error", func(t *testing.T) {
		is := is.New(t)

		var order []string
		flow := handlers.NewThreadFlow("test")
		flow.Use(newRecordingMiddleware("mid1", &order))

		flow.Group(func(f *handlers.ThreadFlow) {
			f.Use(newErrorMiddleware("group"))
			f.Handle(newRecordingHandler("handler", &order))
		})

		ctx := minds.NewThreadContext(context.Background())
		_, err := flow.HandleThread(ctx, nil)

		is.True(err != nil)
		is.Equal("group error", err.Error())
		is.Equal([]string{"before mid1"}, order)
	})
}
