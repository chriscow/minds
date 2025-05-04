package middleware_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/middleware"
	"github.com/chriscow/minds/middleware/retry"
	"github.com/matryer/is"
)

func TestRetryMiddleware(t *testing.T) {
	is := is.New(t)

	t.Run("succeeds on first attempt", func(t *testing.T) {
		is := is.New(t)
		callCount := 0

		mockHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			callCount++
			return tc, nil
		})

		retry := middleware.Retry("retry_test", retry.WithAttempts(3))
		handler := retry.Wrap(mockHandler)

		ctx := minds.NewThreadContext(context.Background())
		_, err := handler.HandleThread(ctx, nil)

		is.NoErr(err)          // Expect no error
		is.Equal(callCount, 1) // Should only be called once
	})

	t.Run("retries up to max attempts and fails", func(t *testing.T) {
		is := is.New(t)
		callCount := 0
		expectedError := errors.New("failure")

		mockHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			callCount++
			return tc, expectedError
		})

		retry := middleware.Retry("retry_test", retry.WithAttempts(3))
		handler := retry.Wrap(mockHandler)

		ctx := minds.NewThreadContext(context.Background())
		_, err := handler.HandleThread(ctx, nil)

		is.True(err != nil) // Expect an error
		is.Equal(err.Error(), "retry_test: all 3 attempts failed, last error: failure")
		is.Equal(callCount, 3) // Should be called exactly 3 times
	})

	t.Run("succeeds after retries", func(t *testing.T) {
		is := is.New(t)
		callCount := 0

		mockHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			callCount++
			if callCount < 3 {
				return tc, errors.New("temporary error")
			}
			return tc, nil
		})

		retry := middleware.Retry("retry_test", retry.WithAttempts(5))
		handler := retry.Wrap(mockHandler)

		ctx := minds.NewThreadContext(context.Background())
		_, err := handler.HandleThread(ctx, nil)

		is.NoErr(err)          // Expect no error
		is.Equal(callCount, 3) // Should succeed on the 3rd attempt
	})

	t.Run("handles zero attempts gracefully", func(t *testing.T) {
		is := is.New(t)
		callCount := 0

		mockHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			callCount++
			return tc, errors.New("failure")
		})

		retry := middleware.Retry("retry_test", retry.WithAttempts(0))
		handler := retry.Wrap(mockHandler)

		ctx := minds.NewThreadContext(context.Background())
		result, err := handler.HandleThread(ctx, nil)

		is.Equal(err.Error(), "failure") // No error expected
		is.Equal(callCount, 1)           // Handler should be called once but no retries
		is.Equal(result, ctx)            // Context should remain unchanged
	})

	t.Run("context remains unchanged", func(t *testing.T) {
		is := is.New(t)
		mockHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			return tc, nil
		})

		retry := middleware.Retry("retry_test", retry.WithAttempts(3))
		handler := retry.Wrap(mockHandler)

		originalCtx := minds.NewThreadContext(context.Background())
		resultCtx, err := handler.HandleThread(originalCtx, nil)

		is.NoErr(err)                    // Expect no error
		is.Equal(resultCtx, originalCtx) // Context should remain the same
	})
}

func TestRetryOptions(t *testing.T) {
	t.Run("custom retry attempts", func(t *testing.T) {
		is := is.New(t)
		callCount := 0
		originalError := errors.New("always fails")

		mockHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			callCount++
			return tc, originalError
		})

		retry := middleware.Retry("retry_attempts", retry.WithAttempts(5))
		handler := retry.Wrap(mockHandler)

		ctx := minds.NewThreadContext(context.Background())
		_, err := handler.HandleThread(ctx, nil)

		is.True(err != nil)                    // Expect an error
		is.Equal(callCount, 5)                 // Should retry exactly 5 times
		is.True(errors.Is(err, originalError)) // Check if the error contains the original error
	})

	t.Run("retry with backoff strategy", func(t *testing.T) {
		is := is.New(t)
		callCount := 0
		startTime := time.Now()

		mockHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			callCount++
			return tc, errors.New("always fails")
		})

		backoff := func(attempt int) time.Duration {
			return time.Duration(attempt) * 10 * time.Millisecond
		}

		retry := middleware.Retry("retry_backoff", retry.WithAttempts(3), retry.WithBackoff(backoff))
		handler := retry.Wrap(mockHandler)

		ctx := minds.NewThreadContext(context.Background())
		_, err := handler.HandleThread(ctx, nil)

		is.True(err != nil)                                   // Expect an error
		is.Equal(callCount, 3)                                // Should retry exactly 3 times
		is.True(time.Since(startTime) >= 30*time.Millisecond) // Total backoff duration
	})

	t.Run("retry with context timeout propagation", func(t *testing.T) {
		is := is.New(t)
		callCount := 0

		mockHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			callCount++
			time.Sleep(10 * time.Millisecond)
			return tc, errors.New("always fails")
		})

		ctx, _ := context.WithTimeout(context.Background(), 25*time.Millisecond)
		// defer cancel()

		retry := middleware.Retry("retry_timeout", retry.WithAttempts(150))
		handler := retry.Wrap(mockHandler)

		_, err := handler.HandleThread(minds.NewThreadContext(ctx), nil)

		is.True(err != nil)                               // Expect an error due to context timeout
		is.True(errors.Is(err, context.DeadlineExceeded)) // Error should indicate timeout
		is.True(callCount <= 3)                           // Call count should stop once timeout is exceeded
	})

	t.Run("retry without timeout propagation", func(t *testing.T) {
		is := is.New(t)
		callCount := 0

		mockHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			callCount++
			time.Sleep(10 * time.Millisecond)
			return tc, errors.New("always fails")
		})

		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
		defer cancel()

		retry := middleware.Retry("retry_no_timeout", retry.WithAttempts(5), retry.WithoutTimeoutPropagation())
		handler := retry.Wrap(mockHandler)

		_, err := handler.HandleThread(minds.NewThreadContext(ctx), nil)

		is.True(err != nil)    // Expect an error
		is.Equal(callCount, 5) // Should attempt all retries despite timeout
	})
}

func TestRetryMiddleware_Composition(t *testing.T) {
	t.Run("middleware chaining", func(t *testing.T) {
		is := is.New(t)

		// Create multiple middleware layers
		loggingMiddleware := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			tc.SetKeyValue("logged", true)
			return tc, nil
		})

		callCount := 0
		mockHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			callCount++
			logged, _ := tc.Metadata()["logged"].(bool)
			is.True(logged)

			if callCount < 3 {
				return tc, errors.New("temporary error")
			}
			return tc, nil
		})

		// Chain middleware
		retry := middleware.Retry("retry_test", retry.WithAttempts(3))

		// Wrap handler with multiple middleware
		handler := loggingMiddleware.Wrap(retry.Wrap(mockHandler))

		ctx := minds.NewThreadContext(context.Background())
		_, err := handler.HandleThread(ctx, nil)

		is.NoErr(err)
		is.Equal(callCount, 3)
	})
}

// Custom error types
type TemporaryError struct {
	msg string
}
type PermanentError struct {
	msg string
}

func (e TemporaryError) Error() string { return e.msg }

func (e PermanentError) Error() string   { return e.msg }
func (e TemporaryError) Temporary() bool { return true }

func TestRetryMiddleware_InterfaceCompliance(t *testing.T) {
	is := is.New(t)

	// Verify Retry implements Middleware interface
	retry := middleware.Retry("test")

	is.True(retry != nil)

	// Check Wrap method
	mockHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
		return tc, nil
	})

	wrappedHandler := retry.Wrap(mockHandler)
	is.True(wrappedHandler != nil)
}

func TestRetryMiddleware_ContextAwareCriteria(t *testing.T) {
	is := is.New(t)
	callCount := 0

	// Mock handler that always fails
	mockHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
		callCount++
		return tc, errors.New("always fails")
	})

	// Custom retry criteria: Stop retrying if attempt >= 2
	customCriteria := func(tc minds.ThreadContext, attempt int, err error) bool {
		return attempt < 2 // Retry only twice
	}

	retry := middleware.Retry("retry_test", retry.WithAttempts(5), retry.WithRetryCriteria(customCriteria))
	handler := retry.Wrap(mockHandler)

	ctx := minds.NewThreadContext(context.Background())
	_, err := handler.HandleThread(ctx, nil)

	is.True(err != nil)                                    // Expect an error
	is.Equal(callCount, 3)                                 // Should retry exactly 2 times (1 initial + 2 retries)
	is.True(strings.Contains(err.Error(), "always fails")) // Check original error
}

func TestRetryMiddleware_MetadataAwareCriteria(t *testing.T) {
	is := is.New(t)
	callCount := 0

	// Mock handler that fails if metadata["force_fail"] == true
	mockHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
		callCount++
		if meta, ok := tc.Metadata()["force_fail"].(bool); ok && meta {
			return tc, errors.New("metadata failure")
		}
		return tc, nil
	})

	// Custom retry criteria: Stop if "no_retries" metadata is true
	customCriteria := func(tc minds.ThreadContext, attempt int, err error) bool {
		if meta, ok := tc.Metadata()["no_retries"].(bool); ok && meta {
			return false // Don't retry
		}
		return true
	}

	// Create context with metadata
	tc := minds.NewThreadContext(context.Background()).WithMetadata(map[string]any{
		"force_fail": true, // Force failure
		"no_retries": true, // Disable retries
	})

	retry := middleware.Retry("metadata-retry", retry.WithAttempts(5), retry.WithRetryCriteria(customCriteria))
	handler := retry.Wrap(mockHandler)

	_, err := handler.HandleThread(tc, nil)

	is.True(err != nil)    // Expect an error
	is.Equal(callCount, 1) // Should NOT retry due to metadata
}

func TestRetryMiddleware_StopOnCertainErrors(t *testing.T) {
	is := is.New(t)
	callCount := 0

	// Mock handler that fails with different errors
	mockHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
		callCount++
		if callCount == 1 {
			return tc, errors.New("temporary error")
		}
		return tc, errors.New("critical error")
	})

	// Custom retry criteria: Retry only on "temporary error"
	customCriteria := func(tc minds.ThreadContext, attempt int, err error) bool {
		return err.Error() == "temporary error"
	}

	retry := middleware.Retry("error-specific-retry", retry.WithAttempts(5), retry.WithRetryCriteria(customCriteria))
	handler := retry.Wrap(mockHandler)

	ctx := minds.NewThreadContext(context.Background())
	_, err := handler.HandleThread(ctx, nil)

	is.True(err != nil)    // Expect an error
	is.Equal(callCount, 2) // Should stop retrying after "critical error"
}
