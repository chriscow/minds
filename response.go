package minds

import (
	"strings"
)

// ResponseType indicates what kind of response we received
type ResponseType int

const (
	ResponseTypeUnknown ResponseType = iota
	ResponseTypeText
	ResponseTypeToolCall
)

type ResponseSchema struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Definition  Definition `json:"schema"`
}

func NewResponseSchema(name, desc string, v any) (*ResponseSchema, error) {
	def, err := GenerateSchema(v)
	if err != nil {
		return nil, err
	}

	return &ResponseSchema{
		Name:        name,
		Description: desc,
		Definition:  *def,
	}, nil
}

type Response interface {
	// String returns a string representation of the response
	String() string

	// Type returns the type of this response (text, function call, etc.)
	Type() ResponseType

	// Text returns the text content if this is a text response.
	Text() (string, bool)

	// ToolCall returns the tool call details if this is a tool call response.
	ToolCalls() ([]ToolCall, bool)

	ToMessages() (Messages, error)
}

type ResponseHandler func(resp Response) error

func (h ResponseHandler) HandleResponse(resp Response) error {
	return h(resp)
}

type funcCallResponse struct {
	calls []ToolCall
}

func NewToolCallResponse(calls []ToolCall) (Response, error) {
	return &funcCallResponse{
		calls: calls,
	}, nil
}

func (r *funcCallResponse) Type() ResponseType {
	return ResponseTypeToolCall
}

func (r *funcCallResponse) String() string {
	names := []string{}
	for _, call := range r.calls {
		names = append(names, call.Function.Name)
	}
	return strings.Join(names, ", ")
}

func (r *funcCallResponse) Text() (string, bool) {
	return "", false
}

func (r *funcCallResponse) ToolCalls() ([]ToolCall, bool) {
	return r.calls, true
}

func (r *funcCallResponse) ToMessages() (Messages, error) {
	messages := make(Messages, 0, len(r.calls))
	for _, call := range r.calls {
		messages = append(messages, Message{
			Role:       RoleTool,
			Content:    string(call.Function.Result),
			ToolCallID: call.ID,
		})
	}

	return messages, nil
}
