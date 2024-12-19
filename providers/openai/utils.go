package openai

import (
	"context"
	"fmt"

	"github.com/chriscow/minds"
)

func Ask(ctx context.Context, question string, opts ...Option) (string, error) {
	llm, err := NewProvider(opts...)
	if err != nil {
		return "", fmt.Errorf("new provider: %w", err)
	}

	resp, _ := llm.GenerateContent(ctx, minds.Request{
		Messages: minds.Messages{{Content: question}}})

	return resp.String(), nil

}
