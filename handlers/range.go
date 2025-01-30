package handlers

import (
	"fmt"

	"github.com/chriscow/minds"
)

// Range executes a handler multiple times with different values. For each value in
// the provided list, the handler executes with the value stored in the thread
// context's metadata under the key "range_value". The handler supports middleware
// that will be applied to each iteration.
type Range struct {
	name       string
	handler    minds.ThreadHandler
	values     []any
	middleware []minds.Middleware
}

// NewRange creates a handler that processes a thread with a series of values.
//
// Parameters:
//   - name: Identifier for this range handler
//   - handler: The handler to execute for each value
//   - values: Values to iterate over
//
// Returns:
//   - A Range handler that processes the thread once for each value
//
// Example:
//
//	rng := NewRange("process",
//	    processHandler,
//	    "value1", "value2", "value3",
//	)
func NewRange(name string, handler minds.ThreadHandler, values ...any) *Range {
	if handler == nil {
		panic(fmt.Sprintf("%s: handler cannot be nil", name))
	}

	return &Range{
		name:       name,
		handler:    handler,
		values:     values,
		middleware: make([]minds.Middleware, 0),
	}
}

// Use adds middleware to the handler
func (r *Range) Use(middleware ...minds.Middleware) {
	r.middleware = append(r.middleware, middleware...)
}

// With returns a new handler with the provided middleware
func (r *Range) With(middleware ...minds.Middleware) minds.ThreadHandler {
	newRange := &Range{
		name:       r.name,
		handler:    r.handler,
		values:     r.values,
		middleware: append([]minds.Middleware{}, r.middleware...),
	}
	newRange.Use(middleware...)
	return newRange
}

// HandleThread implements the ThreadHandler interface
func (r *Range) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	current := tc

	for _, value := range r.values {
		if current.Context().Err() != nil {
			return current, current.Context().Err()
		}

		// Setup iteration context with value
		meta := current.Metadata()
		meta["range_value"] = value
		current = current.WithMetadata(meta)

		// Create wrapped handler for this iteration
		wrappedHandler := r.handler

		// Apply middleware in reverse order for proper nesting
		for i := len(r.middleware) - 1; i >= 0; i-- {
			wrappedHandler = r.middleware[i].Wrap(wrappedHandler)
		}

		var err error
		current, err = wrappedHandler.HandleThread(current, nil)
		if err != nil {
			return current, fmt.Errorf("%s: %w", r.name, err)
		}
	}

	// If there's a next handler, execute it with the final context
	if next != nil {
		return next.HandleThread(current, nil)
	}

	return current, nil
}

// String returns a string representation of the Range handler
func (r *Range) String() string {
	return fmt.Sprintf("Range(%s, %d values)", r.name, len(r.values))
}
