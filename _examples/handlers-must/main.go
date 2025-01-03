package main

import (
	"context"
	"fmt"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/chriscow/minds/providers/gemini"
	"github.com/fatih/color"
)

// The example demonstrates the must handler. The `Must` handler executes all
// handlers in parallel. It ensures that all handlers succeed. If any handler
// fails, the others are canceled, and the first error is returned.
//
// This is useful for policy enforcement where multiple validators must pass
// before the next handler is executed.
func main() {
	red := color.New(color.FgRed).SprintFunc()

	ctx := context.Background()
	llm, err := gemini.NewProvider(ctx)
	if err != nil {
		fmt.Printf("%s: %v", red("error"), err)
	}

	// This sets up a pipeline that validates messages for dad jokes, coffee
	// obsession, and unnecessary jargon in parallel using the `must` handler.
	// If any of the validators fail, the others are canceled.
	validationPipeline := validationPipeline(llm)

	//
	// Some example message threads to test the validation pipeline
	//
	jargon := minds.NewThreadContext(context.Background()).WithMessages(minds.Messages{
		{
			Role: minds.RoleUser,
			Content: `Leveraging our synergistic capabilities, we aim to 
				proactively optimize cross-functional alignment and drive 
				scalable value-add solutions for our stakeholders. By 
				implementing a paradigm-shifting approach to our core 
				competencies, we can seamlessly catalyze transformative 
				outcomes. This ensures a robust framework for sustained 
				competitive differentiation in a dynamic market landscape.`,
		},
	})

	dad := minds.NewThreadContext(context.Background()).WithMessages(minds.Messages{
		{
			Role:    minds.RoleUser,
			Content: "Hi hungry, I'm dad",
		},
	})

	coffee := minds.NewThreadContext(context.Background()).WithMessages(minds.Messages{
		{
			Role: minds.RoleUser,
			Content: "Why didn't the coffee file a police report? Because it got mugged! " +
				"Speaking of which, time for cup number 6!",
		},
	})

	// Final handler (end of the pipeline)
	finalHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		fmt.Println("[finalHandler]: " + tc.Messages().Last().Content)
		return tc, nil
	})

	// Test the validation pipeline for jargon
	if _, err := validationPipeline.HandleThread(jargon, finalHandler); err != nil {
		fmt.Printf("%s: %v\n", red("Validation failed"), err)
	}

	// Test the validation pipeline for dad jokes
	if _, err := validationPipeline.HandleThread(dad, finalHandler); err != nil {
		fmt.Printf("%s: %v\n", red("Validation failed"), err)
	}

	// Test the validation pipeline for coffee jokes
	if _, err := validationPipeline.HandleThread(coffee, finalHandler); err != nil {
		fmt.Printf("%s: %v\n", red("Validation failed"), err)
	}
}

func validationPipeline(llm minds.ContentGenerator) minds.ThreadHandler {
	// Create policy validators with humorous but detectable rules
	validators := []minds.ThreadHandler{
		handlers.Policy(
			llm,
			"detects_dad_jokes",
			`Monitor conversation for classic dad joke patterns like:
			- "Hi hungry, I'm dad"
			- Puns that make people groan
			- Questions with obvious punchlines
			Flag if more than 2 dad jokes appear in a 5-message window.
			Explain why they are definitely dad jokes.`,
			nil,
		),
		handlers.Policy(
			llm,
			"detects_coffee_obsession",
			`Analyze messages for signs of extreme coffee dependence:
			- Mentions of drinking > 5 cups per day
			- Using coffee-based time measurements
			- Personifying coffee machines
			- Expressing emotional attachment to coffee
			Provide concerned feedback about caffeine intake.`,
			nil,
		),
		handlers.Policy(
			llm,
			"detects_unnecessary_jargon",
			`Monitor for excessive business speak like:
			- "Leverage synergies"
			- "Circle back"
			- "Touch base"
			- "Move the needle"
			- Using "utilize" instead of "use"
			Suggest simpler alternatives in a disappointed tone.`,
			nil,
		),
	}

	return handlers.Must("validators-must-succeed", validators...)
}
