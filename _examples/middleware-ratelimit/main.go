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
	var err error
	ctx := context.Background()
	llm1, err := gemini.NewProvider(ctx)
	if err != nil {
		log.Fatalf("Error creating Gemini provider: %v", err)
	}

	llm2, err := openai.NewProvider()
	if err != nil {
		log.Fatalf("Error creating OpenAI provider: %v", err)
	}

	// Create a rate limiter that allows 1 request every 5 seconds
	limiter := NewRateLimiter("rate_limiter", 1, 5*time.Second)

	// Create handlers for each LLM
	geminiJoker := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		// Add prompt for joke response
		messages := tc.Messages()
		messages = append(messages, &minds.Message{
			Role:    minds.RoleUser,
			Content: "Respond to the previous joke with a funnier joke. Do not give away the answer unless they give up. Keep it clean and family-friendly.",
		})

		// Process with Gemini
		tc, err := llm1.HandleThread(tc.WithMessages(messages), nil)
		if err != nil {
			return tc, err
		}

		// Print Gemini's joke
		fmt.Printf("Gemini: %s\n", tc.Messages().Last().Content)

		// Pass to next handler
		if next != nil {
			return next.HandleThread(tc, nil)
		}
		return tc, nil
	})

	openAIJoker := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
		// Add prompt for joke response
		messages := tc.Messages()
		messages = append(messages, &minds.Message{
			Role:    minds.RoleUser,
			Content: "That's a good one! Now respond with an even funnier joke. Do not give away the answer unless they give up. Keep it clean and family-friendly.",
		})

		// Process with OpenAI
		tc, err := llm2.HandleThread(tc.WithMessages(messages), nil)
		if err != nil {
			return tc, err
		}

		// Print OpenAI's joke
		fmt.Printf("OpenAI: %s\n", tc.Messages().Last().Content)

		// Pass to next handler (which will be Gemini again)
		if next != nil {
			return next.HandleThread(tc, nil)
		}
		return tc, nil
	})

	circular := handlers.Sequential("ping_pong", geminiJoker, openAIJoker)
	circular.Use(limiter)

	// Start with an initial joke to get the ball rolling
	initialThread := minds.NewThreadContext(ctx).WithMessages(minds.Messages{
		{Role: minds.RoleUser, Content: "Tell me a clean, family-friendly joke to start our joke-telling competition."},
	})

	// Run for 5 rounds (10 jokes total)
	for i := 0; i < 5; i++ {
		fmt.Printf("\nRound %d:\n", i+1)
		initialThread, err = circular.HandleThread(initialThread, nil)
		if err != nil {
			log.Fatalf("Error in joke exchange: %v", err)
		}
	}
}
