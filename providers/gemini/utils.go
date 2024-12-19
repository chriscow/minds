package gemini

import (
	"context"
	"fmt"

	"github.com/chriscow/minds"
)

func Ask(ctx context.Context, prompt string, opts ...Option) (string, error) {
	llm, err := NewProvider(ctx, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to create LLM provider: %v", err)
	}

	resp, err := llm.GenerateContent(ctx, minds.Request{
		Messages: minds.Messages{{Content: prompt}}})

	return resp.String(), err
}
