package handlers

import (
	"context"
	"fmt"
	"sync"

	"github.com/chriscow/minds"
)

// mustHandler executes multiple handlers in parallel. If any handler fails, it cancels
// the remaining handlers and returns the first encountered error.
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
	ctx, cancel := context.WithCancel(tc.Context())
	defer cancel()

	var wg sync.WaitGroup
	errChan := make(chan error, len(h.handlers))

	for _, handler := range h.handlers {
		wg.Add(1)
		go func(handler minds.ThreadHandler) {
			defer wg.Done()

			// Check for context cancellation before execution
			if ctx.Err() != nil {
				return
			}

			// Execute the handler
			if _, err := handler.HandleThread(tc, nil); err != nil {
				select {
				case errChan <- err:
					cancel() // Cancel other handlers on error
				default:
					// Avoid blocking if another error has already been sent
				}
			}
		}(handler)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Check for errors or context cancellation
	select {
	case err := <-errChan:
		if err != nil {
			return tc, err // Return the first handler error
		}
	case <-ctx.Done():
		return tc, ctx.Err() // Return context cancellation error
	}

	// Call the next handler if no errors occurred
	if next != nil {
		return next.HandleThread(tc, nil)
	}

	return tc, nil
}
