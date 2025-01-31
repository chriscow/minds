package middleware

import (
	"fmt"
	"time"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/middleware/retry"
)

// Retry creates a middleware that provides automatic retry functionality for handlers.
//
// The middleware offers configurable retry behavior, including:
//   - Customizable number of retry attempts
//   - Flexible backoff strategies
//   - Configurable retry criteria
//   - Optional timeout propagation
//
// Default behavior:
//   - 3 retry attempts
//   - No delay between attempts
//   - Retries on any error
//   - Timeout propagation enabled
//
// Example usage:
//
//	flow.Use(Retry("api-retry",
//	  retry.WithAttempts(5),
//	  retry.WithBackoff(retry.DefaultBackoff(time.Second)),
//	  retry.WithRetryCriteria(func(err error) bool {
//	    return errors.Is(err, io.ErrTemporary)
//	  }),
//	))
//
// The middleware will stop retrying if:
//   - An attempt succeeds
//   - Maximum attempts are reached
//   - Retry criteria returns false
//   - Context is canceled (if timeout propagation is enabled)
func Retry(name string, opts ...retry.Option) minds.Middleware {
	return minds.MiddlewareFunc(func(next minds.ThreadHandler) minds.ThreadHandler {
		return &retryMiddleware{
			name:   name,
			next:   next,
			config: configureRetry(opts...),
		}
	})
}

// retryMiddleware wraps a handler with retry logic.
type retryMiddleware struct {
	name   string
	next   minds.ThreadHandler
	config *retry.Options
}

// Wrap applies the retry behavior to a handler.
func (r *retryMiddleware) Wrap(next minds.ThreadHandler) minds.ThreadHandler {
	return &retryMiddleware{
		name:   r.name,
		next:   next,
		config: r.config,
	}
}

// HandleThread executes the handler, retrying on failure.
func (r *retryMiddleware) HandleThread(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
	if r.config.Attempts <= 0 {
		return r.next.HandleThread(tc, nil)
	}

	var lastErr error

	for attempt := 0; attempt < r.config.Attempts; attempt++ {
		// Handle timeout propagation
		if r.config.PropagateTimeout {
			select {
			case <-tc.Context().Done():
				return tc, fmt.Errorf("%s: context canceled: %w", r.name, tc.Context().Err())
			default:
			}
		}

		// Attempt execution
		result, err := r.next.HandleThread(tc, nil)
		if err == nil {
			return result, nil
		}

		// Stop retrying if error does not meet retry criteria
		if !r.config.ShouldRetry(tc, attempt, err) {
			return tc, fmt.Errorf("%s: retry stopped due to error: %w", r.name, err)
		}

		lastErr = err

		// Apply backoff strategy if defined
		if r.config.Backoff != nil {
			time.Sleep(r.config.Backoff(attempt))
		}
	}

	// Return last error if all attempts fail
	return tc, fmt.Errorf("%s: all %d attempts failed, last error: %w", r.name, r.config.Attempts, lastErr)
}

// configureRetry applies provided options to a retry configuration.
func configureRetry(opts ...retry.Option) *retry.Options {
	config := retry.NewDefaultOptions()
	for _, opt := range opts {
		opt(config)
	}
	return config
}
