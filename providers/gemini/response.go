package gemini

import (
	"fmt"

	"github.com/chriscow/minds"

	"github.com/google/generative-ai-go/genai"
)

// Response represents a unified response type that can handle both
// text responses and function calls from the Gemini API
type Response struct {
	raw   *genai.GenerateContentResponse
	calls []minds.ToolCall
}

// NewResponse creates a new Response from a Gemini API response
func NewResponse(resp *genai.GenerateContentResponse, calls []minds.ToolCall) (*Response, error) {
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in Gemini response")
	}

	if len(resp.Candidates) > 1 {
		return nil, fmt.Errorf("more than one candidate in Gemini response")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil {
		return nil, fmt.Errorf("candidate content is nil")
	}

	if len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("candidate content has no parts")
	}

	if calls == nil {
		calls = make([]minds.ToolCall, 0)
	}

	return &Response{raw: resp, calls: calls}, nil
}

// String returns the text content if this is a text response.
// For function calls, it returns a formatted representation of the call.
func (r *Response) String() string {
	if len(r.raw.Candidates) == 0 {
		return ""
	}

	if len(r.raw.Candidates[0].Content.Parts) == 0 {
		return ""
	}

	text, ok := r.raw.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return ""
	}

	return string(text)
}

// ToolCalls returns the function call details if this is a function call response.
// Returns nil and false if this isn't a function call.
func (r *Response) ToolCalls() []minds.ToolCall {
	return r.calls
}

// Raw returns the underlying Gemini response
func (r *Response) Raw() *genai.GenerateContentResponse {
	return r.raw
}

func (r *Response) Messages() (minds.Messages, error) {
	messages := minds.Messages{}

	for _, call := range r.calls {
		messages = append(messages, minds.Message{
			Role:       minds.RoleTool,
			Content:    string(call.Function.Result),
			ToolCallID: call.ID,
		})
	}

	for _, part := range r.raw.Candidates[0].Content.Parts {
		text, ok := part.(genai.Text)
		if ok {
			messages = append(messages, minds.Message{
				Role:    minds.Role(r.raw.Candidates[0].Content.Role),
				Content: string(text),
			})
		}
	}

	return messages, nil
}
