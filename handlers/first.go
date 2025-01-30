package handlers

import (
	"context"
	"fmt"
	"sync"

	"github.com/chriscow/minds"
)

// First represents a handler that executes multiple handlers in parallel and returns
// on first success. Each handler runs in its own goroutine, and execution of remaining
// handlers is canceled once a successful result is obtained.
type First struct {
	name       string
	handlers   []minds.ThreadHandler
	middleware []minds.Middleware
}

// NewFirst creates a handler that runs multiple handlers in parallel and returns on
// first success. If all handlers fail, an error containing all handler errors is
// returned. If no handlers are provided, the thread context is passed to the next
// handler unmodified.
//
// Parameters:
//   - name: Identifier for this parallel handler group
//   - handlers: Variadic list of handlers to execute in parallel
//
// Returns:
//   - A First handler configured with the provided handlers
//
// Example:
//
//	first := handlers.NewFirst("validation",
//	    validateA,
//	    validateB,
//	    validateC,
//	)
func NewFirst(name string, handlers ...minds.ThreadHandler) *First {
	return &First{
		name:     name,
		handlers: handlers,
	}
}

// Use applies middleware to the First handler, wrapping its child handlers.
func (f *First) Use(middleware ...minds.Middleware) {
	f.middleware = append(f.middleware, middleware...)
}

// With returns a new First handler with the provided middleware applied.
func (f *First) With(middleware ...minds.Middleware) minds.ThreadHandler {
	newFirst := &First{
		name:     f.name,
		handlers: append([]minds.ThreadHandler{}, f.handlers...),
	}
	newFirst.Use(middleware...)
	return newFirst
}

// HandleThread executes the First handler by running child handlers in parallel
// and returning on first success.
func (f *First) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	if len(f.handlers) == 0 {
		if next != nil {
			return next.HandleThread(tc, nil)
		}
		return tc, nil
	}

	// Create a cancellable context for parallel execution
	ctx, cancel := context.WithCancel(tc.Context())
	defer cancel()

	var wg sync.WaitGroup
	resultChan := make(chan struct {
		tc  minds.ThreadContext
		err error
	}, len(f.handlers))

	// Execute each handler in parallel
	for i, h := range f.handlers {
		wg.Add(1)
		go func(handler minds.ThreadHandler, idx int) {
			defer wg.Done()

			// Check for cancellation before starting
			if ctx.Err() != nil {
				return
			}

			// Clone the context to prevent modifications from affecting other handlers
			handlerCtx := tc.Clone().WithContext(ctx)

			// Add handler metadata for middleware context
			meta := handlerCtx.Metadata()
			meta["handler_name"] = fmt.Sprintf("h%d", idx+1)
			handlerCtx = handlerCtx.WithMetadata(meta)

			// Apply middleware in reverse order
			wrappedHandler := handler
			for i := len(f.middleware) - 1; i >= 0; i-- {
				wrappedHandler = f.middleware[i].Wrap(wrappedHandler)
			}

			// Execute the wrapped handler
			result, err := wrappedHandler.HandleThread(handlerCtx, nil)

			// Send result if not canceled
			select {
			case resultChan <- struct {
				tc  minds.ThreadContext
				err error
			}{result, err}:
				if err == nil {
					cancel() // Cancel other handlers on success
				}
			case <-ctx.Done():
			}
		}(h, i)
	}

	// Close result channel when all handlers complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect errors and watch for success
	var errors []error
	for result := range resultChan {
		if result.err == nil {
			if next != nil {
				return next.HandleThread(result.tc, nil)
			}
			return result.tc, nil
		}
		errors = append(errors, fmt.Errorf("%w", result.err))
	}

	// Handle context cancellation
	if ctx.Err() != nil {
		return tc, ctx.Err()
	}

	// Return combined errors if all handlers failed
	if len(errors) > 0 {
		return tc, fmt.Errorf("%s: all handlers failed: %v", f.name, errors)
	}

	return tc, nil
}

// String returns a string representation of the First handler.
func (f *First) String() string {
	return fmt.Sprintf("First(%s)", f.name)
}
