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

	resp, err := llm.GenerateContent(ctx, minds.Request{
		Messages: minds.Messages{{Role: minds.RoleUser, Content: question}}})
	if err != nil {
		return "", fmt.Errorf("generate content: %w", err)
	}

	return resp.String(), nil

}
