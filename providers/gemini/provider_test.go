package gemini

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chriscow/minds"
	"github.com/google/generative-ai-go/genai"
	"github.com/matryer/is"
)

func TestProvider_GenerateContent(t *testing.T) {
	t.Skip("Skipping test: The genai package expects the server to respond with some kind of Protobuf message.")

	is := is.New(t)

	mockResponse := newMockResponse("ai", genai.Text("Hello, world!"))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(mockResponse)
		is.NoErr(err) // Encoding should not fail
	}))
	defer server.Close()

	ctx := context.Background()
	provider, err := NewProvider(ctx, WithBaseURL(server.URL))
	is.NoErr(err) // Provider initialization should not fail

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
	t.Skip("Skipping test: The genai package expects the server to respond with some kind of Protobuf message.")

	is := is.New(t)

	mockResponse := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []genai.Part{
						genai.Text("Hello, World!"),
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	var provider minds.ContentGenerator
	ctx := context.Background()
	provider, err := NewProvider(ctx, WithBaseURL(server.URL))
	is.NoErr(err) // Provider initialization should not fail

	thread := minds.NewThreadContext(ctx).WithMessages(minds.Messages{
		{Role: minds.RoleUser, Content: "Hi there!"},
	})

	handler, ok := provider.(minds.ThreadHandler)
	is.True(ok) // provider should implement the ThreadHandler interface

	result, err := handler.HandleThread(thread, nil)
	msgOut := result.Messages()
	is.NoErr(err)                                 // HandleThread should not return an error
	is.True(len(msgOut) == 2)                     // There should be two messages: user and assistant
	is.Equal(msgOut[1].Role, minds.RoleAssistant) // Ensure the response role is assistant
	is.Equal(msgOut[1].Content, "Hello, world!")  // Ensure the mock response matches
}

func TestProvider_GenerateContent_WithToolRegistry(t *testing.T) {
	t.Skip("Skipping test: The genai package expects the server to respond with some kind of Protobuf message.")

	is := is.New(t)

	// Define a mock function to register in the tool registry
	mockFunction := func(_ context.Context, args []byte) ([]byte, error) {
		var params struct {
			Value int `json:"value"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return nil, err
		}
		result := map[string]int{"result": params.Value * 2}
		return json.Marshal(result)
	}

	// Wrap the mock function in a tool
	tool, err := minds.WrapFunction(
		"mock_function",
		"Doubles the input value",
		struct {
			Value int `json:"value" description:"The value to double"`
		}{},
		mockFunction,
	)
	is.NoErr(err) // Function wrapping should not fail

	toolRegistry := minds.NewToolRegistry()
	err = toolRegistry.Register(tool)
	is.NoErr(err) // Tool registration should not fail

	mockResponse := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []genai.Part{
						&genai.FunctionCall{
							Name: "mock_function",
							Args: map[string]any{"value": 3},
						},
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	ctx := context.Background()
	provider, err := NewProvider(ctx, WithBaseURL(server.URL))
	is.NoErr(err) // Provider initialization should not fail

	req := minds.Request{
		Messages: minds.Messages{
			{
				Role:    minds.RoleUser,
				Content: "Call mock_function with value 3",
			},
		},
	}

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
