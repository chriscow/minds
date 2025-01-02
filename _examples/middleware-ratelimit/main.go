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
	geminiJoker, err := gemini.NewProvider(ctx)
	if err != nil {
		log.Fatalf("Error creating Gemini provider: %v", err)
	}

	openAIJoker, err := openai.NewProvider()
	if err != nil {
		log.Fatalf("Error creating OpenAI provider: %v", err)
	}

	// Create a rate limiter that allows 1 request every 5 seconds
	limiter := NewRateLimiter("rate_limiter", 1, 5*time.Second)

	printJoke := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		fmt.Printf("Joke: %s\n", tc.Messages().Last().Content)
		return tc, nil
	})

	// Create a cycle that alternates between both LLMs, each followed by printing the joke
	jokeExchange := handlers.Cycle("joke_exchange", 5,
		geminiJoker,
		printJoke,
		openAIJoker,
		printJoke,
	)
	jokeExchange.Use(limiter)

	// Initial prompt
	initialThread := minds.NewThreadContext(ctx).WithMessages(minds.Messages{
		{
			Role:    minds.RoleUser,
			Content: "Tell me a clean, family-friendly joke. Keep it clean and make me laugh!",
		},
	})

	// Let them exchange jokes until context is canceled
	if _, err := jokeExchange.HandleThread(initialThread, nil); err != nil {
		log.Fatalf("Error in joke exchange: %v", err)
	}
}
