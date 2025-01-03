package handlers

import (
	"context"
	"fmt"
	"sync"

	"github.com/chriscow/minds"
)

type anyHandler struct {
	name     string
	handlers []minds.ThreadHandler
}

// Any creates a handler that executes all provided handlers in parallel and returns
// when the first handler succeeds. If all handlers fail, it returns all errors.
func Any(name string, handlers ...minds.ThreadHandler) *anyHandler {
	return &anyHandler{name: name, handlers: handlers}
}

func (h *anyHandler) String() string {
	return fmt.Sprintf("Any(%s: %d handlers)", h.name, len(h.handlers))
}

func (h *anyHandler) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
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
