package openai

import (
	"fmt"

	"github.com/chriscow/minds"
)

// HandleMessage implements the ThreadHandler interface for the OpenAI provider.
func (p *Provider) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {

	req := minds.Request{
		Messages: tc.Messages(),
	}

	for i, m := range req.Messages {
		switch m.Role {
		case minds.RoleModel:
			req.Messages[i].Role = minds.RoleAssistant
		}
	}

	resp, err := p.GenerateContent(tc.Context(), req)
	if err != nil {
		return tc, fmt.Errorf("failed to generate content: %w", err)
	}

	messages, err := resp.(*Response).Messages()
	tc.AppendMessages(messages...)

	if next != nil {
		if err != nil {
			return tc, fmt.Errorf("failed to get messages from response: %w", err)
		}

		return next.HandleThread(tc, nil)
	}

	return tc, nil
}
