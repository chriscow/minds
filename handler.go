package minds

import "fmt"

type ThreadHandler interface {
	HandleThread(thread ThreadContext, next ThreadHandler) (ThreadContext, error)
}

type ThreadHandlerFunc func(thread ThreadContext, next ThreadHandler) (ThreadContext, error)

func (f ThreadHandlerFunc) HandleThread(thread ThreadContext, next ThreadHandler) (ThreadContext, error) {
	return f(thread, next)
}

type NoopThreadHandler struct{}

func (h NoopThreadHandler) HandleThread(tc ThreadContext, next ThreadHandler) (ThreadContext, error) {
	return tc, nil
}

type MiddlewareHandler interface {
	ThreadHandler

	// Use adds middleware to the handler
	Use(middleware ...Middleware)

	// WithMiddleware returns a new handler with the provided middleware applied
	With(middleware ...Middleware) ThreadHandler
}

// HandlerExecutor defines a strategy for executing child handlers. It receives
// the current ThreadContext, a slice of handlers to execute, and an optional next
// handler to call after all handlers have been processed. The executor is responsible
// for defining how and when each handler is executed.
type HandlerExecutor func(tc ThreadContext, handlers []ThreadHandler, next ThreadHandler) (ThreadContext, error)

// Middleware represents a function that can wrap a ThreadHandler
type Middleware interface {
	Wrap(next ThreadHandler) ThreadHandler
}

// MiddlewareFunc is a function that implements the Middleware interface
type MiddlewareFunc func(next ThreadHandler) ThreadHandler

func (f MiddlewareFunc) Wrap(next ThreadHandler) ThreadHandler {
	return f(next)
}

// NewMiddleware creates a new middleware that runs a function before passing control
// to the next handler. The provided function can modify the ThreadContext and
// return an error to halt processing.
func NewMiddleware(name string, fn func(tc ThreadContext) error) Middleware {
	return MiddlewareFunc(func(next ThreadHandler) ThreadHandler {
		return ThreadHandlerFunc(func(tc ThreadContext, _ ThreadHandler) (ThreadContext, error) {
			if err := fn(tc); err != nil {
				return tc, fmt.Errorf("%s: %w", name, err)
			}
			if next == nil {
				next = NoopThreadHandler{}
			}
			return next.HandleThread(tc, nil)
		})
	})
}

// SupportsMiddleware checks if the given handler supports middleware operations
func SupportsMiddleware(h ThreadHandler) (MiddlewareHandler, bool) {
	mh, ok := h.(MiddlewareHandler)
	return mh, ok
}
