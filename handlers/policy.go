package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chriscow/minds"
)

// PolicyResultFunc defines a function to handle the result of a policy validation.
// It takes a context, thread context, and validation result, and returns an error
// if the validation fails or cannot be processed.
type PolicyResultFunc func(ctx context.Context, tc minds.ThreadContext, res PolicyResult) error

// PolicyResult represents the outcome of a policy validation.
type PolicyResult struct {
	Valid     bool   `json:"valid" description:"Whether the content passes policy validation"`
	Reason    string `json:"reason" description:"Explanation for the validation result"`
	Violation string `json:"violation" description:"Description of the specific violation if any"`
}

// Policy performs policy validation on thread content using a content generator (LLM).
type Policy struct {
	llm          minds.ContentGenerator // LLM used for generating validation responses
	name         string                 // Name of the policy validator
	systemPrompt string                 // System message used to guide the LLM during validation
	resultFn     PolicyResultFunc       // Optional function to process validation results
}

// NewPolicy creates a new policy validator handler.
//
// The handler uses an LLM to validate thread content against a given policy. A system prompt
// is used to guide the validation process, and the result is processed by the optional
// result function. If the result function is nil, the handler defaults to checking
// the `Valid` field of the validation result.
//
// Parameters:
//   - llm: A content generator for generating validation responses.
//   - name: The name of the policy validator.
//   - systemPrompt: A prompt describing the policy validation rules.
//   - resultFn: (Optional) Function to process validation results.
//
// Returns:
//   - A thread handler that validates thread content against a policy.
func NewPolicy(llm minds.ContentGenerator, name, systemPrompt string, resultFn PolicyResultFunc) *Policy {
	p := &Policy{
		llm:          llm,
		name:         name,
		systemPrompt: systemPrompt,
		resultFn:     resultFn,
	}

	return p
}

// String returns a string representation of the policy validator.
func (h *Policy) String() string {
	return fmt.Sprintf("Policy(%s)", h.name)
}

// validate implements the core policy validation logic
func (h *Policy) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	ctx := tc.Context()
	if ctx.Err() != nil {
		return tc, ctx.Err()
	}

	schema, err := minds.GenerateSchema(PolicyResult{})
	if err != nil {
		return tc, fmt.Errorf("failed to generate schema: %w", err)
	}

	systemMsg := minds.Message{
		Role:    minds.RoleSystem,
		Content: h.systemPrompt,
	}

	req := minds.Request{
		Options: minds.RequestOptions{
			ResponseSchema: &minds.ResponseSchema{
				Name:        "ValidationResult",
				Description: "Result of policy validation check",
				Definition:  *schema,
			},
		},
		Messages: append(minds.Messages{systemMsg}, tc.Messages()...),
	}

	resp, err := h.llm.GenerateContent(ctx, req)
	if err != nil {
		return tc, fmt.Errorf("policy validation failed to generate: %w", err)
	}

	text := resp.String()
	if text == "" {
		return tc, fmt.Errorf("no response from policy validation")
	}

	var result PolicyResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return tc, fmt.Errorf("failed to unmarshal validation result from response (%s): %w", text, err)
	}

	if ctx.Err() != nil {
		return tc, ctx.Err()
	}

	if h.resultFn != nil {
		if err := h.resultFn(ctx, tc, result); err != nil {
			return tc, err
		}
	} else if !result.Valid {
		return tc, fmt.Errorf("policy validation failed: %s", result.Reason)
	}

	if next != nil {
		return next.HandleThread(tc, nil)
	}

	return tc, nil
}
