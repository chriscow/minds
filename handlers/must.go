package handlers

import (
	"context"
	"fmt"
	"sync"

	"github.com/chriscow/minds"
)

type mustHandler struct {
	name       string
	handlers   []minds.ThreadHandler
	aggregator ResultAggregator
}

type HandlerResult struct {
	Handler minds.ThreadHandler
	Context minds.ThreadContext
	Error   error
}

type ResultAggregator func([]HandlerResult) (minds.ThreadContext, error)

// DefaultAggregator combines results by merging contexts in order. Merging is done by
// starting with the first successful context and merging all subsequent successful
// contexts into it. If no successful contexts are found, an error is returned.
//
// Metadata merging is done with the minds.KeepNew strategy, which overwrites existing
// values with new values. Messages are appended in order.
//
// Parameters:
//   - results: List of handler results to aggregate.
//
// Returns:
//   - A single thread context that combines all successful results.
func DefaultAggregator(results []HandlerResult) (minds.ThreadContext, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to aggregate")
	}

	// Start with the first successful context
	var base minds.ThreadContext
	for _, r := range results {
		if r.Error == nil && r.Context != nil {
			base = r.Context
			break
		}
	}

	if base == nil {
		return nil, fmt.Errorf("no successful results found")
	}

	// Merge remaining successful contexts
	for _, r := range results {
		if r.Error != nil || r.Context == nil {
			continue
		}
		if r.Context != base { // Skip if it's the same context we started with
			base = mergeContexts(base, r.Context)
		}
	}

	return base, nil
}

// Must returns a handler that runs multiple handlers in parallel with all-success semantics.
//
// All handlers execute concurrently and must complete successfully. If any handler fails,
// remaining handlers are canceled and the first error encountered is returned. The timing
// of which error is returned is nondeterministic due to parallel execution.
//
// Parameters:
//   - name: Identifier for this parallel handler group.
//   - handlers: Variadic list of handlers that must all succeed.
//
// Returns:
//   - A handler that implements parallel execution with all-success semantics.
//
// Example:
//
//	must := handlers.Must("validation",
//	    validateA,
//	    validateB,
//	    validateC,
//	)
func Must(name string, agg ResultAggregator, handlers ...minds.ThreadHandler) *mustHandler {
	if agg == nil {
		agg = DefaultAggregator
	}

	h := &mustHandler{
		name:       name,
		handlers:   handlers,
		aggregator: agg,
	}
	return h
}

func (h *mustHandler) String() string {
	return fmt.Sprintf("Must(%s: %d handlers)", h.name, len(h.handlers))
}

func (h *mustHandler) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	// The timing of which error is returned here is nondeterministic. The first error
	// encountered by any handler is returned, but the order in which handlers are
	// executed is not guaranteed. Also, if the context is canceled, the error from
	// the canceled context could be returned.
	ctx, cancel := context.WithCancel(tc.Context())
	defer cancel()

	tc = tc.WithContext(ctx)

	var wg sync.WaitGroup
	resultChan := make(chan HandlerResult, len(h.handlers))

	// Launch each handler
	for _, handler := range h.handlers {
		wg.Add(1)
		go func(handler minds.ThreadHandler) {
			defer wg.Done()

			result := HandlerResult{Handler: handler}

			if ctx.Err() != nil {
				result.Error = ctx.Err()
				resultChan <- result
				return
			}

			newCtx, err := handler.HandleThread(tc, nil)
			result.Context = newCtx
			result.Error = err

			if err != nil {
				result.Error = fmt.Errorf("%s: %w", h.name, err)
				cancel() // Cancel other handlers on first error
			}

			resultChan <- result
		}(handler)
	}

	// Close channel when all handlers complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var results []HandlerResult
	var firstError error

	// Collect results
	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				// No more results
				if firstError != nil {
					return tc, firstError
				}

				// Use aggregator to combine results
				finalTC, err := h.aggregator(results)
				if err != nil {
					return tc, fmt.Errorf("%s aggregation: %w", h.name, err)
				}

				if next != nil {
					return next.HandleThread(finalTC, nil)
				}
				return finalTC, nil
			}

			results = append(results, result)
			if result.Error != nil && firstError == nil {
				firstError = result.Error
			}

		case <-ctx.Done():
			// Drain remaining results
			for result := range resultChan {
				results = append(results, result)
				if result.Error != nil && firstError == nil {
					firstError = result.Error
				}
			}
			if firstError != nil {
				return tc, firstError
			}
			return tc, ctx.Err()
		}
	}
}

func mergeContexts(base, new minds.ThreadContext) minds.ThreadContext {
	mergedMeta := base.Metadata().Merge(new.Metadata(), minds.KeepNew)
	base = base.WithMetadata(mergedMeta)
	base.AppendMessages(new.Messages()...)
	return base
}
