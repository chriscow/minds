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

// switcher implements a conditional branching handler that executes the first matching
// case's handler or falls back to a default handler if no conditions match.
type switcher struct {
	name           string
	cases          []SwitchCase
	defaultHandler minds.ThreadHandler
}

// Switch creates a new Switch handler that executes the first matching case's
// handler. If no cases match, it executes the default handler. The name
// parameter is used for debugging and logging purposes.

// The SwitchCase struct pairs a `SwitchCondition` interface with a handler.
// When the condition evaluates to true, the corresponding handler is executed.

// Example:

// 	    // The MetadataEquals condition checks if the metadata key "type" equals "question"
// 		metadata := MetadataEquals{Key: "type", Value: "question"}
// 		questionHandler := SomeQuestionHandler()
// 		defaultHandler := DefaultHandler()

// sw := Switch("type-switch",

// 	defaultHandler,
// 	SwitchCase{metadata, questionHandler},

// )
func Switch(name string, defaultHandler minds.ThreadHandler, cases ...SwitchCase) *switcher {
	return &switcher{
		name:           name,
		cases:          cases,
		defaultHandler: defaultHandler,
	}
}

// String returns a string representation of the switcher, primarily for debugging
// and logging purposes.
func (s *switcher) String() string {
	return fmt.Sprintf("Switch: %s", s.name)
}

// HandleThread processes the thread context by evaluating each case's condition in order
// until a match is found, then executes the corresponding handler. If no cases match,
// it executes the default handler if one is provided.
//
// It returns the modified thread context and any error that occurred during processing.
// The next handler is passed to the matched case's handler for further processing.
func (s *switcher) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	for _, c := range s.cases {
		matches, err := c.Condition.Evaluate(tc)
		if err != nil {
			return tc, fmt.Errorf("error evaluating condition: %w", err)
		}
		if matches {
			return c.Handler.HandleThread(tc, next)
		}
	}

	if s.defaultHandler != nil {
		return s.defaultHandler.HandleThread(tc, next)
	}

	return tc, nil
}

// MetadataEquals implements SwitchCondition to check if a metadata key equals a specific value.
// The comparison uses Go's standard equality operator (==).
type MetadataEquals struct {
	Key   string      // The metadata key to check
	Value interface{} // The value to compare against
}

// Evaluate checks if the metadata value for the specified key equals the target value.
// Returns false if the key doesn't exist in the metadata.
func (m MetadataEquals) Evaluate(tc minds.ThreadContext) (bool, error) {
	val, exists := tc.Metadata()[m.Key]
	if !exists {
		return false, nil
	}
	return val == m.Value, nil
}

// LLMCondition implements SwitchCondition using an LLM to evaluate a condition based
// on a provided prompt. It uses the content generator to process the prompt and
// expects a boolean response.
type LLMCondition struct {
	Generator minds.ContentGenerator // The LLM content generator to use
	Prompt    string                 // The prompt to send to the LLM
}

// boolResp represents the expected JSON response format from the LLM.
type boolResp struct {
	Bool bool `json:"bool"`
}

// Evaluate sends the prompt to the LLM along with context from the thread's last message
// and interprets the response as a boolean value. The LLM's response must conform to
// the boolean_response schema.
//
// It returns an error if the LLM generation fails or if the response cannot be parsed
// as a boolean value.
func (l LLMCondition) Evaluate(tc minds.ThreadContext) (bool, error) {
	// Create a new context with the prompt
	messages := minds.Messages{{Role: minds.RoleUser, Content: l.Prompt}}

	// Add the current thread's last message as context
	if len(tc.Messages()) > 0 {
		lastMsg := tc.Messages().Last()
		messages = append(messages, minds.Message{
			Role:    minds.RoleSystem,
			Content: fmt.Sprintf("Previous message: %s", lastMsg.Content),
		})
	}

	schema, err := minds.NewResponseSchema("boolean_response", "Requires a boolean response", boolResp{})
	if err != nil {
		return false, fmt.Errorf("error creating response schema: %w", err)
	}

	req := minds.NewRequest(messages, minds.WithResponseSchema(*schema))

	resp, err := l.Generator.GenerateContent(tc.Context(), req)
	if err != nil {
		return false, fmt.Errorf("error generating LLM response: %w", err)
	}

	result := boolResp{}
	if err := json.Unmarshal([]byte(resp.String()), &result); err != nil {
		return false, fmt.Errorf("error unmarshalling response: %w", err)
	}

	return result.Bool, nil
}
