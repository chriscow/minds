package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/gemini"
	"github.com/chriscow/minds/handlers"
	"github.com/chriscow/minds/openai"
)

const prompt = "What is the meaning of life?"

// This example demonstrates how to compose multiple handlers into a single pipeline
// using the familiar "middleware" pattern of Go's `net/http` package.
func main() {
	if os.Getenv("GEMINI_API_KEY") == "" {
		log.Fatalf("GEMINI_API_KEY is not set")
	}

	ctx := context.Background()

	llm := getGeminiProvider(ctx)
	runPipeline(ctx, llm)

	llm = getOpenAIProvider()
	runPipeline(ctx, llm)
}

func runPipeline(ctx context.Context, llm minds.ThreadHandler) {
	// Compose the handlers into a single pipeline.
	// The pipeline is an ordered list of handlers that each process the thread in some way.
	// The final handler in the pipeline is responsible for handling the final result.
	pipeline := handlers.Sequential("pipeline", exampleHandler(), llm)
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

func getGeminiProvider(ctx context.Context) minds.ThreadHandler {
	// These provider implementations also implement the `ThreadHandler` interface
	llm, err := gemini.NewProvider(ctx)
	if err != nil {
		log.Fatalf("failed to create LLM provider: %v", err)
	}

	return llm
}

func getOpenAIProvider() minds.ThreadHandler {
	// These provider implementations also implement the `ThreadHandler` interface
	llm, err := openai.NewProvider()
	if err != nil {
		log.Fatalf("failed to create LLM provider: %v", err)
	}

	return llm
}
