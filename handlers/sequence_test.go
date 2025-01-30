package handlers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

func TestSequence(t *testing.T) {
	t.Run("basic execution", func(t *testing.T) {
		is := is.New(t)
		var executed []string

		h1 := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			executed = append(executed, "h1")
			return tc, nil
		})
		h2 := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			executed = append(executed, "h2")
			return tc, nil
		})

		seq := handlers.NewSequence("test", h1, h2)
		tc := minds.NewThreadContext(context.Background())

		_, err := seq.HandleThread(tc, nil)
		is.NoErr(err)
		is.Equal(executed, []string{"h1", "h2"})
	})

	t.Run("middleware wraps each handler", func(t *testing.T) {
		is := is.New(t)
		var executed []string

		h1 := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			executed = append(executed, "h1")
			return tc, nil
		})
		h2 := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			executed = append(executed, "h2")
			return tc, nil
		})

		mw := minds.MiddlewareFunc(func(next minds.ThreadHandler) minds.ThreadHandler {
			return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
				executed = append(executed, "mw_start")
				result, err := next.HandleThread(tc, nil)
				if err == nil {
					executed = append(executed, "mw_end")
				}
				return result, err
			})
		})

		seq := handlers.NewSequence("test", h1, h2)
		seq.Use(mw)
		tc := minds.NewThreadContext(context.Background())

		_, err := seq.HandleThread(tc, nil)
		is.NoErr(err)

		expected := []string{
			"mw_start", // h1's middleware starts
			"h1",       // h1 executes
			"mw_end",   // h1's middleware ends
			"mw_start", // h2's middleware starts
			"h2",       // h2 executes
			"mw_end",   // h2's middleware ends
		}
		is.Equal(executed, expected)
	})

	t.Run("handler error stops sequence", func(t *testing.T) {
		is := is.New(t)
		var executed []string

		h1 := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			executed = append(executed, "h1")
			return tc, nil
		})
		h2 := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			executed = append(executed, "h2")
			return tc, errors.New("h2 failed")
		})
		h3 := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			executed = append(executed, "h3")
			return tc, nil
		})

		seq := handlers.NewSequence("test", h1, h2, h3)
		tc := minds.NewThreadContext(context.Background())

		_, err := seq.HandleThread(tc, nil)
		is.True(err != nil)
		is.Equal(err.Error(), "test: handler error: h2 failed")
		is.Equal(executed, []string{"h1", "h2"}) // h3 should not execute
	})

	t.Run("next handler executes after sequence", func(t *testing.T) {
		is := is.New(t)
		var executed []string

		h1 := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			executed = append(executed, "h1")
			return tc, nil
		})
		next := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			executed = append(executed, "next")
			return tc, nil
		})

		seq := handlers.NewSequence("test", h1)
		tc := minds.NewThreadContext(context.Background())

		_, err := seq.HandleThread(tc, next)
		is.NoErr(err)
		is.Equal(executed, []string{"h1", "next"})
	})

	t.Run("operation-level middleware via parent", func(t *testing.T) {
		is := is.New(t)
		var executed []string

		h1 := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			executed = append(executed, "h1")
			return tc, nil
		})
		h2 := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			executed = append(executed, "h2")
			return tc, nil
		})

		// Create child middleware
		childMw := minds.MiddlewareFunc(func(next minds.ThreadHandler) minds.ThreadHandler {
			return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
				executed = append(executed, "child_mw_start")
				result, err := next.HandleThread(tc, nil)
				if err == nil {
					executed = append(executed, "child_mw_end")
				}
				return result, err
			})
		})

		// Create parent middleware
		parentMw := minds.MiddlewareFunc(func(next minds.ThreadHandler) minds.ThreadHandler {
			return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
				executed = append(executed, "parent_mw_start")
				result, err := next.HandleThread(tc, nil)
				if err == nil {
					executed = append(executed, "parent_mw_end")
				}
				return result, err
			})
		})

		// Create sequence with child middleware
		seq := handlers.NewSequence("child", h1, h2)
		seq.Use(childMw)

		// Create parent with operation-level middleware
		parent := handlers.NewSequence("parent", seq)
		parent.Use(parentMw)

		tc := minds.NewThreadContext(context.Background())
		_, err := parent.HandleThread(tc, nil)
		is.NoErr(err)

		expected := []string{
			"parent_mw_start", // Parent sequence middleware starts
			"child_mw_start",  // h1's middleware starts
			"h1",              // h1 executes
			"child_mw_end",    // h1's middleware ends
			"child_mw_start",  // h2's middleware starts
			"h2",              // h2 executes
			"child_mw_end",    // h2's middleware ends
			"parent_mw_end",   // Parent sequence middleware ends
		}
		is.Equal(executed, expected)
	})
}
