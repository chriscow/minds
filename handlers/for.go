package handlers

import (
	"fmt"

	"github.com/chriscow/minds"
)

type looper struct {
	name       string
	handler    minds.ThreadHandler
	middleware minds.ThreadHandler
	iterations int // number of times to cycle, 0 means infinite
	continueFn ForConditionFn
}

type ForConditionFn func(tc minds.ThreadContext, iteration int) bool

// For creates a handler that repeats processing based on iterations and conditions.
//
// A continuation function can optionally control the loop based on the ThreadContext
// and iteration count. The handler runs either until the iteration count is reached,
// the continuation function returns false, or infinitely if iterations is 0.
//
// Parameters:
//   - name: Identifier for this loop handler.
//   - iterations: Number of iterations (0 for infinite).
//   - handler: The handler to repeat.
//   - fn: Optional function controlling loop continuation.
//
// Returns:
//   - A handler that implements controlled repetition of processing.
//
// Example:
//
//	loop := handlers.For("validation", 3, validateHandler, func(tc ThreadContext, i int) bool {
//	    return tc.ShouldContinue()
//	})
func For(name string, iterations int, handler minds.ThreadHandler, fn ForConditionFn) *looper {
	return &looper{
		name:       name,
		handler:    handler,
		iterations: iterations,
		continueFn: fn,
	}
}

func (c *looper) Use(handler minds.ThreadHandler) {
	c.middleware = handler
}

func (c *looper) String() string {
	return fmt.Sprintf("For: %s", c.name)
}

func (c *looper) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	iter := 0
	for c.iterations == 0 || iter < c.iterations {
		if c.continueFn != nil && !c.continueFn(tc, iter) {
			break
		}
		if c.middleware != nil {
			var err error
			tc, err = c.middleware.HandleThread(tc, c.handler)
			if err != nil {
				return tc, err
			}
		} else {
			var err error
			tc, err = c.handler.HandleThread(tc, nil)
			if err != nil {
				return tc, err
			}
		}
		iter++
	}

	if next != nil {
		return next.HandleThread(tc, nil)
	}

	return tc, nil
}
