package openai

import (
	"errors"
	"fmt"
	"strings"

	"github.com/chriscow/minds"

	"github.com/sashabaranov/go-openai"
)

type Response struct {
	raw openai.ChatCompletionResponse
}

func NewResponse(resp openai.ChatCompletionResponse) (*Response, error) {
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	if len(resp.Choices) > 1 {
		return nil, fmt.Errorf("multiple choices in OpenAI response")
	}

	if resp.Choices[0].FinishReason != openai.FinishReasonStop && resp.Choices[0].FinishReason != openai.FinishReasonToolCalls {
		return nil, errors.New(string(resp.Choices[0].FinishReason))
	}

	return &Response{raw: resp}, nil
}

func (r Response) ToMessages() (minds.Messages, error) {
	switch r.Type() {
	case minds.ResponseTypeText:
		return minds.Messages{
			{
				Role:    minds.Role(r.raw.Choices[0].Message.Role),
				Content: r.raw.Choices[0].Message.Content,
			},
		}, nil
	case minds.ResponseTypeToolCall:
		calls, _ := r.ToolCalls()
		resp, err := minds.NewToolCallResponse(calls)
		if err != nil {
			return nil, fmt.Errorf("failed to create tool call response: %w", err)
		}

		return resp.ToMessages()
	default:
		return nil, fmt.Errorf("unknown response type")
	}
}

func (r Response) String() string {
	responseType := r.Type()

	switch responseType {
	case minds.ResponseTypeText:
		text, _ := r.Text()
		return text
	case minds.ResponseTypeToolCall:
		tools, _ := r.ToolCalls()
		names := make([]string, 0, len(tools))
		for _, tool := range tools {
			names = append(names, tool.Function.Name)
		}

		return fmt.Sprintf(strings.Join(names, ", "))
	default:
		return "Unknown response type"
	}
}

func (r Response) Type() minds.ResponseType {
	if r.raw.Choices[0].FinishReason == openai.FinishReasonToolCalls {
		return minds.ResponseTypeToolCall
	}

	if r.raw.Choices[0].FinishReason == openai.FinishReasonStop || r.raw.Choices[0].FinishReason == openai.FinishReasonLength {
		return minds.ResponseTypeText
	}

	return minds.ResponseTypeUnknown
}

func (r Response) Text() (string, bool) {
	if r.Type() != minds.ResponseTypeText {
		return "", false
	}

	return r.raw.Choices[0].Message.Content, true
}

func (r Response) ToolCalls() ([]minds.ToolCall, bool) {
	if r.Type() != minds.ResponseTypeToolCall {
		return nil, false
	}

	infos := make([]minds.ToolCall, 0)
	for _, tool := range r.raw.Choices[0].Message.ToolCalls {
		name := tool.Function.Name

		infos = append(infos, minds.ToolCall{
			Function: minds.FunctionCall{
				Name:      name,
				Arguments: []byte(tool.Function.Arguments),
			},
		})
	}

	return infos, true
}
