# v0.0.3

## Summary

- Added the new handlers: Cycle, First, Must, Policy, Range.


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