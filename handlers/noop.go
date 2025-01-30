package handlers

import (
	"github.com/chriscow/minds"
)

// noop implements a handler that performs no operation and simply passes through
// the thread context unchanged. It's useful as a default handler in conditional
// flows or as a placeholder in handler chains.
type noop struct{}

// Noop creates a new no-operation handler that simply returns the thread context
// unchanged. It's useful as a default handler in Switch or If handlers when no
// action is desired for the default case.
//
// Example:
//
//	// Use as default in Switch
//	sw := Switch("type-switch", Noop(), cases...)
//
//	// Use as default in If
//	ih := If("type-check", condition, trueHandler, Noop())
func Noop() *noop {
	return &noop{}
}

// String returns the handler name for debugging and logging purposes.
func (n *noop) String() string {
	return "Noop"
}

// HandleThread implements the ThreadHandler interface by returning the thread
// context unchanged. If a next handler is provided, it will be called with
// the unchanged context.
func (n *noop) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	if next != nil {
		return next.HandleThread(tc, nil)
	}
	return tc, nil
}
