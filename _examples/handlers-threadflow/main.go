package main

import (
	"context"
	"fmt"
	"log"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/chriscow/minds/providers/openai"
)

// ThreadFlow manages multiple handlers and their middleware chains. Handlers are executed
// sequentially, with each handler receiving the result from the previous one. Middleware
// is scoped and applied to handlers added within that scope, allowing different middleware
// combinations for different processing paths.

// Example:

// 	validate := NewValidationHandler()
// 	process := NewProcessingHandler()
// 	seq := Sequential("main", validate, process)

// 	// Create flow and add global middleware
// 	flow := NewThreadFlow("example")
// 	flow.Use(NewLogging("audit"))

// 	// Add base handlers
// 	flow.Handle(seq)

// 	// Add handlers with scoped middleware
// 	flow.Group(func(f *ThreadFlow) {
// 	    f.Use(NewRetry("retry", 3))
// 	    f.Use(NewTimeout("timeout", 5))
// 	    f.Handle(NewContentProcessor("content"))
// 	    f.Handle(NewValidator("validate"))
// 	})

// result, err := flow.HandleThread(initialContext, nil)

// 	if err != nil {
// 	    log.Fatalf("Error in flow: %v", err)
// 	}

// fmt.Println("Result:", result.Messages().Last().Content)

func main() {
	// This example demonstrates how to use a ThreadFlow. ThreadFlows are a `top-level`
	// construct that manage middleware and handlers. ThreadFlow are used to define the
	// processing flow for a conversation. In this example, a ThreadFlow is used to
	// manage a code review conversation. The conversation is initiated by a user
	// asking for a code review. The code review assistant responds with a snarky
	// comment about the user's choice of indentation style. The assistant then uses
	// a handler to validate the user's indentation style. The handler uses a provider
	// to generate a response. The response is then returned to the user.

	// Although the LLM isn't actually used in this example, it's included to show how
	// a provider can be used in a handler.

	ctx := context.Background()
	llm, err := openai.NewProvider()
	if err != nil {
		log.Fatal(err)
	}

	flow := handlers.NewThreadFlow("code_review")
	flow.Use(NewSarcasmMiddleware(9001))

	flow.Group(func(f *handlers.ThreadFlow) {
		f.Handle(NewTabsVsSpacesValidator(llm))
	})

	tc := minds.NewThreadContext(ctx).WithMessages(minds.Message{
		Role:    minds.RoleUser,
		Content: "Please review my Python code that uses 4 spaces for indentation",
	})

	result, err := flow.HandleThread(tc, nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, msg := range result.Messages() {
		fmt.Println(msg.Content)
	}
}

// SarcasmMiddleware adds snark to responses
type SarcasmMiddleware struct {
	level int
}

func NewSarcasmMiddleware(level int) handlers.Middleware {
	return &SarcasmMiddleware{level: level}
}

func (m *SarcasmMiddleware) Wrap(next minds.ThreadHandler) minds.ThreadHandler {
	return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
		tc.SetKeyValue("snark_level", m.level)
		result, err := next.HandleThread(tc, nil)
		return result, err
	})
}

// TabsVsSpacesValidator checks indentation
type TabsVsSpacesValidator struct {
	llm minds.ContentGenerator
}

func NewTabsVsSpacesValidator(llm minds.ContentGenerator) minds.ThreadHandler {
	return &TabsVsSpacesValidator{llm: llm}
}

func (v *TabsVsSpacesValidator) HandleThread(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
	tc.AppendMessages(minds.Message{
		Role:    minds.RoleAssistant,
		Content: "Spaces?! *adjusts glasses disapprovingly* We use tabs in this house, young programmer.",
	})
	return tc, nil
}
