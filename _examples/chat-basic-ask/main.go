package main

import (
	"context"
	"fmt"

	"github.com/chriscow/minds/providers/deepseek"
	"github.com/chriscow/minds/providers/gemini"
	"github.com/chriscow/minds/providers/openai"
	"github.com/fatih/color"
)

var (
	cyan   = color.New(color.FgCyan).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	purple = color.New(color.FgHiMagenta).SprintFunc()
)

// Sometimes you just need to ask a quick question without a lot of boilerplate.
// This example demonstrates how to use the Gemini provider to ask a simple question
// using the default model (gemini-1.5-flash).
//
// Ask always uses the fastest and usually cheapest model available, which is
// gemini-1.5-flash in this case.
//
// You need to have the provider's API key set in the environment.
func main() {
	ctx := context.Background()
	prompt := "Knock knock"

	askGemini(ctx, prompt)
	askOpenAI(ctx, prompt)
	askDeepSeek(ctx, prompt)

}

func askGemini(ctx context.Context, prompt string) {
	answer, err := gemini.Ask(ctx, prompt)
	if err != nil {
		answer = "error: " + err.Error()
	}
	fmt.Printf("%s: %s", cyan("Gemini"), answer)
}

func askOpenAI(ctx context.Context, prompt string) {
	answer, err := openai.Ask(ctx, prompt)
	if err != nil {
		answer = "error: " + err.Error()
	}
	fmt.Printf("%s: %s\n", green("OpenAI"), answer)
}

func askDeepSeek(ctx context.Context, prompt string) {
	answer, err := deepseek.Ask(ctx, prompt)
	if err != nil {
		answer = "error: " + err.Error()
	}
	fmt.Printf("%s: %s\n", purple("DeepSeek"), answer)
}
