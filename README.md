
# Minds Toolkit

A lightweight Go library for building LLM-based applications through the
composition of handlers, inspired by the `http.Handler` middleware pattern.

This toolkit takes inspiration from LangChain's runnables and addresses the need
for a modular, extensible framework for conversational AI development in Go. By
leveraging Go's idiomatic patterns, the library provides a composable middleware
design tailored to processing message threads.

The framework applies the same handler-based design to both LLMs and tool
integrations. It includes implementations for OpenAI and Google's Gemini, as
well as a suite of tools in the `minds/openai`, `minds/gemini`, and
`minds/tools` modules.

## Features

- **Composable Middleware**: Build complex pipelines for handling message threads with composable, reusable handlers.
- **Extensible Design**: Add custom handlers and integrate external APIs with ease.
- **Integration Tools**: Built-in support for LLM providers and tools for generative AI workflows.
- **Testing-Friendly**: Well-structured interfaces and unit-tested components for robust development.

## Installation

Use `go get` to install the library:

```bash
go get github.com/chriscow/minds
```

## Usage

### Basic Example

Here’s how you can compose handlers for processing a thread:

```go
func main() {
	ctx := context.Background()
	geminiJoker, _ := gemini.NewProvider(ctx)
	openAIJoker, _ := openai.NewProvider()

	// Create a rate limiter that allows 1 request every 5 seconds
	limiter := NewRateLimiter("rate_limiter", 1, 5*time.Second)

	printJoke := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		fmt.Printf("Joke: %s\n", tc.Messages().Last().Content)
		return tc, nil
	})

	// Create a cycle that alternates between both LLMs, each followed by printing the joke
	jokeExchange := handlers.Cycle("joke_exchange", 5,
		geminiJoker,
		printJoke,
		openAIJoker,
		printJoke,
	)
	jokeExchange.Use(limiter)

	// Initial prompt
	prompt := "Tell me a clean, family-friendly joke. Keep it clean and make me laugh!"
	initialThread := minds.NewThreadContext(ctx).WithMessages(minds.Messages{
		{Role: minds.RoleUser, Content: prompt},
	})

	// Let them exchange jokes until context is canceled
	if _, err := jokeExchange.HandleThread(initialThread, nil); err != nil {
		log.Fatalf("Error in joke exchange: %v", err)
	}
}
```

### Adding a Calculator Tool

The library supports Lua and Starlark for mathematical operations. Here's how to integrate a calculator:

```go
package main

import (
    "context"
    "github.com/chriscow/minds"
    "github.com/chriscow/minds/calculator"
)

func main() {
    calc, _ := calculator.NewCalculator(calculator.Starlark)
    result, _ := calc.Call(context.Background(), []byte("2 + 3"))
    println(string(result)) // Outputs: 5
}
```

## Documentation

Refer to the `_examples` provided for guidance on how to use the modules.

## Contributing

Contributions are welcome! Please see the [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgements

This project is inspired by the `http.Handler` middleware pattern and the need for modular and extensible LLM application development in Go.

