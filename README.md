
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
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/providers/gemini"
	"github.com/chriscow/minds/handlers"
	"github.com/chriscow/minds/providers/openai"
)

const prompt = "What is the meaning of life?"

// This example demonstrates how to compose multiple handlers into a single pipeline
// using the familiar "middleware" pattern of Go's `net/http` package.
func main() {
	if os.Getenv("GEMINI_API_KEY") == "" {
		log.Fatalf("GEMINI_API_KEY is not set")
	}

	ctx := context.Background()

    // ContentGenerator providers implement the ThreadHandler interface
	llm, err := gemini.NewProvider(ctx)
	if err != nil {
		log.Fatalf("failed to create LLM provider: %v", err)
	}
	runPipeline(ctx, llm)

    // Try it with OpenAI ...
	llm, err := openai.NewProvider()
	if err != nil {
		log.Fatalf("failed to create LLM provider: %v", err)
	}
	runPipeline(ctx, llm)
}

func runPipeline(ctx context.Context, llm minds.ThreadHandler) {
	// Compose the handlers into a single pipeline.
	// The pipeline is an ordered list of handlers that each process the thread in some way.
	// The final handler in the pipeline is responsible for handling the final result.
	pipeline := handlers.Sequential("pipeline", exampleHandler(), llm) // Add more handlers ...
	pipeline.Use(validateMiddlware())

	// Initial message thread to start things off
	initialThread := minds.NewThreadContext(ctx).WithMessages(minds.Messages{
		{Role: minds.RoleUser, Content: prompt},
	})

	// Final handler (end of the pipeline)
	finalHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		fmt.Println("[finalHandler]: \n\n" + tc.Messages().Last().Content)
		return tc, nil
	})

	// Execute the pipeline
	if _, err := pipeline.HandleThread(initialThread, finalHandler); err != nil {
		log.Fatalf("Handler chain failed: %v", err)
	}
}

func exampleHandler() minds.ThreadHandlerFunc {
	return func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		fmt.Println("[exampleHandler]")

		//
		// Pass the tread to the next handler in the chain
		//
		if next != nil {
			return next.HandleThread(tc, nil)
		}

		return tc, nil
	}
}

// Middleware are ThreadHandlers like any other, but they are used to wrap other handlers
// to provide additional functionality, such as validation, logging, or error handling.
func validateMiddlware() minds.ThreadHandlerFunc {
	return func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		if len(tc.Messages()) == 0 {
			return tc, fmt.Errorf("thread has no messages")
		}

		// TODO: validate the input thread before processing
		fmt.Println("[validator in] Validated thread before processing")

		if next != nil {
			var err error
			tc, err = next.HandleThread(tc, nil)
			if err != nil {
				return tc, err
			}
		}

		// TODO: validate the output thread after processing
		fmt.Println("[validator out] Validated thread after processing")

		return tc, nil
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

Refer to the [documentation](https://github.com/chriscow/minds/docs) for detailed guides and API references.

## Contributing

Contributions are welcome! Please see the [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgements

This project is inspired by the `http.Handler` middleware pattern and the need for modular and extensible LLM application development in Go.

