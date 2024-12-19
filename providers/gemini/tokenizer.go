package gemini

import (
	"context"

	"github.com/google/generative-ai-go/genai"
)

// Implement the TokenCounter interface for the Gemini provider.
func (p *Provider) CountTokens(text string) (int, error) {
	ctx := context.Background()
	model, err := p.getModel()
	if err != nil {
		return 0, err
	}
	tokResp, err := model.CountTokens(ctx, genai.Text(text))
	if err != nil {
		return 0, nil
	}

	return int(tokResp.TotalTokens), nil
}
