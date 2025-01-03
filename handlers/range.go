package handlers

import (
	"fmt"

	"github.com/chriscow/minds"
)

type ranger struct {
	name       string
	handler    minds.ThreadHandler
	middleware minds.ThreadHandler
	values     []interface{}
}

// Range creates a handler that processes a thread with a series of values.
//
// For each value in the provided list, the handler executes with the value stored
// in the thread context's metadata under the key "range_value". An optional middleware
// handler can be used to wrap each iteration.
//
// Parameters:
//   - name: Identifier for this range handler.
//   - handler: The handler to execute for each value.
//   - values: Variadic list of values to iterate over.
//
// Returns:
//   - A handler that processes the thread once for each value.
//
// Example:
//
//	range := handlers.Range("process",
//	    processHandler,
//	    "value1", "value2", "value3",
//	)
func Range(name string, handler minds.ThreadHandler, values ...interface{}) *ranger {
	return &ranger{
		name:    name,
		handler: handler,
		values:  values,
	}
}

func (r *ranger) Use(handler minds.ThreadHandler) {
	r.middleware = handler
}

func (r *ranger) String() string {
	return fmt.Sprintf("Range: %s", r.name)
}

func (r *ranger) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	for _, value := range r.values {
		meta := tc.Metadata()
		meta["range_value"] = value
		tc = tc.WithMetadata(meta)

		if r.middleware != nil {
			var err error
			tc, err = r.middleware.HandleThread(tc, r.handler)
			if err != nil {
				return tc, err
			}
		} else {
			var err error
			tc, err = r.handler.HandleThread(tc, nil)
			if err != nil {
				return tc, err
			}
		}
	}

	if next != nil {
		return next.HandleThread(tc, nil)
	}

	return tc, nil
}
