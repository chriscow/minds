package handlers

import (
	"context"
	"testing"

	"github.com/chriscow/minds"
	"github.com/matryer/is"
)

// MockContentGenerator simulates an LLM that returns predetermined responses
type MockContentGenerator struct {
	response string
	err      error
}

func (m *MockContentGenerator) ModelName() string {
	return "mock-model"
}

func (m *MockContentGenerator) GenerateContent(_ context.Context, _ minds.Request) (minds.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &MockResponse{content: m.response}, nil
}

func (m *MockContentGenerator) Close() {}

// MockResponse implements the minds.Response interface
type MockResponse struct {
	content string
}

func (m *MockResponse) String() string {
	return m.content
}

func (m *MockResponse) ToolCalls() []minds.ToolCall {
	return nil
}

func TestFreeformExtractor(t *testing.T) {
	is := is.New(t)

	// Mock response with key-value pairs in the new format with string values
	mockResponse := `{"pairs": [{"key": "name", "value": "John Doe"}, {"key": "age", "value": "30"}, {"key": "email", "value": "john@example.com"}]}`

	// Create a mock content generator
	generator := &MockContentGenerator{response: mockResponse}

	// Create a thread context with messages
	tc := minds.NewThreadContext(context.Background())
	tc.AppendMessages(
		minds.Message{Role: minds.RoleUser, Content: "Hello, my name is John Doe"},
		minds.Message{Role: minds.RoleAssistant, Content: "Hi John, how can I help?"},
		minds.Message{Role: minds.RoleUser, Content: "I'm 30 years old and my email is john@example.com"},
	)

	// Create a FreeformExtractor handler
	extractor := NewFreeformExtractor(
		"test-extractor",
		generator,
		"Extract the following information from the conversation: name, age, and email.",
	)

	// Process the thread
	result, err := extractor.HandleThread(tc, nil)

	// Verify no error occurred
	is.NoErr(err)

	// Verify the metadata contains the extracted key-value pairs
	metadata := result.Metadata()
	is.Equal(metadata["name"], "John Doe")
	is.Equal(metadata["age"], int64(30)) // Parsed from string to int64
	is.Equal(metadata["email"], "john@example.com")
}

func TestFreeformExtractor_WithNext(t *testing.T) {
	is := is.New(t)

	// Mock response with key-value pairs in the new format with string values
	mockResponse := `{"pairs": [{"key": "name", "value": "John Doe"}, {"key": "age", "value": "30"}]}`

	// Create a mock content generator
	generator := &MockContentGenerator{response: mockResponse}

	// Create a thread context with messages
	tc := minds.NewThreadContext(context.Background())
	tc.AppendMessages(
		minds.Message{Role: minds.RoleUser, Content: "Hello, my name is John Doe and I'm 30 years old"},
	)

	// Create a next handler that adds more metadata
	nextHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
		tc.SetKeyValue("processed_by_next", true)
		return tc, nil
	})

	// Create a FreeformExtractor handler
	extractor := NewFreeformExtractor(
		"test-extractor",
		generator,
		"Extract the following information from the conversation: name and age.",
	)

	// Process the thread
	result, err := extractor.HandleThread(tc, nextHandler)

	// Verify no error occurred
	is.NoErr(err)

	// Verify the metadata contains both the extracted key-value pairs and the next handler's addition
	metadata := result.Metadata()
	is.Equal(metadata["name"], "John Doe")
	is.Equal(metadata["age"], int64(30))
	is.Equal(metadata["processed_by_next"], true)
}

func TestFreeformExtractor_WithMiddleware(t *testing.T) {
	is := is.New(t)

	// Mock response with key-value pairs in the new format with string values
	mockResponse := `{"pairs": [{"key": "name", "value": "John Doe"}]}`

	// Create a mock content generator
	generator := &MockContentGenerator{response: mockResponse}

	// Create a thread context with messages
	tc := minds.NewThreadContext(context.Background())
	tc.AppendMessages(
		minds.Message{Role: minds.RoleUser, Content: "Hello, my name is John Doe"},
	)

	// Create middleware that adds metadata
	middleware := minds.MiddlewareFunc(func(next minds.ThreadHandler) minds.ThreadHandler {
		return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			tc.SetKeyValue("middleware_applied", true)
			return next.HandleThread(tc, nil)
		})
	})

	// Create a FreeformExtractor handler with middleware
	extractor := NewFreeformExtractor(
		"test-extractor",
		generator,
		"Extract the name from the conversation.",
	)
	extractor.Use(middleware)

	// Process the thread
	result, err := extractor.HandleThread(tc, nil)

	// Verify no error occurred
	is.NoErr(err)

	// Verify the metadata contains both the extracted key-value pairs and the middleware's addition
	metadata := result.Metadata()
	is.Equal(metadata["name"], "John Doe")
	is.Equal(metadata["middleware_applied"], true)
}
