package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chriscow/minds"
	"github.com/matryer/is"
	"github.com/sashabaranov/go-openai"
)

func newMockResponse() openai.ChatCompletionResponse {
	return openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{{
			Message: openai.ChatCompletionMessage{
				Role:    "assistant",
				Content: "Hello, world!",
			},
			FinishReason: "stop",
		},
		},
	}
}

func newMockFunction() minds.CallableFunc {
	return func(_ context.Context, args []byte) ([]byte, error) {
		var params struct {
			Value int `json:"value"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return nil, err
		}
		result := map[string]int{"result": params.Value * 2}
		return json.Marshal(result)
	}
}

func newMockTool() (minds.Tool, error) {
	return minds.WrapFunction(
		"mock_function",
		"Doubles the input value",
		struct {
			Value int `json:"value" description:"The value to double"`
		}{},
		newMockFunction(),
	)
}

func newMockToolCallResponse() openai.ChatCompletionResponse {
	mockResponse := newMockResponse() // Ensure this function returns a valid mock response
	mockResponse.Choices[0].Message.ToolCalls = []openai.ToolCall{
		{
			Function: openai.FunctionCall{
				Name:      "mock_function",
				Arguments: `{"value": 3}`,
			},
		},
	}
	mockResponse.Choices[0].FinishReason = openai.FinishReasonToolCalls
	return mockResponse
}

// provider_test.go
func TestProvider_GenerateContent(t *testing.T) {
	is := is.New(t)

	// Create a test server that returns mock responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(newMockResponse())
	}))
	defer server.Close()

	provider, err := NewProvider(WithBaseURL(server.URL))
	is.NoErr(err) // Provider initialization should not fail

	ctx := context.Background()
	req := minds.Request{
		Messages: minds.Messages{
			{Role: minds.RoleUser, Content: "Hello!"},
		},
	}

	resp, err := provider.GenerateContent(ctx, req)
	is.NoErr(err)                                 // GenerateContent should not return an error
	is.True(resp != nil)                          // Response should not be nil
	is.Equal(resp.Type(), minds.ResponseTypeText) // Ensure it is a text response
	is.Equal(resp.String(), "Hello, world!")      // Ensure the mock response matches

	text, ok := resp.Text()
	is.True(ok)                     // Should be able to extract text
	is.Equal(text, "Hello, world!") // Ensure the mock response matches
}

func TestProvider_HandleThread(t *testing.T) {
	is := is.New(t)

	// Create a test server that returns mock responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(newMockResponse())
	}))
	defer server.Close()

	handler, err := NewProvider(WithBaseURL(server.URL))
	is.NoErr(err) // Provider initialization should not fail

	ctx := context.Background()

	thread := minds.NewThreadContext(ctx).WithMessages(minds.Messages{
		{Role: minds.RoleUser, Content: "Hi there!"},
	})

	result, err := handler.HandleThread(thread, nil)
	is.NoErr(err)                                            // HandleThread should not return an error
	is.True(len(result.Messages()) == 2)                     // There should be two messages: user and assistant
	is.Equal(result.Messages()[1].Role, minds.RoleAssistant) // Ensure the response role is assistant
	is.Equal(result.Messages()[1].Content, "Hello, world!")  // Ensure the mock response matches
}

// Define a mock function, register in the tool registry and pass it to the provider.
// Ensure the provider calls the function and returns the result.
func TestProvider_GenerateContent_WithToolRegistry(t *testing.T) {
	is := is.New(t)

	//
	// Setup
	//

	// Define a mock function to register in the tool registry
	tool, err := newMockTool()
	is.NoErr(err) // Function wrapping should not fail

	toolRegistry := minds.NewToolRegistry()
	err = toolRegistry.Register(tool)
	is.NoErr(err) // Tool registration should not fail

	mockResponse := newMockToolCallResponse()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	provider, err := NewProvider(WithBaseURL(server.URL), WithToolRegistry(toolRegistry))
	is.NoErr(err) // Provider initialization should not fail

	req := minds.Request{
		Messages: minds.Messages{
			{
				Role:    minds.RoleUser,
				Content: "Call mock_function with value 3",
			},
		},
	}

	//
	// Run the test
	//
	ctx := context.Background()
	resp, err := provider.GenerateContent(ctx, req)
	is.NoErr(err)                                     // GenerateContent should not return an error
	is.True(resp != nil)                              // Response should not be nil
	is.Equal(resp.Type(), minds.ResponseTypeToolCall) // Ensure it is a tool call response
	str := resp.String()
	is.Equal(str, "mock_function") // Ensure the mock function was called

	toolCalls, ok := resp.ToolCalls()
	is.True(ok)                                           // Should be able to extract tool calls
	is.Equal(len(toolCalls), 1)                           // Ensure there is exactly one tool call
	is.Equal(toolCalls[0].Function.Name, "mock_function") // Ensure the function name matches

	var result map[string]int
	is.NoErr(json.Unmarshal(toolCalls[0].Function.Result, &result)) // Should be able to parse the result
	is.Equal(result["result"], 6)                                   // Ensure the mock function was called correctly
}
