package handlers

import (
	"fmt"

	"github.com/chriscow/minds"
)

// Middleware interface defines how middleware can wrap handlers
type Middleware interface {
	Wrap(minds.ThreadHandler) minds.ThreadHandler
}

type MiddlewareFunc func(minds.ThreadHandler) minds.ThreadHandler

func (f MiddlewareFunc) Wrap(next minds.ThreadHandler) minds.ThreadHandler {
	return f(next)
}

// ThreadFlow manages multiple handlers and their middleware chains. Handlers are executed
// sequentially, with each handler receiving the result from the previous one. Middleware
// is scoped and applied to handlers added within that scope, allowing different middleware
// combinations for different processing paths.

// Example:

// 	validate := NewValidationHandler()
// 	process := NewProcessingHandler()
// 	seq := Sequential("main", validate, process)

// 	// Create flow and add global middleware
// 	flow := NewThreadFlow("example")
// 	flow.Use(NewLogging("audit"))

// 	// Add base handlers
// 	flow.Handle(seq)

// 	// Add handlers with scoped middleware
// 	flow.Group(func(f *ThreadFlow) {
// 	    f.Use(NewRetry("retry", 3))
// 	    f.Use(NewTimeout("timeout", 5))
// 	    f.Handle(NewContentProcessor("content"))
// 	    f.Handle(NewValidator("validate"))
// 	})

// result, err := flow.HandleThread(initialContext, nil)
//
//	if err != nil {
//	    log.Fatalf("Error in flow: %v", err)
//	}
//
// fmt.Println("Result:", result.Messages().Last().Content)
type ThreadFlow struct {
	name       string
	handlers   []minds.ThreadHandler
	middleware []Middleware
}

// NewThreadFlow creates a new ThreadFlow with the given handler
func NewThreadFlow(name string) *ThreadFlow {
	return &ThreadFlow{
		name: name,
	}
}

func (f *ThreadFlow) Handle(handler minds.ThreadHandler) {
	// Add a new handler with current middleware scope
	f.handlers = append(f.handlers, handler)
}

// Use adds middleware to the ThreadFlow. Middleware is executed in the order it's added,
// with each new middleware wrapping the previous ones.
func (f *ThreadFlow) Use(middleware ...Middleware) {
	f.middleware = append(f.middleware, middleware...)
}

// Group creates a new middleware scope. Middleware added within the group function
// will only apply to that group and won't affect parent scopes.
func (f *ThreadFlow) Group(fn func(*ThreadFlow)) {
	// Create a new flow with the current middleware stack
	groupFlow := &ThreadFlow{
		name: f.name + "-group",
		// middleware: append([]Middleware{}, f.middleware...), // Copy current middleware
	}

	// Let the function add handlers and middleware to the group
	fn(groupFlow)

	// Add all handlers from the group to the parent flow
	// Each handler gets wrapped with the group's middleware
	for _, handler := range groupFlow.handlers {
		wrappedHandler := handler
		// Apply group middleware in reverse order
		for i := len(groupFlow.middleware) - 1; i >= 0; i-- {
			wrappedHandler = groupFlow.middleware[i].Wrap(wrappedHandler)
		}
		f.handlers = append(f.handlers, wrappedHandler)
	}
}

// String returns a string representation of the ThreadFlow
func (f *ThreadFlow) String() string {
	return fmt.Sprintf("ThreadFlow(%s: %d middleware)", f.name, len(f.middleware))
}

func (f *ThreadFlow) HandleThread(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
	var result minds.ThreadContext = tc
	var err error

	// Each handler already has its group middleware applied during Group()
	// We only need to wrap with global middleware once
	for _, h := range f.handlers {
		chainedHandler := h

		// Apply global middleware only once
		for i := len(f.middleware) - 1; i >= 0; i-- {
			currentMiddleware := f.middleware[i]
			chainedHandler = currentMiddleware.Wrap(chainedHandler)
		}

		result, err = chainedHandler.HandleThread(result, nil)
		if err != nil {
			return result, err
		}
	}

	return result, nil
}
