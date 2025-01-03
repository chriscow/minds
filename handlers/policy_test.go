package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

type mockResponse struct {
	Valid     bool   `json:"valid"`
	Reason    string `json:"reason"`
	Violation string `json:"violation"`
}

func (m *mockResponse) String() string {
	return m.Reason
}

func (m *mockResponse) Type() minds.ResponseType {
	return minds.ResponseTypeText
}

func (m *mockResponse) Text() (string, bool) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", false
	}

	return string(b), true
}

func (m *mockResponse) ToolCalls() ([]minds.ToolCall, bool) {
	return nil, false
}

func (m *mockResponse) ToMessages() (minds.Messages, error) {
	return minds.Messages{
		{
			Role:    minds.RoleAssistant,
			Content: m.Reason,
		},
	}, nil
}

type mockContentGenerator struct {
	mockResponse mockResponse
	mockError    error
}

func (m *mockContentGenerator) ModelName() string {
	return "mock-model"
}

func (m *mockContentGenerator) GenerateContent(ctx context.Context, req minds.Request) (minds.Response, error) {
	if m.mockError != nil {
		return nil, m.mockError
	}

	return &m.mockResponse, nil
}

func (m *mockContentGenerator) Close() {}

type mockResultFn struct {
	result handlers.PolicyResult
	err    error
}

func (m *mockResultFn) handleResult(ctx context.Context, tc minds.ThreadContext, res handlers.PolicyResult) error {
	m.result = res
	return m.err
}

func TestPolicy_ContextCanceled(t *testing.T) {
	is := is.New(t)

	mockLLM := &mockContentGenerator{}

	validator := handlers.Policy(mockLLM, "TestPolicy", "Validate the following content.", nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tc := minds.NewThreadContext(ctx).WithMessages(minds.Messages{
		{Role: minds.RoleUser, Content: "Test message."},
	})
	_, err := validator.HandleThread(tc, nil)
	is.True(err != nil) // Ensure an error occurred
	is.Equal(err.Error(), "context canceled")
}

func TestPolicy_String(t *testing.T) {
	is := is.New(t)

	mockLLM := &mockContentGenerator{}

	validator := handlers.Policy(mockLLM, "TestPolicy", "Validate the following content.", nil)

	is.Equal(validator.String(), "Policy(TestPolicy)")
}

func TestPolicy_HandleThread_Success(t *testing.T) {
	is := is.New(t)

	mockLLM := &mockContentGenerator{
		mockResponse: mockResponse{
			Valid:     true,
			Reason:    "Content is valid.",
			Violation: "",
		},
	}

	rf := &mockResultFn{}

	validator := handlers.Policy(mockLLM, "TestPolicy", "Validate the following content.", rf.handleResult)

	ctx := context.Background()
	msgIn := minds.Messages{
		{Role: minds.RoleUser, Content: "Test message."},
	}

	tc := minds.NewThreadContext(ctx).WithMessages(msgIn)

	tcOut, err := validator.HandleThread(tc, minds.NoopThreadHandler{})
	msgOut := tcOut.Messages()

	is.NoErr(err)                                   // Ensure no error occurred
	is.Equal(len(msgOut), len(msgIn))               // Ensure the message count matches
	is.Equal(msgOut[0].Role, msgIn[0].Role)         // Ensure the role matches
	is.Equal(msgOut[0].Content, msgIn[0].Content)   // Ensure the content matches
	is.Equal(rf.result.Valid, true)                 // Ensure the result is valid
	is.Equal(rf.result.Reason, "Content is valid.") // Ensure the reason matches
	is.Equal(rf.result.Violation, "")               // Ensure the violation is empty
}

func TestPolicy_HandleThread_InvalidResponse(t *testing.T) {
	is := is.New(t)

	mockLLM := &mockContentGenerator{
		mockResponse: mockResponse{
			Valid:     false,
			Reason:    "Content is invalid.",
			Violation: "Test violation.",
		},
	}

	t.Run("NoResultFn", func(t *testing.T) {
		validator := handlers.Policy(mockLLM, "TestPolicy", "Validate the following content.", nil)

		msgIn := minds.Messages{
			{Role: minds.RoleUser, Content: "Test message."},
		}

		tc := minds.NewThreadContext(context.Background()).WithMessages(msgIn)

		_, err := validator.HandleThread(tc, nil)
		is.True(err != nil) // Ensure an error occurred
		is.Equal(err.Error(), "policy validation failed: Content is invalid.")
	})

	t.Run("With ResultFnError", func(t *testing.T) {

		fn := &mockResultFn{
			err: errors.New("Content is invalid."),
		}

		validator := handlers.Policy(mockLLM, "TestPolicy", "Validate the following content.", fn.handleResult)

		msgIn := minds.Messages{
			{Role: minds.RoleUser, Content: "Test message."},
		}

		tc := minds.NewThreadContext(context.Background()).WithMessages(msgIn)

		_, err := validator.HandleThread(tc, nil)
		is.True(err != nil) // Ensure an error occurred
		is.Equal(err.Error(), "Content is invalid.")
	})
}

func TestPolicy_HandleThread_GenerationError(t *testing.T) {
	is := is.New(t)

	mockLLM := &mockContentGenerator{
		mockError: errors.New("generation failed"),
	}

	validator := handlers.Policy(mockLLM, "TestPolicy", "Validate the following content.", nil)

	msgIn := minds.Messages{
		{Role: minds.RoleUser, Content: "Test message."},
	}

	tc := minds.NewThreadContext(context.Background()).WithMessages(msgIn)

	_, err := validator.HandleThread(tc, nil)
	is.True(err != nil)                                                              // Ensure an error occurred
	is.Equal(err.Error(), "policy validation failed to generate: generation failed") // Ensure error message matches
}

func TestPolicy_HandleThread_Failure_NoResultFn(t *testing.T) {
	is := is.New(t)

	mockLLM := &mockContentGenerator{
		mockResponse: mockResponse{
			Valid:     false,
			Reason:    "Content is invalid.",
			Violation: "Test violation.",
		},
	}

	validator := handlers.Policy(mockLLM, "TestPolicy", "Validate the following content.", nil)

	msgIn := minds.Messages{
		{Role: minds.RoleUser, Content: "Test message."},
	}

	tc := minds.NewThreadContext(context.Background()).WithMessages(msgIn)

	_, err := validator.HandleThread(tc, nil)
	is.True(err != nil)                                                    // Ensure an error occurred
	is.Equal(err.Error(), "policy validation failed: Content is invalid.") // Ensure error message matches
}
