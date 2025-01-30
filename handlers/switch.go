package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/chriscow/minds"
)

// SwitchCondition represents a condition that can be evaluated against a thread context.
// Implementations of this interface are used by the Switch handler to determine which
// case should be executed.
type SwitchCondition interface {
	// Evaluate examines the thread context and returns true if the condition is met.
	// It returns an error if the evaluation fails.
	Evaluate(tc minds.ThreadContext) (bool, error)
}

// SwitchCase pairs a condition with a handler. When the condition evaluates to true,
// the corresponding handler is executed.
type SwitchCase struct {
	Condition SwitchCondition
	Handler   minds.ThreadHandler
}

// Switch executes the first matching case's handler, or the default handler if no cases match.
type Switch struct {
	name           string
	cases          []SwitchCase
	defaultHandler minds.ThreadHandler
	middleware     []minds.Middleware
}

// NewSwitch creates a new Switch handler that executes the first matching case's
// handler. If no cases match, it executes the default handler. The name
// parameter is used for debugging and logging purposes.
//
// The SwitchCase struct pairs a `SwitchCondition` interface with a handler.
// When the condition evaluates to true, the corresponding handler is executed.
//
// Example:
//
//	    // The MetadataEquals condition checks if the metadata key "type" equals "question"
//		metadata := MetadataEquals{Key: "type", Value: "question"}
//		questionHandler := SomeQuestionHandler()
//		defaultHandler := DefaultHandler()
//
// sw := Switch("type-switch",
//
//	defaultHandler,
//	SwitchCase{metadata, questionHandler},
//
// )
func NewSwitch(name string, defaultHandler minds.ThreadHandler, cases ...SwitchCase) *Switch {
	return &Switch{
		name:           name,
		cases:          cases,
		defaultHandler: defaultHandler,
	}
}

// Use applies middleware to the Switch handler.
func (s *Switch) Use(middleware ...minds.Middleware) {
	s.middleware = append(s.middleware, middleware...)
}

// With returns a new Switch handler with additional middleware, preserving existing state.
func (s *Switch) With(middleware ...minds.Middleware) *Switch {
	newSwitch := &Switch{
		name:           s.name,
		cases:          append([]SwitchCase{}, s.cases...),
		defaultHandler: s.defaultHandler,
		middleware:     append([]minds.Middleware{}, s.middleware...),
	}
	newSwitch.Use(middleware...)
	return newSwitch
}

// HandleThread processes the thread context, executing the first matching case's handler.
func (s *Switch) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	for _, c := range s.cases {
		matches, err := c.Condition.Evaluate(tc)
		if err != nil {
			return tc, fmt.Errorf("%s: error evaluating condition: %w", s.name, err)
		}
		if matches {
			return s.executeWithMiddleware(tc, c.Handler, next)
		}
	}

	// No cases matched, use the default handler
	if s.defaultHandler != nil {
		return s.executeWithMiddleware(tc, s.defaultHandler, next)
	}

	return tc, nil
}

// executeWithMiddleware applies middleware to a handler before execution.
func (s *Switch) executeWithMiddleware(tc minds.ThreadContext, handler minds.ThreadHandler, next minds.ThreadHandler) (minds.ThreadContext, error) {
	wrappedHandler := handler

	// Apply middleware in reverse order
	for i := len(s.middleware) - 1; i >= 0; i-- {
		wrappedHandler = s.middleware[i].Wrap(wrappedHandler)
	}

	return wrappedHandler.HandleThread(tc, next)
}

// String returns a string representation of the Switch handler.
func (s *Switch) String() string {
	return fmt.Sprintf("Switch(%s)", s.name)
}

// --- Condition Implementations ---

// MetadataEquals checks if a metadata key equals a specific value.
type MetadataEquals struct {
	Key   string
	Value interface{}
}

// Evaluate checks if the metadata value for the specified key equals the target value.
func (m MetadataEquals) Evaluate(tc minds.ThreadContext) (bool, error) {
	val, exists := tc.Metadata()[m.Key]
	if !exists {
		return false, nil
	}
	return val == m.Value, nil
}

// LLMCondition evaluates a condition using an LLM response.
type LLMCondition struct {
	Generator minds.ContentGenerator
	Prompt    string
}

// BoolResp represents the expected JSON response format from the LLM.
type BoolResp struct {
	Bool bool `json:"bool"`
}

// Evaluate sends a prompt to the LLM and expects a boolean response.
func (l LLMCondition) Evaluate(tc minds.ThreadContext) (bool, error) {
	messages := minds.Messages{{Role: minds.RoleUser, Content: l.Prompt}}

	// Add the current thread's last message as context
	if len(tc.Messages()) > 0 {
		lastMsg := tc.Messages().Last()
		messages = append(messages, minds.Message{
			Role:    minds.RoleSystem,
			Content: fmt.Sprintf("Previous message: %s", lastMsg.Content),
		})
	}

	schema, err := minds.NewResponseSchema("boolean_response", "Requires a boolean response", BoolResp{})
	if err != nil {
		return false, fmt.Errorf("error creating response schema: %w", err)
	}

	req := minds.NewRequest(messages, minds.WithResponseSchema(*schema))
	resp, err := l.Generator.GenerateContent(tc.Context(), req)
	if err != nil {
		return false, fmt.Errorf("error generating LLM response: %w", err)
	}

	result := BoolResp{}
	if err := json.Unmarshal([]byte(resp.String()), &result); err != nil {
		return false, fmt.Errorf("error unmarshalling response: %w", err)
	}

	return result.Bool, nil
}
