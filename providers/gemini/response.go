package gemini

import (
	"encoding/json"
	"fmt"

	"github.com/chriscow/minds"

	"github.com/google/generative-ai-go/genai"
)

// Response represents a unified response type that can handle both
// text responses and function calls from the Gemini API
type Response struct {
	raw *genai.GenerateContentResponse
}

// NewResponse creates a new Response from a Gemini API response
func NewResponse(resp *genai.GenerateContentResponse) (*Response, error) {
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in Gemini response")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil {
		return nil, fmt.Errorf("candidate content is nil")
	}

	if len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("candidate content has no parts")
	}

	return &Response{raw: resp}, nil
}

func (r *Response) ToMessages() (minds.Messages, error) {
	switch r.Type() {
	case minds.ResponseTypeText:
		text, _ := r.Text()
		return minds.Messages{
			{
				Role:    minds.Role(r.raw.Candidates[0].Content.Role),
				Content: text,
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

// String returns the text content if this is a text response.
// For function calls, it returns a formatted representation of the call.
func (r *Response) String() string {
	responseType := r.Type()

	switch responseType {
	case minds.ResponseTypeText:
		text, _ := r.Text()
		return text
	case minds.ResponseTypeToolCall:
		calls, _ := r.ToolCalls()
		return fmt.Sprintf("function call: %v", calls)
	default:
		return "Unknown response type"
	}
}

// Type returns the type of this response
func (r *Response) Type() minds.ResponseType {
	if len(r.raw.Candidates) == 0 {
		return minds.ResponseTypeUnknown
	}

	candidate := r.raw.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return minds.ResponseTypeUnknown
	}

	part := candidate.Content.Parts[0]
	switch part.(type) {
	case genai.Text:
		return minds.ResponseTypeText
	case genai.FunctionCall:
		return minds.ResponseTypeToolCall
	default:
		return minds.ResponseTypeUnknown
	}
}

// Text returns the text content if this is a text response.
// Returns empty string and false if this isn't a text response.
func (r *Response) Text() (string, bool) {
	if r.Type() != minds.ResponseTypeText {
		return "", false
	}

	text, ok := r.raw.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", false
	}

	return string(text), true
}

// ToolCalls returns the function call details if this is a function call response.
// Returns nil and false if this isn't a function call.
func (r *Response) ToolCalls() ([]minds.ToolCall, bool) {
	if r.Type() != minds.ResponseTypeToolCall {
		return nil, false
	}

	funcCall, ok := r.raw.Candidates[0].Content.Parts[0].(genai.FunctionCall)
	if !ok {
		return nil, false
	}

	b, err := json.Marshal(funcCall.Args)
	if err != nil {
		return nil, false
	}

	return []minds.ToolCall{
		{
			Function: minds.FunctionCall{
				Name:       funcCall.Name,
				Parameters: b,
			},
		},
	}, true
}

// Raw returns the underlying Gemini response
func (r *Response) Raw() *genai.GenerateContentResponse {
	return r.raw
}
