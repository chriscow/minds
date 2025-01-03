package openai

import (
	"github.com/chriscow/minds"
)

// HandleMessage implements the ThreadHandler interface for the OpenAI provider.
func (p *Provider) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {

	req := minds.Request{
		Messages: tc.Messages(),
	}

	resp, err := p.GenerateContent(tc.Context(), req)
	if err != nil {
		return tc, err
	}

	messages := tc.Messages()
	messages = append(messages, minds.Message{
		Role:    minds.RoleAssistant,
		Content: resp.String(),
	})

	if next != nil {
		return next.HandleThread(tc.WithMessages(messages), nil)
	}

	return tc.WithMessages(messages), nil
}
