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
		return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			// Apply retry configuration
			config := retry.NewDefaultOptions()
			for _, opt := range opts {
				opt(config)
			}

			// Validate configuration
			if config.Attempts <= 0 {
				return tc, nil
			}

			var lastErr error

			// Retry loop
			for attempt := 0; attempt < config.Attempts; attempt++ {
				// Check for context cancellation if timeout propagation is enabled
				if config.PropagateTimeout {
					select {
					case <-tc.Context().Done():
						return tc, fmt.Errorf("%s: context canceled: %w", name, tc.Context().Err())
					default:
					}
				}

				// Attempt to execute the handler
				result, err := next.HandleThread(tc, nil)
				if err == nil {
					return result, nil
				}

				// Check if the error meets retry criteria
				if !config.ShouldRetry(err) {
					return tc, fmt.Errorf("%s: retry stopped due to error: %w", name, err)
				}

				lastErr = err

				// Apply backoff strategy if configured
				if config.Backoff != nil {
					backoffDuration := config.Backoff(attempt)
					time.Sleep(backoffDuration)
				}
			}

			// Return error if all attempts fail
			return tc, fmt.Errorf("%s: all %d attempts failed, last error: %w",
				name, config.Attempts, lastErr)
		})
	})
}
