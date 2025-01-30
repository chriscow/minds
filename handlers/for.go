package handlers

import (
	"fmt"

	"github.com/chriscow/minds"
)

// ForConditionFn defines a function type that controls loop continuation based on
// the current thread context and iteration count.
type ForConditionFn func(tc minds.ThreadContext, iteration int) bool

// For represents a handler that repeats processing based on iterations and conditions.
// It supports both fixed iteration counts and conditional execution through a
// continuation function.
type For struct {
	name       string
	handler    minds.ThreadHandler
	iterations int
	continueFn ForConditionFn
	middleware []minds.Middleware
}

// NewFor creates a handler that repeats processing based on iterations and conditions.
//
// A continuation function can optionally control the loop based on the ThreadContext
// and iteration count. The handler runs either until the iteration count is reached,
// the continuation function returns false, or infinitely if iterations is 0.
//
// Parameters:
//   - name: Identifier for this loop handler
//   - iterations: Number of iterations (0 for infinite)
//   - handler: The handler to repeat
//   - fn: Optional function controlling loop continuation
//
// Returns:
//   - A handler that implements controlled repetition of processing
//
// Example:
//
//	loop := NewFor("validation", 3, validateHandler, func(tc ThreadContext, i int) bool {
//	    return tc.ShouldContinue()
//	})
func NewFor(name string, iterations int, handler minds.ThreadHandler, fn ForConditionFn) *For {
	if handler == nil {
		panic(fmt.Sprintf("%s: handler cannot be nil", name))
	}

	return &For{
		name:       name,
		handler:    handler,
		iterations: iterations,
		continueFn: fn,
		middleware: make([]minds.Middleware, 0),
	}
}

// Use adds middleware to the handler
func (f *For) Use(middleware ...minds.Middleware) {
	f.middleware = append(f.middleware, middleware...)
}

// With returns a new handler with the provided middleware
func (f *For) With(middleware ...minds.Middleware) minds.ThreadHandler {
	newFor := &For{
		name:       f.name,
		handler:    f.handler,
		iterations: f.iterations,
		continueFn: f.continueFn,
		middleware: append([]minds.Middleware{}, f.middleware...),
	}
	newFor.Use(middleware...)
	return newFor
}

// HandleThread implements the ThreadHandler interface
func (f *For) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	iter := 0
	current := tc

	// Continue until iteration count is reached or condition function returns false
	for f.iterations == 0 || iter < f.iterations {
		// Check continuation condition if provided
		if f.continueFn != nil && !f.continueFn(current, iter) {
			break
		}

		// Check for context cancellation
		if current.Context().Err() != nil {
			return current, current.Context().Err()
		}

		// Setup iteration context with metadata
		iterCtx := current.Clone()
		meta := iterCtx.Metadata()
		meta["iteration"] = iter
		iterCtx = iterCtx.WithMetadata(meta)

		// Create wrapped handler for this iteration
		wrappedHandler := f.handler

		// Apply middleware in reverse order for proper nesting
		for i := len(f.middleware) - 1; i >= 0; i-- {
			wrappedHandler = f.middleware[i].Wrap(wrappedHandler)
		}

		// Execute the wrapped handler
		var err error
		current, err = wrappedHandler.HandleThread(iterCtx, nil)
		if err != nil {
			return current, fmt.Errorf("%s: iteration %d failed: %w", f.name, iter, err)
		}

		iter++
	}

	// If there's a next handler, execute it with the final context
	if next != nil {
		return next.HandleThread(current, nil)
	}

	return current, nil
}

// String returns a string representation of the For handler
func (f *For) String() string {
	iterations := "infinite"
	if f.iterations > 0 {
		iterations = fmt.Sprintf("%d", f.iterations)
	}
	return fmt.Sprintf("For(%s, %s iterations)", f.name, iterations)
}
