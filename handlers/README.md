# First Handler

### Documentation for the `First` Handler

The `First` handler is designed to execute multiple handlers in parallel and return the result of the first successful handler. This is particularly useful in scenarios where you have multiple ways to achieve the same goal, and you want to proceed as soon as one of them succeeds. If all handlers fail, it aggregates the errors and returns them.

---

### Analogy: The Race to Success

Think of the `First` handler as a **race** where multiple runners (handlers) are competing to reach the finish line (success). The moment one runner crosses the line, the race is over, and the other runners stop. If no runner finishes successfully, the race is considered a failure, and you get a report of what went wrong with each runner.

This pattern is similar to:
- **Load Balancing with Fallbacks**: Trying multiple servers to handle a request and using the first successful response.
- **Redundant Validation**: Validating data using multiple methods and accepting the first valid result.
- **Competing Algorithms**: Running different algorithms in parallel and using the first one that produces a valid result.

---

### Key Features:
1. **Concurrent Execution**: All handlers run in parallel, maximizing efficiency.
2. **First Success Wins**: As soon as one handler succeeds, the rest are canceled.
3. **Error Aggregation**: If all handlers fail, all errors are collected and returned.
4. **Graceful Fallback**: If no handlers are provided, the input is passed through to the next handler in the chain.

---

### Example Use Case

Imagine you’re building a conversational AI system, and you want to validate user input using multiple methods. You might have:
- A **rule-based validator** (fast but less accurate).
- A **machine learning validator** (slower but more accurate).
- A **fallback validator** (basic checks).

You want to use the first validator that succeeds, and if all fail, return an error.

```go
first := handlers.First("validation",
    ruleBasedValidator,
    mlValidator,
    fallbackValidator,
)

result, err := first.HandleThread(inputContext, nil)
if err != nil {
    log.Fatalf("Validation failed: %v", err)
}
fmt.Println("Validated input:", result)
```

---

### How It Works

1. **Setup**: The `First` handler is initialized with a name and a list of handlers to execute.
2. **Execution**:
   - All handlers are run concurrently using goroutines.
   - The first handler to succeed sends its result to a channel.
   - Once a result is received, the context is canceled to stop the remaining handlers.
3. **Result Handling**:
   - If a handler succeeds, its result is returned.
   - If all handlers fail, an aggregated error is returned.
4. **Fallback**: If no handlers are provided, the input is passed through to the next handler in the chain.

---

### Code Breakdown

#### Initialization
```go
first := handlers.First("validation",
    validateA,
    validateB,
    validateC,
)
```
- `"validation"`: A name for the handler group (useful for debugging and logging).
- `validateA`, `validateB`, `validateC`: Handlers to execute in parallel.

#### Execution
```go
result, err := first.HandleThread(inputContext, nil)
```
- `inputContext`: The initial conversation thread or data to process.
- `nil`: The next handler in the chain (optional).

#### Error Handling
```go
if err != nil {
    log.Fatalf("Validation failed: %v", err)
}
```
- If all handlers fail, the error contains details about each failure.

---

### Why Use the `First` Handler?

1. **Efficiency**: By running handlers in parallel, you reduce the overall processing time.
2. **Flexibility**: You can define multiple strategies for handling a task and let the system choose the best one.
3. **Robustness**: If one strategy fails, others can still succeed, improving reliability.
4. **Error Insights**: Aggregated errors provide a comprehensive view of what went wrong.

---

### Real-World Scenarios

1. **Validation**: Use multiple validation methods and accept the first valid result.
2. **Fallback Mechanisms**: Try multiple ways to generate a response (e.g., LLM, rule-based, cached responses).
3. **Redundant Systems**: Attempt to fetch data from multiple sources and use the first successful response.
4. **Competing Algorithms**: Run different algorithms in parallel and use the first valid output.

---

### Example: Fallback Response Generation

```go
flow := NewThreadFlow("conversation")
flow.Use(NewLogging("audit"))

flow.Handle(handlers.First("response-generation",
    generateLLMResponse,
    generateRuleBasedResponse,
    generateCachedResponse,
))

result, err := flow.HandleThread(inputContext, nil)
if err != nil {
    log.Fatalf("Failed to generate response: %v", err)
}
fmt.Println("Generated response:", result)
```

In this example:
- The system first tries to generate a response using an LLM.
- If that fails, it falls back to a rule-based response.
- If that also fails, it uses a cached response.
- The first successful response is returned.

---

### Summary

The `First` handler is a powerful tool for building resilient and efficient systems. By running multiple strategies in parallel and using the first successful result, you can improve both performance and reliability. Whether you're validating input, generating responses, or fetching data, the `First` handler ensures that your system can adapt and succeed even when individual components fail.


# ThreadFlow

A **ThreadFlow** is similar to an HTTP router like Chi or Flow, but instead of routing HTTP requests, it manages the flow of conversation threads through different processing stages. Just as an HTTP router directs requests through middleware and handlers, ThreadFlow guides conversation threads through a series of processing steps.

Consider this HTTP router example:

```go
router := chi.NewRouter()
router.Use(middleware.Logger)        // Global middleware
router.Use(middleware.Recoverer)     // Another global middleware

router.Group(func(r chi.Router) {
    r.Use(middleware.Timeout(5 * time.Second))  // Group-specific middleware
    r.Get("/api/users", handleUsers)           // Route handler
    r.Get("/api/posts", handlePosts)           // Another handler
})
```

ThreadFlow follows a similar pattern but for conversation processing:

```go
flow := NewThreadFlow("conversation")
flow.Use(NewLogging("audit"))          // Global middleware logs all operations

flow.Handle(validateInput)             // Base handler for input validation

flow.Group(func(f *ThreadFlow) {
    f.Use(NewTimeout("timeout", 5))    // Group-specific timeout
    f.Use(NewRetry("retry", 3))        // Group-specific retry logic
    f.Handle(generateResponse)         // Handler for LLM response
    f.Handle(validateOutput)           // Handler for output validation
})
```

The key difference is that while HTTP routers process web requests, ThreadFlow processes conversation threads. Each handler in ThreadFlow can modify the conversation state, add messages, or update metadata, while middleware can add cross-cutting concerns like logging, timeouts, or retries.

Just as HTTP middleware can inspect or modify a request before it reaches the handler, ThreadFlow middleware can examine or transform the conversation thread before it reaches each processing stage. This makes it easy to add capabilities like:

1. **Logging**: Track each stage of conversation processing.
2. **Timeouts**: Set limits for how long LLM operations can take.
3. **Retries**: Automatically retry failed operations.
4. **Validation**: Ensure conversation content meets specific criteria.
5. **State Management**: Maintain and update conversation context.

The ThreadFlow pattern is particularly valuable when building complex conversational applications that require multiple processing steps, each with different requirements for reliability, timing, and validation.

This architecture keeps your conversation processing logic organized and maintainable, just as a well-structured HTTP router keeps your web application organized. The ability to group handlers and apply specific middleware to those groups gives you fine-grained control over how different parts of your conversation processing pipeline behave.
