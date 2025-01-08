package openai

import (
	"errors"
	"fmt"

	"github.com/chriscow/minds"

	"github.com/sashabaranov/go-openai"
)

type Response struct {
	raw   openai.ChatCompletionResponse
	calls []minds.ToolCall
}

func NewResponse(resp openai.ChatCompletionResponse, calls []minds.ToolCall) (*Response, error) {
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	if len(resp.Choices) > 1 {
		return nil, fmt.Errorf("multiple choices in OpenAI response not supported")
	}

	if resp.Choices[0].FinishReason != openai.FinishReasonStop && resp.Choices[0].FinishReason != openai.FinishReasonToolCalls {
		return nil, errors.New(string(resp.Choices[0].FinishReason))
	}

	if calls == nil {
		calls = make([]minds.ToolCall, 0)
	}

	return &Response{raw: resp, calls: calls}, nil
}

func (r Response) Messages() (minds.Messages, error) {
	messages := minds.Messages{}

	if len(r.raw.Choices) == 0 {
		return messages, nil
	}

	for _, call := range r.ToolCalls() {
		messages = append(messages, minds.Message{
			Role:       minds.RoleTool,
			Content:    string(call.Function.Result),
			ToolCallID: call.ID,
		})
	}

	if r.raw.Choices[0].Message.Content != "" {
		messages = append(messages, minds.Message{
			Role:    minds.RoleAssistant,
			Content: r.raw.Choices[0].Message.Content,
		})
	}

	return messages, nil
}

func (r Response) String() string {
	if len(r.raw.Choices) == 0 {
		return "No response from OpenAI"
	}

	return r.raw.Choices[0].Message.Content
}

func (r Response) ToolCalls() []minds.ToolCall {
	return r.calls
}
