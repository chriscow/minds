package handlers

import (
	"context"
	"fmt"
	"sync"

	"github.com/chriscow/minds"
)

// HandlerResult represents the result of executing a single handler
type HandlerResult struct {
	Handler minds.ThreadHandler
	Context minds.ThreadContext
	Error   error
}

// ResultAggregator defines a function type for combining multiple handler results
type ResultAggregator func([]HandlerResult) (minds.ThreadContext, error)

// Must represents a handler that runs multiple handlers in parallel requiring all to succeed
type Must struct {
	name       string
	handlers   []minds.ThreadHandler
	middleware []minds.Middleware
	aggregator ResultAggregator
}

// NewMust creates a handler that runs multiple handlers in parallel with all-success semantics.
//
// All handlers execute concurrently and must complete successfully. If any handler fails,
// remaining handlers are canceled and the first error encountered is returned. The timing
// of which error is returned is nondeterministic due to parallel execution.
//
// Parameters:
//   - name: Identifier for this parallel handler group
//   - agg: ResultAggregator function to combine successful results
//   - handlers: Variadic list of handlers that must all succeed
//
// Returns:
//   - A handler that implements parallel execution with all-success semantics
//
// Example:
//
//	must := handlers.NewMust("validation",
//	    handlers.DefaultAggregator,
//	    validateA,
//	    validateB,
//	    validateC,
//	)
func NewMust(name string, agg ResultAggregator, handlers ...minds.ThreadHandler) *Must {
	if agg == nil {
		agg = DefaultAggregator
	}
	return &Must{
		name:       name,
		handlers:   handlers,
		aggregator: agg,
	}
}

// DefaultAggregator combines results by merging contexts in order. Merging is done by
// starting with the first successful context and merging all subsequent successful
// contexts into it. If no successful contexts are found, an error is returned.
//
// Metadata merging is done with the minds.KeepNew strategy, which overwrites existing
// values with new values. Messages are appended in order.
//
// Parameters:
//   - results: List of handler results to aggregate
//
// Returns:
//   - A single thread context that combines all successful results
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

// Use applies middleware to the Must handler, wrapping its child handlers.
func (m *Must) Use(middleware ...minds.Middleware) {
	m.middleware = append(m.middleware, middleware...)
}

// With returns a new Must handler with additional middleware, preserving existing state.
func (m *Must) With(middleware ...minds.Middleware) *Must {
	newMust := &Must{
		name:       m.name,
		handlers:   append([]minds.ThreadHandler{}, m.handlers...),
		middleware: append([]minds.Middleware{}, m.middleware...),
		aggregator: m.aggregator,
	}
	newMust.Use(middleware...)
	return newMust
}

// HandleThread executes all child handlers in parallel, requiring all to succeed.
func (m *Must) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	if len(m.handlers) == 0 {
		if next != nil {
			return next.HandleThread(tc, nil)
		}
		return tc, nil
	}

	// Create a cancellable context for parallel execution
	ctx, cancel := context.WithCancel(tc.Context())
	defer cancel()

	tc = tc.WithContext(ctx)

	var wg sync.WaitGroup
	resultChan := make(chan HandlerResult, len(m.handlers))

	// Execute each handler in parallel
	for _, h := range m.handlers {
		wg.Add(1)
		go func(handler minds.ThreadHandler) {
			defer wg.Done()

			result := HandlerResult{Handler: handler}

			// Check for cancellation before executing
			if ctx.Err() != nil {
				result.Error = ctx.Err()
				resultChan <- result
				return
			}

			// Clone the context for isolation
			handlerCtx := tc.Clone().WithContext(ctx)

			// Apply middleware in reverse order
			wrappedHandler := handler
			for i := len(m.middleware) - 1; i >= 0; i-- {
				wrappedHandler = m.middleware[i].Wrap(wrappedHandler)
			}

			// Execute wrapped handler
			newCtx, err := wrappedHandler.HandleThread(handlerCtx, nil)
			result.Context = newCtx
			result.Error = err

			if err != nil {
				result.Error = fmt.Errorf("%s: %w", m.name, err)
				cancel() // Cancel all other handlers on failure
			}

			resultChan <- result
		}(h)
	}

	// Close channel when all handlers complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var results []HandlerResult
	var firstError error

	// Collect results
	for result := range resultChan {
		results = append(results, result)
		if result.Error != nil && firstError == nil {
			firstError = result.Error
		}
	}

	if firstError != nil {
		return tc, firstError
	}

	// Use aggregator to combine results
	finalTC, err := m.aggregator(results)
	if err != nil {
		return tc, fmt.Errorf("%s aggregation: %w", m.name, err)
	}

	if next != nil {
		return next.HandleThread(finalTC, nil)
	}
	return finalTC, nil
}

// String returns a string representation of the Must handler.
func (m *Must) String() string {
	return fmt.Sprintf("Must(%s: %d handlers)", m.name, len(m.handlers))
}

func mergeContexts(base, new minds.ThreadContext) minds.ThreadContext {
	mergedMeta := base.Metadata().Merge(new.Metadata(), minds.KeepNew)
	base = base.WithMetadata(mergedMeta)
	base.AppendMessages(new.Messages()...)
	return base
}
