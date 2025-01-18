package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/chriscow/minds/providers/openai"
)

// This example demonstrates the `Sequential` handler.  The `Sequential` handler
// executes each handler in order. Here we compose multiple handlers into a
// single pipeline.
func main() {
	if os.Getenv("GEMINI_API_KEY") == "" {
		log.Fatalf("GEMINI_API_KEY is not set")
	}

	ctx := context.Background()

	managerPrompt := `You are a pointy-haired boss from Dilbert. You speak in management buzzwords
	and crazy ideas. Keep responses short and focused on maximizing synergy and
	disrupting paradigms. You have no technical understanding but pretend you do.
	Never break character. Prefix your response with [Boss:]`

	engineerPrompt := `You are Dilbert, a cynical software engineer. You respond to management 
	with dry wit and technical accuracy while pointing out logical flaws. Keep responses
	short and sardonic. Never break character. Prefix your response with [Dilbert:]`

	manager, err := openai.NewProvider(openai.WithSystemPrompt(managerPrompt))
	if err != nil {
		log.Fatal(err)
	}

	engineer, err := openai.NewProvider(openai.WithSystemPrompt(engineerPrompt))
	if err != nil {
		log.Fatal(err)
	}

	tc := minds.NewThreadContext(ctx).WithMessages(minds.Message{
		Role: minds.RoleUser, Content: "We need to leverage blockchain AI to disrupt our coffee machine's paradigm",
	})

	comic := handlers.Sequential("comic", manager, engineer, manager, engineer)
	result, err := comic.HandleThread(tc, comic)
	if err != nil {
		log.Fatal(err)
	}

	for _, message := range result.Messages() {
		fmt.Printf("%s\n", message.Content)
	}

	// Output should look something like:
	// Boss: We need to synergize our coffee consumption metrics with blockchain-enabled AI...
	// Dilbert: So... you want to put a computer chip in the coffee maker?
	// Boss: Exactly! And we'll call it CoffeeChain 3.0 - it's like Web3 but for caffeine...
	// Dilbert: *sigh* I'll just order a new Mr. Coffee from Amazon.
}
