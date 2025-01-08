package main

import (
	"context"
	"fmt"
	"log"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/chriscow/minds/providers/openai"
	"github.com/chriscow/minds/tools/calculator"
	"github.com/chriscow/minds/tools/extensions"
)

func main() {
	llm, _ := openai.NewProvider()

	// Create specialized handlers for different tasks
	calc, _ := calculator.NewCalculator(calculator.Lua)
	calcHandler := calc.(minds.ThreadHandler)

	questionHandler := llm // let the llm answer questions
	summaryHandler := handlers.Summarize(llm, "")
	defaultHandler := llm

	// Define conditions and their handlers
	intentSwitch := handlers.Switch("intent-router",
		defaultHandler, // fallback handler
		handlers.SwitchCase{
			// Use LLM to check if message is a math question
			Condition: handlers.LLMCondition{
				Generator: llm,
				Prompt:    "Does this message contain a mathematical calculation?",
			},
			Handler: calcHandler,
		},
		handlers.SwitchCase{
			// Check metadata for specific routing
			Condition: handlers.MetadataEquals{
				Key:   "type",
				Value: "question",
			},
			Handler: questionHandler,
		},
		handlers.SwitchCase{
			// Use Lua for complex condition
			Condition: extensions.LuaCondition{
				Script: `
                    -- Check if message is long and needs summarization
                    return string.len(last_message) > 500
                `,
			},
			Handler: summaryHandler,
		},
	)

	// Initial thread with metadata
	thread := minds.NewThreadContext(context.Background()).
		WithMessages(minds.Message{Role: minds.RoleUser, Content: "What is 7 * 12 + 5?"}).
		WithMetadata(map[string]interface{}{
			"type": "calculation",
		})

	// Process the thread
	result, err := intentSwitch.HandleThread(thread, nil)
	if err != nil {
		log.Fatalf("Error processing thread: %v", err)
	}

	fmt.Println("Response:", result.Messages().Last().Content)
}
