package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/chriscow/minds"
)

// summarize is a MessageHandler that takes the list of messages passed to it
// and prompts the LLM provider to summarize the conversation so far. It returns
// a single message with the system message appended with the current summary.
type summarize struct {
	provider  minds.ContentGenerator
	systemMsg string
	summary   string
}

// Summerize creates a handler that maintains a running summary of thread messages.
//
// The handler prompts an LLM to generate a concise summary of all messages in the thread,
// focusing on key information. The summary is appended to the system message in
// XML tags and persists across handler invocations.
//
// Parameters:
//   - provider: LLM content generator for creating summaries.
//   - systemMsg: Initial system message to prepend to summaries.
//
// Returns:
//   - A handler that maintains thread summaries via LLM generation.
//
// Note: The original thread context is not modified; a new context with the
// updated system message is created.
func Summerize(provider minds.ContentGenerator, systemMsg string) *summarize {
	return &summarize{
		provider:  provider,
		systemMsg: systemMsg,
	}
}

func (s *summarize) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	ctx := tc.Context()
	messages, err := json.Marshal(tc.Messages())
	if err != nil {
		return tc, err
	}

	prompt := fmt.Sprintf(`
Your task is to create a concise running summary of responses and information
in the provided text, focusing on key and potentially important information
to remember. You will receive the current summary and your latest responses.
Combine them, adding relevant key information from the latest development
in 1st person past tense and keeping the summary concise.

Current Summary:
""" %s """

Latest Responses:
""" %s """`, s.summary, string(messages))

	req := minds.Request{
		Messages: minds.Messages{
			{
				Role:    minds.RoleUser,
				Content: prompt,
			},
		},
	}

	resp, err := s.provider.GenerateContent(ctx, req)
	if err != nil {
		return tc, err
	}
	s.summary = resp.String()

	tc = tc.WithMessages(minds.Messages{
		{
			Role:    minds.RoleSystem,
			Content: fmt.Sprintf("%s\n\n<summary>%s</summary>", s.systemMsg, s.summary),
		},
	})

	if next != nil {
		return next.HandleThread(tc, next)
	}

	return tc, err
}
