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