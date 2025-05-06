package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/chriscow/minds/providers/openai"
)

// Define a struct for structured data extraction
type CustomerInfo struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Email   string `json:"email"`
	Problem string `json:"problem"`
}

func main() {
	// Create a context
	ctx := context.Background()

	// Create an OpenAI provider
	llm, err := openai.NewProvider()
	if err != nil {
		fmt.Printf("Error creating OpenAI provider: %v\n", err)
		os.Exit(1)
	}
	defer llm.Close()

	// Create a thread context with a customer support conversation
	tc := minds.NewThreadContext(ctx)
	tc.AppendMessages(
		minds.Message{Role: minds.RoleUser, Content: "Hi, I'm John Smith. I've been having trouble with my account."},
		minds.Message{Role: minds.RoleAssistant, Content: "Hello John, I'm sorry to hear that. Could you provide more details about the issue?"},
		minds.Message{Role: minds.RoleUser, Content: "I'm 35 years old and I've been trying to log in but it says my password is incorrect. My email is john.smith@example.com."},
		minds.Message{Role: minds.RoleAssistant, Content: "I understand. Let me check your account."},
	)

	// Example 1: Use FreeformExtractor to extract key-value pairs
	fmt.Println("Example 1: Freeform Extraction")
	fmt.Println("-------------------------------")

	freeformPrompt := `Extract key information from this customer support conversation. 
Look for the customer's name, age, email, and the problem they're experiencing. 
Format the response as an array of key-value pairs, where each pair includes a "key" field and a "value" field.
For example: [{"key": "name", "value": "John Smith"}, {"key": "age", "value": "35"}]`

	freeformExtractor := handlers.NewFreeformExtractor("customer-info", llm, freeformPrompt)

	resultTc, err := freeformExtractor.HandleThread(tc, nil)
	if err != nil {
		fmt.Printf("Error processing with FreeformExtractor: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Extracted Metadata:")
	for key, value := range resultTc.Metadata() {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// Example 2: Use StructuredExtractor with a schema
	fmt.Println("\nExample 2: Structured Extraction")
	fmt.Println("--------------------------------")

	structuredPrompt := `Extract customer information from this support conversation.
Identify the customer's name, age, email address, and the problem they're experiencing.
Format the response according to the provided schema.`

	schema, err := minds.NewResponseSchema("customer_info", "Customer information", CustomerInfo{})
	if err != nil {
		fmt.Printf("Error creating schema: %v\n", err)
		os.Exit(1)
	}

	structuredExtractor := handlers.NewStructuredExtractor("customer-details", llm, structuredPrompt, *schema)

	resultTc2, err := structuredExtractor.HandleThread(tc, nil)
	if err != nil {
		fmt.Printf("Error processing with StructuredExtractor: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Extracted Structured Data:")
	customerInfo := resultTc2.Metadata()["customer_info"]
	fmt.Printf("  %v\n", customerInfo)

	// Example 3: Chaining extractors with middleware
	fmt.Println("\nExample 3: Chaining Extractors")
	fmt.Println("------------------------------")

	// Create middleware for logging
	loggingMiddleware := minds.MiddlewareFunc(func(next minds.ThreadHandler) minds.ThreadHandler {
		return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			fmt.Println("  Processing conversation...")
			result, err := next.HandleThread(tc, nil)
			if err == nil {
				fmt.Println("  Extraction complete!")
			}
			return result, err
		})
	})

	// Apply middleware to structured extractor
	chainedExtractor := structuredExtractor.With(loggingMiddleware)

	resultTc3, err := chainedExtractor.HandleThread(tc, nil)
	if err != nil {
		fmt.Printf("Error in chained extraction: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Chained Extraction Result:")
	customerInfo = resultTc3.Metadata()["customer_info"]
	fmt.Printf("  %v\n", customerInfo)
}
