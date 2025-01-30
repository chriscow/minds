package handlers

import (
	"fmt"

	"github.com/chriscow/minds"
)

// Sequence executes a series of handlers in order, stopping if any handler returns
// an error. It supports middleware through the MiddlewareHandler interface.
type Sequence struct {
	name       string
	handlers   []minds.ThreadHandler
	middleware []minds.Middleware
}

// NewSequence creates a new Sequence handler with the given name and handlers.
// The sequence executes handlers in order, stopping on the first error.
//
// Example:
//
//	validate := NewValidationHandler()
//	process := NewProcessingHandler()
//
//	// Create a sequence with retry middleware
//	seq := NewSequence("main", validate, process)
//	seq.Use(RetryMiddleware())
//
//	// Or create a variation with different middleware
//	timeoutSeq := seq.With(TimeoutMiddleware())
//
//	result, err := seq.HandleThread(tc, nil)
func NewSequence(name string, handlers ...minds.ThreadHandler) *Sequence {
	return &Sequence{
		name:       name,
		handlers:   handlers,
		middleware: make([]minds.Middleware, 0),
	}
}

// Use adds middleware that will wrap each child handler
func (s *Sequence) Use(middleware ...minds.Middleware) {
	s.middleware = append(s.middleware, middleware...)
}

// HandleThread processes each handler in sequence, wrapping each with middleware
func (s *Sequence) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	current := tc

	// Execute each handler in sequence
	for _, h := range s.handlers {
		// Start with the base handler
		wrappedHandler := h

		// Apply middleware in reverse order for proper nesting
		for i := len(s.middleware) - 1; i >= 0; i-- {
			wrappedHandler = s.middleware[i].Wrap(wrappedHandler)
		}

		// Execute the wrapped handler
		var err error
		current, err = wrappedHandler.HandleThread(current, nil)
		if err != nil {
			return current, fmt.Errorf("%s: handler error: %w", s.name, err)
		}
	}

	// Execute next handler if provided
	if next != nil {
		return next.HandleThread(current, nil)
	}

	return current, nil
}

func (s *Sequence) String() string {
	return fmt.Sprintf("Sequence(%s)", s.name)
}
