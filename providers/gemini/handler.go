package gemini

import (
	"fmt"

	"github.com/chriscow/minds"
)

// HandleMessage implements the ThreadHandler interface for the Gemini provider.
func (p *Provider) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	req := minds.Request{
		Messages: tc.Messages(),
	}

	resp, err := p.GenerateContent(tc.Context(), req)
	if err != nil {
		return tc, err
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
