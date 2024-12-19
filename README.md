
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
- **Scripting Support**: Evaluate mathematical expressions and scripts with support for Starlark and Lua.
- **Extensible Design**: Add custom handlers and integrate external APIs with ease.
- **Integration Tools**: Built-in support for LLM providers and services like SerpAPI for search and generative AI workflows.
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
	"github.com/chriscow/minds/gemini"
	"github.com/chriscow/minds/handlers"
)

func main() {
	if os.Getenv("GEMINI_API_KEY") == "" {
		log.Fatal("GEMINI_API_KEY is not set")
	}

	ctx := context.Background()
	client, err := gemini.NewClient(ctx, os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	llm, err := gemini.NewProvider(client)
	if err != nil {
		log.Fatalf("Failed to create Gemini provider: %v", err)
	}

	// Compose a simple handler pipeline
	pipeline := handlers.Sequential("pipeline", "simple pipeline",
		printThreadHandler(),
		llm,
		finalHandler(),
	)

	// Initial thread
	thread := minds.NewThreadContext(nil, minds.NewMessageThread(nil, []minds.Message{
		{Role: minds.RoleUser, Content: "What is the meaning of life?"},
	}))

	// Execute the pipeline
	if err := pipeline.HandleThread(ctx, thread, nil); err != nil {
		log.Fatalf("Pipeline failed: %v", err)
	}
}

func printThreadHandler() minds.ThreadHandlerFunc {
	return func(ctx context.Context, tc minds.ThreadContext, next minds.ThreadHandler) error {
		fmt.Printf("User said: %s\n", tc.Thread().LastMessage().Content)
		return next.HandleThread(ctx, tc, next)
	}
}

func finalHandler() minds.ThreadHandlerFunc {
	return func(ctx context.Context, tc minds.ThreadContext, _ minds.ThreadHandler) error {
		fmt.Printf("Assistant replied: %s\n", tc.Thread().LastMessage().Content)
		return nil
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

---

This draft provides an overview of the project, showcases its capabilities with examples, and offers practical installation and usage details. Let me know if you'd like adjustments!