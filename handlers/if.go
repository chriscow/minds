package handlers

import (
	"fmt"

	"github.com/chriscow/minds"
)

// If represents a conditional handler that executes one of two handlers based on
// a condition evaluation. It's a simplified version of Switch for binary conditions.
type If struct {
	name            string
	condition       SwitchCondition
	trueHandler     minds.ThreadHandler
	fallbackHandler minds.ThreadHandler
	middleware      []minds.Middleware
}

// NewIf creates a new If handler that executes trueHandler when the condition evaluates
// to true, otherwise executes fallbackHandler if provided.
//
// Parameters:
//   - name: Identifier for this conditional handler
//   - condition: The condition to evaluate
//   - trueHandler: Handler to execute when condition is true
//   - fallbackHandler: Optional handler to execute when condition is false
//
// Returns:
//   - A handler that implements conditional execution based on the provided condition
//
// Example:
//
//	metadata := MetadataEquals{Key: "type", Value: "question"}
//	questionHandler := SomeQuestionHandler()
//	defaultHandler := DefaultHandler()
//
//	ih := NewIf("type-check", metadata, questionHandler, defaultHandler)
func NewIf(name string, condition SwitchCondition, trueHandler minds.ThreadHandler, fallbackHandler minds.ThreadHandler) *If {
	return &If{
		name:            name,
		condition:       condition,
		trueHandler:     trueHandler,
		fallbackHandler: fallbackHandler,
		middleware:      make([]minds.Middleware, 0),
	}
}

// Use adds middleware to the handler
func (h *If) Use(middleware ...minds.Middleware) {
	h.middleware = append(h.middleware, middleware...)
}

// With returns a new handler with the provided middleware
func (h *If) With(middleware ...minds.Middleware) minds.ThreadHandler {
	newIf := &If{
		name:            h.name,
		condition:       h.condition,
		trueHandler:     h.trueHandler,
		fallbackHandler: h.fallbackHandler,
		middleware:      append([]minds.Middleware{}, h.middleware...),
	}
	newIf.Use(middleware...)
	return newIf
}

// HandleThread implements the ThreadHandler interface
func (h *If) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	// Evaluate the condition
	matches, err := h.condition.Evaluate(tc)
	if err != nil {
		return tc, fmt.Errorf("%s: error evaluating condition: %w", h.name, err)
	}

	// Determine which handler to execute
	var handlerToExecute minds.ThreadHandler
	if matches && h.trueHandler != nil {
		handlerToExecute = h.trueHandler
	} else if !matches && h.fallbackHandler != nil {
		handlerToExecute = h.fallbackHandler
	}

	// If we have a handler to execute, wrap it with middleware and run it
	if handlerToExecute != nil {
		// Create wrapped handler
		wrappedHandler := handlerToExecute

		// Apply middleware in reverse order for proper nesting
		for i := len(h.middleware) - 1; i >= 0; i-- {
			wrappedHandler = h.middleware[i].Wrap(wrappedHandler)
		}

		return wrappedHandler.HandleThread(tc, next)
	}

	// No handler to execute, call next if provided
	if next != nil {
		return next.HandleThread(tc, nil)
	}

	return tc, nil
}

// String returns a string representation of the If handler
func (h *If) String() string {
	return fmt.Sprintf("If(%s)", h.name)
}
