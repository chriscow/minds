package handlers_test

import (
	"context"
	"testing"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

func TestNoop(t *testing.T) {
	is := is.New(t)

	t.Run("returns unmodified context without next handler", func(t *testing.T) {
		is := is.New(t)

		noop := handlers.Noop()
		tc := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"key": "value"})

		result, err := noop.HandleThread(tc, nil)
		is.NoErr(err)
		is.Equal(result, tc) // Should return unmodified context
	})

	t.Run("calls next handler if provided", func(t *testing.T) {
		is := is.New(t)

		next := &mockHandler{name: "next"}
		noop := handlers.Noop()
		tc := minds.NewThreadContext(context.Background())

		_, err := noop.HandleThread(tc, next)
		is.NoErr(err)
		is.True(next.Called()) // Next handler should be called
	})

	t.Run("propagates next handler's response", func(t *testing.T) {
		is := is.New(t)

		expectedContext := minds.NewThreadContext(context.Background()).
			WithMetadata(minds.Metadata{"modified": true})

		next := &mockHandler{
			name:     "next",
			tcResult: expectedContext,
		}

		noop := handlers.Noop()
		tc := minds.NewThreadContext(context.Background())

		result, err := noop.HandleThread(tc, next)
		is.NoErr(err)
		is.Equal(result, expectedContext) // Should return next handler's response
	})

	t.Run("propagates next handler's error", func(t *testing.T) {
		is := is.New(t)

		next := &mockHandler{
			name:        "next",
			expectedErr: context.DeadlineExceeded,
		}

		noop := handlers.Noop()
		tc := minds.NewThreadContext(context.Background())

		_, err := noop.HandleThread(tc, next)
		is.Equal(err, next.expectedErr) // Should propagate next handler's error
	})
}
