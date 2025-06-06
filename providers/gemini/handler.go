package gemini

import (
	"github.com/chriscow/minds"
)

/*

// HandleMessage implements the ThreadHandler interface for the OpenAI provider.
func (p *Provider) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {

	messages := minds.Messages{}
	if p.options.systemPrompt != nil {
		messages = append(messages, minds.Message{
			Role:    minds.RoleSystem,
			Content: *p.options.systemPrompt,
		})
	}

	messages = append(messages, tc.Messages()...)

	req := minds.Request{
		Messages: messages,
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
	// fmt.Printf("[%s] %s\n", p.options.name, resp.String())

	msg := minds.Message{
		Role:    minds.RoleAssistant,
		Name:    p.options.name,
		Content: resp.String(),
	}

	tc.AppendMessages(msg)

	if next != nil {
		return next.HandleThread(tc, nil)
	}

	return tc, nil
}

func (p *Provider) String() string {
	return fmt.Sprintf("OpenAI Provider: %s", p.options.name)
}

*/

// HandleMessage implements the ThreadHandler interface for the Gemini provider.
func (p *Provider) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {

	messages := minds.Messages{}
	if p.options.systemPrompt != nil {
		messages = append(messages, minds.Message{
			Role:    minds.RoleSystem,
			Content: *p.options.systemPrompt,
		})
	}

	messages = append(messages, tc.Messages()...)

	req := minds.Request{
		Messages: messages,
	}

	for i, m := range req.Messages {
		switch m.Role {
		case minds.RoleModel:
			req.Messages[i].Role = minds.RoleAssistant
		}
	}

	resp, err := p.GenerateContent(tc.Context(), req)
	if err != nil {
		return tc, err
	}

	msg := minds.Message{
		Role:    minds.RoleAssistant,
		Name:    p.options.name,
		Content: resp.String(),
	}

	tc.AppendMessages(msg)

	if next != nil {
		return next.HandleThread(tc, nil)
	}

	return tc, nil
}
