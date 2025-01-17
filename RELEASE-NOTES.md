## v0.0.5 - 2023-10-05

## Summary

### Added
- **ThreadContext**: 
    - Added `SetKeyValue`, changed to copy on write strategy.
    - Added `Merge` to `Metadata` with merging strategies.
    - `AppendMessages` no longer returns ThreadContext; mutates in place.
- **ThreadFlow**: Added `ThreadFlow` scoped middleware handler and handler chains.


### Changed
- **Gemini/OpenAI providers**: Default to retryable HTTP client.
- **Must Handler**: Now accepts result aggregator.
- **First, For, Range and Sequential**: Removed `Use` middleware method in lieu of `ThreadFlow` middleware pattern.


## New Handlers

### **ThreadFlow**
ThreadFlow manages multiple handlers and their middleware chains. Handlers are executed
sequentially, with each handler receiving the result from the previous one. Middleware
is scoped and applied to handlers added within that scope, allowing different middleware
combinations for different processing paths.

Example:

	validate := NewValidationHandler()
	process := NewProcessingHandler()
	seq := Sequential("main", validate, process)

	// Create flow and add global middleware
	flow := NewThreadFlow("example")
	flow.Use(NewLogging("audit"))

	// Add base handlers
	flow.Handle(seq)

	// Add handlers with scoped middleware
	flow.Group(func(f *ThreadFlow) {
	    f.Use(NewRetry("retry", 3))
	    f.Use(NewTimeout("timeout", 5))
	    f.Handle(NewContentProcessor("content"))
	    f.Handle(NewValidator("validate"))
	})

	result, err := flow.HandleThread(initialContext, nil)
	if err != nil {
	    log.Fatalf("Error in flow: %v", err)
	}
	fmt.Println("Result:", result.Messages().Last().Content)

## New Middleware

### **Retry**
Retry creates a middleware that automatically retries failed handler executions
with configurable backoff and retry criteria.

The middleware supports customization through options including:
  - Number of retry attempts
  - Backoff strategy between attempts
  - Retry criteria based on error evaluation
  - Timeout propagation control

If no options are provided, the middleware uses default settings:
  - 3 retry attempts
  - No delay between attempts
  - Retries on any error
  - Timeout propagation enabled

Parameters:
  - name: Identifier for the middleware instance
  - opts: Optional configuration using retry.Option functions

Example usage:

	flow.Use(Retry("api-retry",
	  retry.WithAttempts(5),
	  retry.WithBackoff(retry.DefaultBackoff(time.Second)),
	  retry.WithRetryCriteria(func(err error) bool {
	    return errors.Is(err, io.ErrTemporary)
	  }),
	))

The middleware stops retrying if:
  - An attempt succeeds
  - The maximum number of attempts is reached
  - The retry criteria returns false
  - Context cancellation (if timeout propagation is enabled)

Returns:
  - A middleware that implements the retry logic around a handler

# v0.0.4

## Summary

- Added new handler: switch
- Renamed ThreadContext `AppendMessage` to `AppendMessages(...Message)`


## New Handlers

### **Switch**

Switch creates a new Switch handler that executes the first matching case's
handler. If no cases match, it executes the default handler. The name
parameter is used for debugging and logging purposes.

The SwitchCase struct pairs a `SwitchCondition` interface with a handler.
When the condition evaluates to true, the corresponding handler is executed.

Example:

The MetadataEquals condition checks if the metadata key "type" equals "question"
```go
metadata := MetadataEquals{Key: "type", Value: "question"}
questionHandler := SomeQuestionHandler()
defaultHandler := DefaultHandler()

sw := Switch("type-switch",
	defaultHandler,
	SwitchCase{metadata, questionHandler},
)
```


# v0.0.3

## Summary

- Added the new handlers: For, First, Must, Policy, Range.
- Greatly simplified Response interface (for now). More changes coming.

## New Handlers

### **First** 

The `First` handler executes all provided handlers in parallel and
returns when the first handler succeeds. If all handlers fail, it returns all
errors.

### **For** 

The `For` handler repeats the provided handler for a specified
number of iterations or infinitely. The handler can be used to repeat a handler
a fixed number of times or until a condition is met.

### **Must** 

The `Must` handler executes all provided handlers in parallel and
returns when all handlers succeed. If any handler fails, it cancels execution of
any remaining handlers.

### **Policy** 

The `Policy` handler uses the LLM to validate thread content against a given policy.
The result is processed by the optional result function. If the result function is nil,
the handler defaults to checking the `Valid` field of the validation result.

### **Range**

The `Range` handler processes a thread with a series of values.

For each value in the provided list, the handler executes with the value stored
in the thread context's metadata under the key "range_value". An optional middleware
handler can be used to wrap each iteration.


# v0.0.2

# v0.0.1