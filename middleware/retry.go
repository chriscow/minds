package middleware

import (
	"fmt"
	"time"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/chriscow/minds/middleware/retry"
)

// Retry creates a middleware that automatically retries failed handler executions
// with configurable backoff and retry criteria.

// The middleware supports customization through options including:
//   - Number of retry attempts
//   - Backoff strategy between attempts
//   - Retry criteria based on error evaluation
//   - Timeout propagation control

// If no options are provided, the middleware uses default settings:
//   - 3 retry attempts
//   - No delay between attempts
//   - Retries on any error
//   - Timeout propagation enabled

// Parameters:
//   - name: Identifier for the middleware instance
//   - opts: Optional configuration using retry.Option functions

// Example usage:

// 	flow.Use(Retry("api-retry",
// 	  retry.WithAttempts(5),
// 	  retry.WithBackoff(retry.DefaultBackoff(time.Second)),
// 	  retry.WithRetryCriteria(func(err error) bool {
// 	    return errors.Is(err, io.ErrTemporary)
// 	  }),
// 	))

// The middleware stops retrying if:
//   - An attempt succeeds
//   - The maximum number of attempts is reached
//   - The retry criteria returns false
//   - Context cancellation (if timeout propagation is enabled)

// Returns:
//   - A middleware that implements the retry logic around a handler
func Retry(name string, opts ...retry.Option) handlers.Middleware {
	config := retry.NewDefaultOptions()
	for _, opt := range opts {
		opt(config)
	}

	return &middleware{
		name: name,
		fn: func(next minds.ThreadHandler) minds.ThreadHandler {
			return &retryHandler{
				name:    name,
				next:    next,
				options: config,
			}
		},
	}
}

type retryHandler struct {
	name    string
	next    minds.ThreadHandler
	options *retry.Options
}

func (h *retryHandler) HandleThread(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
	if h.options == nil {
		return tc, fmt.Errorf("%s: options not set", h.name)
	}

	if h.options.Attempts <= 0 {
		// no-op
		return tc, nil
	}

	var lastErr error

	for i := 0; i < h.options.Attempts; i++ {
		// Stop retrying if the context is canceled
		if h.options.PropagateTimeout {
			select {
			case <-tc.Context().Done():
				return tc, fmt.Errorf("%s: context canceled: %w", h.name, tc.Context().Err())
			default:
			}
		}

		// Attempt execution
		result, err := h.next.HandleThread(tc, nil)
		if err == nil {
			return result, nil
		}

		// Check retry criteria
		if !h.options.ShouldRetry(err) {
			return tc, fmt.Errorf("%s: retry stopped due to error: %w", h.name, err)
		}

		lastErr = err

		// Apply backoff
		if h.options.Backoff != nil {
			time.Sleep(h.options.Backoff(i))
		}
	}

	return tc, fmt.Errorf("%s: all %d attempts failed, last error: %w", h.name, h.options.Attempts, lastErr)
}
