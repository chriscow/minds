package handlers

import (
	"context"
	"fmt"
	"sync"

	"github.com/chriscow/minds"
)

type firstHandler struct {
	name     string
	handlers []minds.ThreadHandler
}

// First creates a handler that runs multiple handlers in parallel and returns on first success.
//
// Handlers are executed concurrently and the first successful result is returned.
// If all handlers fail, an error containing all handler errors is returned.
// Processing is canceled for remaining handlers once a successful result is obtained.
// If no handlers are provided, the thread context is passed to the next handler.
//
// Parameters:
//   - name: Identifier for this parallel handler group.
//   - handlers: Variadic list of handlers to execute in parallel.
//
// Returns:
//   - A handler that implements parallel execution with first-success semantics.
//
// Example:
//
//	first := handlers.First("validation",
//	    validateA,
//	    validateB,
//	    validateC,
//	)
func First(name string, handlers ...minds.ThreadHandler) *firstHandler {
	return &firstHandler{name: name, handlers: handlers}
}

func (h *firstHandler) String() string {
	return fmt.Sprintf("Any(%s: %d handlers)", h.name, len(h.handlers))
}

func (h *firstHandler) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	if len(h.handlers) == 0 {
		if next != nil {
			return next.HandleThread(tc, nil)
		}
		return tc, nil
	}

	ctx, cancel := context.WithCancel(tc.Context())
	defer cancel()

	var wg sync.WaitGroup
	resultChan := make(chan struct {
		tc  minds.ThreadContext
		err error
	}, len(h.handlers))

	for _, handler := range h.handlers {
		wg.Add(1)
		go func(handler minds.ThreadHandler) {
			defer wg.Done()

			if ctx.Err() != nil {
				return
			}

			newTc, err := handler.HandleThread(tc, nil)
			select {
			case resultChan <- struct {
				tc  minds.ThreadContext
				err error
			}{newTc, err}:
				if err == nil {
					cancel() // Cancel other handlers on success
				}
			case <-ctx.Done():
			}
		}(handler)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var errors []error
	for result := range resultChan {
		if result.err == nil {
			if next != nil {
				return next.HandleThread(result.tc, nil)
			}
			return result.tc, nil
		}
		errors = append(errors, result.err)
	}

	if ctx.Err() != nil {
		return tc, ctx.Err()
	}

	if len(errors) > 0 {
		return tc, fmt.Errorf("all handlers failed: %v", errors)
	}

	return tc, nil
}
