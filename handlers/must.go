package handlers

import (
	"context"
	"fmt"
	"sync"

	"github.com/chriscow/minds"
)

// Must creates a handler that runs multiple handlers in parallel with all-success semantics.
//
// All handlers execute concurrently and must complete successfully. If any handler fails,
// remaining handlers are canceled and the first error encountered is returned. The timing
// of which error is returned is nondeterministic due to parallel execution.
//
// Parameters:
//   - name: Identifier for this parallel handler group.
//   - handlers: Variadic list of handlers that must all succeed.
//
// Returns:
//   - A handler that implements parallel execution with all-success semantics.
//
// Example:
//
//	must := handlers.Must("validation",
//	    validateA,
//	    validateB,
//	    validateC,
//	)
type mustHandler struct {
	name     string
	handlers []minds.ThreadHandler
}

// Must creates a handler that ensures all provided handlers succeed in parallel.
// If any handler fails, the others are canceled, and the first error is returned.
func Must(name string, handlers ...minds.ThreadHandler) *mustHandler {
	return &mustHandler{name: name, handlers: handlers}
}

func (h *mustHandler) String() string {
	return fmt.Sprintf("Must(%s: %d handlers)", h.name, len(h.handlers))
}

func (h *mustHandler) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	// The timing of which error is returned here is nondeterministic. The first error
	// encountered by any handler is returned, but the order in which handlers are
	// executed is not guaranteed. Also, if the context is canceled, the error from
	// the canceled context could be returned.
	ctx, cancel := context.WithCancel(tc.Context())
	defer cancel()

	tc = tc.WithContext(ctx)

	var wg sync.WaitGroup
	errChan := make(chan error, len(h.handlers))

	// Launch each handler
	for _, handler := range h.handlers {
		wg.Add(1)
		go func(handler minds.ThreadHandler) {
			defer wg.Done()
			if ctx.Err() != nil {
				return
			}
			if _, err := handler.HandleThread(tc, nil); err != nil {
				select {
				case errChan <- fmt.Errorf("%s: %w", h.name, err): // Wrap the error
					cancel()
				default:
				}
			}
		}(handler)
	}

	// Close the channel once all handlers are done
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Check errors
	var handlerError error
	for {
		select {
		case err, ok := <-errChan:
			if !ok {
				// No more errors
				if handlerError != nil {
					return tc, handlerError
				}
				if next != nil {
					return next.HandleThread(tc, nil)
				}
				return tc, nil
			}
			// Capture the first HandlerError
			if handlerError == nil {
				handlerError = err
			}

		case <-ctx.Done():
			// Ensure errors from errChan take precedence over context cancellation
			for err := range errChan {
				if handlerError == nil {
					handlerError = err
				}
			}
			if handlerError != nil {
				return tc, handlerError
			}
			return tc, ctx.Err()
		}
	}
}
