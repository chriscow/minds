package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/chriscow/minds/providers/gemini"
	"github.com/chriscow/minds/providers/openai"
	"github.com/fatih/color"
	"golang.org/x/time/rate"
)

// The example demonstrates how to create a rate limiter middleware for a
// joke-telling competition between two language models (LLMs). The rate limiter
// allows one joke every 5 seconds. The example uses the Gemini and OpenAI LLMs
// to exchange jokes in a ping-pong fashion. The rate limiter ensures that the
// joke exchange is doesn't exceed the rate limit.
//
// You could easily add a `handlers.Must` handler to ensure that the jokes are
// family-friendly and clean. The `handlers.Must` handler would cancel the
// joke exchange if any joke is inappropriate.

// RateLimiter provides rate limiting for thread handlers
type RateLimiter struct {
	limiter *rate.Limiter
	name    string
}

// NewRateLimiter creates a rate limiter that allows 'n' requests per duration
func NewRateLimiter(name string, n int, d time.Duration) *RateLimiter {
	return &RateLimiter{
		name:    name,
		limiter: rate.NewLimiter(rate.Every(d/time.Duration(n)), n),
	}
}

// HandleThread implements the ThreadHandler interface
func (r *RateLimiter) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	// Wait for rate limiter
	if err := r.limiter.Wait(tc.Context()); err != nil {
		return tc, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Pass thread to next handler if we haven't exceeded the rate limit
	if next != nil {
		return next.HandleThread(tc, nil)
	}

	return tc, nil
}

func main() {
	ctx := context.Background()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	sysPrompt := `Your name is %s. Prefix your responses with your name in this format: [%s]. If you hear a joke from the user, rate it 1 to 5. Then reply with a joke of your own.`

	geminiJoker, err := gemini.NewProvider(ctx, gemini.WithSystemPrompt(fmt.Sprintf(sysPrompt, cyan("Gemini"), cyan("Gemini"))))
	if err != nil {
		log.Fatalf("Error creating Gemini provider: %v", err)
	}

	openAIJoker, err := openai.NewProvider(openai.WithSystemPrompt(fmt.Sprintf(sysPrompt, green("OpenAI"), green("OpenAI"))))
	if err != nil {
		log.Fatalf("Error creating OpenAI provider: %v", err)
	}

	// Create a rate limiter that allows 1 request every 5 seconds
	limiter := NewRateLimiter("rate_limiter", 1, 5*time.Second)

	printJoke := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		fmt.Printf("%s\n", tc.Messages().Last().Content)
		return tc, nil
	})

	round := handlers.Sequential("joke_round", geminiJoker, printJoke, openAIJoker, printJoke)

	// Create a cycle that alternates between both LLMs, each followed by printing the joke
	jokeCompetition := handlers.For("joke_exchange", 5, round, nil)
	jokeCompetition.Use(limiter)

	// Initial prompt
	initialThread := minds.NewThreadContext(ctx).WithMessages(minds.Messages{
		{
			Role:    minds.RoleUser,
			Content: "You are in a joke telling contest. You go first.",
		},
	})

	// Let them exchange jokes until context is canceled
	if _, err := jokeCompetition.HandleThread(initialThread, nil); err != nil {
		log.Fatalf("Error in joke exchange: %v", err)
	}
}
