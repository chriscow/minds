package tools

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/chriscow/minds"
)

func TestGetDefaultModel(t *testing.T) {
	// Save original env vars to restore later
	origLLM := os.Getenv("LLM_DEFAULT_MODEL")
	origOpenAI := os.Getenv("OPENAI_DEFAULT_MODEL")
	defer func() {
		os.Setenv("LLM_DEFAULT_MODEL", origLLM)
		os.Setenv("OPENAI_DEFAULT_MODEL", origOpenAI)
	}()

	// Test LLM_DEFAULT_MODEL takes precedence
	os.Setenv("LLM_DEFAULT_MODEL", MockModel)
	os.Setenv("OPENAI_DEFAULT_MODEL", GPT41)
	model := getDefaultModel()
	if model != MockModel {
		t.Fatalf("expected %s, got %s", MockModel, model)
	}

	// Test fallback to OPENAI_DEFAULT_MODEL
	os.Setenv("LLM_DEFAULT_MODEL", "")
	os.Setenv("OPENAI_DEFAULT_MODEL", GPT41)
	model = getDefaultModel()
	if model != GPT41 {
		t.Fatalf("expected %s, got %s", GPT41, model)
	}

	// Test fallback to GPT41Nano when no env vars set
	os.Setenv("LLM_DEFAULT_MODEL", "")
	os.Setenv("OPENAI_DEFAULT_MODEL", "")
	model = getDefaultModel()
	if model != GPT41Nano {
		t.Fatalf("expected %s, got %s", GPT41Nano, model)
	}
}

func TestWithModel(t *testing.T) {
	o := &options{model: "default-model"}

	// Apply the WithModel option
	opt := WithModel(GPT41)
	opt(o)

	if o.model != GPT41 {
		t.Fatalf("WithModel failed to set model, expected %s, got %s", GPT41, o.model)
	}
}

func TestAsk(t *testing.T) {
	// Save original env vars to restore later
	origLLM := os.Getenv("LLM_DEFAULT_MODEL")
	defer os.Setenv("LLM_DEFAULT_MODEL", origLLM)

	// Test with Mock model via environment variable
	os.Setenv("LLM_DEFAULT_MODEL", MockModel)
	MockLLMResponse = "mock-llm-response"
	MockLLMError = nil

	response, err := Ask(context.Background(), "What is the capital of the moon?")
	if err != nil {
		t.Fatalf("failed to ask: %v", err)
	}

	if response != MockLLMResponse {
		t.Fatalf("expected %s, got %s", MockLLMResponse, response)
	}

	// Test with explicit WithModel
	os.Setenv("LLM_DEFAULT_MODEL", "something-else")
	response, err = Ask(context.Background(), "What is the capital of the moon?", WithModel(MockModel))
	if err != nil {
		t.Fatalf("failed to ask with explicit model: %v", err)
	}

	if response != MockLLMResponse {
		t.Fatalf("expected %s, got %s with explicit model", MockLLMResponse, response)
	}

	// Test unknown model error
	response, err = Ask(context.Background(), "prompt", WithModel("unknown-model"))
	if err == nil {
		t.Fatal("expected error for unknown model, got nil")
	}
	if response != "" {
		t.Fatalf("expected empty response for error case, got %s", response)
	}
}

func TestAskWithError(t *testing.T) {
	// Test with Mock model that returns an error
	os.Setenv("LLM_DEFAULT_MODEL", MockModel)
	MockLLMResponse = "error response"
	expectedErr := errors.New("mock error")
	MockLLMError = expectedErr

	// This should not error since we're not using the error in the mock path for Ask
	response, err := Ask(context.Background(), "What is the capital of the moon?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response != MockLLMResponse {
		t.Fatalf("expected %s, got %s", MockLLMResponse, response)
	}
}

func TestStructuredAsk(t *testing.T) {
	// Save original env vars to restore later
	origLLM := os.Getenv("LLM_DEFAULT_MODEL")
	defer os.Setenv("LLM_DEFAULT_MODEL", origLLM)

	// Test with Mock model
	os.Setenv("LLM_DEFAULT_MODEL", MockModel)
	MockLLMResponse = `{"answer": "mock-llm-response"}`
	MockLLMError = nil

	type result struct {
		Answer string `json:"answer"`
	}

	response, err := StructuredAsk[result](context.Background(), "answer", "What is the capital of the moon?")
	if err != nil {
		t.Fatalf("failed to ask: %v", err)
	}

	expectedAnswer := "mock-llm-response"
	if response.Answer != expectedAnswer {
		t.Fatalf("expected %s, got %s", expectedAnswer, response.Answer)
	}

	// Test with explicit WithModel
	os.Setenv("LLM_DEFAULT_MODEL", "something-else")
	response, err = StructuredAsk[result](context.Background(), "answer", "What is the capital of the moon?", WithModel(MockModel))
	if err != nil {
		t.Fatalf("failed to ask with explicit model: %v", err)
	}

	if response.Answer != expectedAnswer {
		t.Fatalf("expected %s, got %s with explicit model", expectedAnswer, response.Answer)
	}

	// Test unknown model error
	var emptyResponse result
	response, err = StructuredAsk[result](context.Background(), "answer", "prompt", WithModel("unknown-model"))
	if err == nil {
		t.Fatal("expected error for unknown model, got nil")
	}
	if response != emptyResponse {
		t.Fatalf("expected empty response for error case, got %+v", response)
	}
}

func TestStructuredAskIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Fatalf("OPENAI_API_KEY is not set")
	}

	type result struct {
		LuckyNumber int `json:"lucky_number"`
	}

	response, err := StructuredAsk[result](context.Background(), "lucky_number", "What is today's lucky number greater than 0?")
	if err != nil {
		t.Fatalf("failed to ask: %v", err)
	}

	if response.LuckyNumber <= 0 {
		t.Fatalf("expected lucky number greater than 0, got %d", response.LuckyNumber)
	}

	// test with deepseek
	response, err = StructuredAsk[result](context.Background(), "lucky_number", "What is today's lucky number greater than 0? Respond in JSON.", WithModel(DeepSeekChat))
	if err != nil {
		t.Fatalf("failed to ask: %v", err)
	}

	if response.LuckyNumber <= 0 {
		t.Fatalf("expected lucky number greater than 0, got %d", response.LuckyNumber)
	}
}

func TestStructuredAskWithError(t *testing.T) {
	// Test unmarshal error
	os.Setenv("LLM_DEFAULT_MODEL", MockModel)
	MockLLMResponse = `{"invalid json`
	MockLLMError = nil

	type result struct {
		Answer string `json:"answer"`
	}

	var emptyResponse result
	response, err := StructuredAsk[result](context.Background(), "answer", "What is the capital of the moon?")
	if err == nil {
		t.Fatalf("expected unmarshal error, got nil")
	}
	if response != emptyResponse {
		t.Fatalf("expected empty response for error case, got %+v", response)
	}

	// Test with Mock error
	MockLLMResponse = `{"answer": "mock-llm-response"}`
	MockLLMError = errors.New("mock error")

	_, err = StructuredAsk[result](context.Background(), "answer", "What is the capital of the moon?")
	if err != MockLLMError {
		t.Fatalf("expected %v, got %v", MockLLMError, err)
	}
}

func TestAskWithDeepSeekURL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("skipping test because OPENAI_API_KEY is not set")
	}

	// Set up test context
	ctx := context.Background()
	prompt := "Say 'cheese!"

	// Test with DeepSeek API URL via WithModel option
	resp1, err := Ask(ctx, prompt, WithModel(DeepSeekChat))
	if err != nil {
		t.Fatalf("Ask with DeepSeek URL failed: %v", err)
	}

	t.Logf("Response: %s", resp1)
	if resp1 == "" {
		t.Fatal("Expected non-empty response, got empty")
	}
}

func TestAPIKeyPassthrough(t *testing.T) {
	// Save original env var to restore later
	origAPIKey := os.Getenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origAPIKey)

	// Clear environment variable to ensure we're testing option passing
	os.Setenv("OPENAI_API_KEY", "")

	// Test 1: Custom API key provided via WithAPIKey option
	customAPIKey := "test-api-key"
	opts := GetOptionsFromAskOptions(WithModel(DeepSeekChat), WithAPIKey(customAPIKey))

	// Verify the API key was passed correctly
	if opts.apiKey != customAPIKey {
		t.Errorf("API key not passed correctly via options. Expected: %s, Got: %s", customAPIKey, opts.apiKey)
	}

	// Test 2: API key from environment
	testEnvAPIKey := "env-test-api-key"
	os.Setenv("OPENAI_API_KEY", testEnvAPIKey)

	opts = GetOptionsFromAskOptions(WithModel(DeepSeekChat))

	// Verify the API key was retrieved from environment
	if opts.apiKey != testEnvAPIKey {
		t.Errorf("API key not retrieved from environment. Expected: %s, Got: %s", testEnvAPIKey, opts.apiKey)
	}

	// Test 3: Use the mock model to verify the flow works end to end
	origLLM := os.Getenv("LLM_DEFAULT_MODEL")
	defer os.Setenv("LLM_DEFAULT_MODEL", origLLM)

	MockLLMResponse = "mock-api-key-test-response"
	MockLLMError = nil

	response, err := Ask(context.Background(), "test prompt", WithModel(MockModel), WithAPIKey(customAPIKey))
	if err != nil {
		t.Fatalf("Ask failed with custom API key: %v", err)
	}

	if response != MockLLMResponse {
		t.Errorf("Expected response %s, got %s", MockLLMResponse, response)
	}
}

func TestWithMessages(t *testing.T) {
	// Test basic WithMessages functionality
	t.Run("basic_functionality", func(t *testing.T) {
		messages := minds.Messages{
			{Role: minds.RoleSystem, Content: "You are a helpful assistant."},
			{Role: minds.RoleUser, Content: "Hello!"},
		}

		o := &options{}
		opt := WithMessages(messages)
		opt(o)

		if !o.explicitMessages {
			t.Error("explicitMessages should be true after WithMessages")
		}

		if len(o.messages) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(o.messages))
		}

		if o.messages[0].Role != minds.RoleSystem || o.messages[0].Content != "You are a helpful assistant." {
			t.Errorf("system message not set correctly: %+v", o.messages[0])
		}

		if o.messages[1].Role != minds.RoleUser || o.messages[1].Content != "Hello!" {
			t.Errorf("user message not set correctly: %+v", o.messages[1])
		}
	})

	// Test WithMessages with empty slice
	t.Run("empty_messages", func(t *testing.T) {
		messages := minds.Messages{}

		o := &options{}
		opt := WithMessages(messages)
		opt(o)

		if !o.explicitMessages {
			t.Error("explicitMessages should be true even with empty messages")
		}

		if len(o.messages) != 0 {
			t.Fatalf("expected 0 messages, got %d", len(o.messages))
		}
	})

	// Test WithMessages overwrites previous messages
	t.Run("overwrites_previous", func(t *testing.T) {
		o := &options{
			messages: minds.Messages{
				{Role: minds.RoleUser, Content: "Previous message"},
			},
		}

		newMessages := minds.Messages{
			{Role: minds.RoleSystem, Content: "New system message"},
			{Role: minds.RoleUser, Content: "New user message"},
		}

		opt := WithMessages(newMessages)
		opt(o)

		if len(o.messages) != 2 {
			t.Fatalf("expected 2 messages after overwrite, got %d", len(o.messages))
		}

		if o.messages[0].Content != "New system message" {
			t.Error("messages were not properly overwritten")
		}
	})
}

func TestWithMessagesIntegration(t *testing.T) {
	// Save original env vars to restore later
	origLLM := os.Getenv("LLM_DEFAULT_MODEL")
	defer os.Setenv("LLM_DEFAULT_MODEL", origLLM)

	// Test WithMessages with Ask function using mock
	t.Run("with_ask_function", func(t *testing.T) {
		os.Setenv("LLM_DEFAULT_MODEL", MockModel)
		MockLLMResponse = "mock-response-from-messages"
		MockLLMError = nil

		messages := minds.Messages{
			{Role: minds.RoleSystem, Content: "You are a helpful assistant."},
			{Role: minds.RoleUser, Content: "What is 2+2?"},
		}

		response, err := Ask(context.Background(), "This prompt should be ignored", WithMessages(messages))
		if err != nil {
			t.Fatalf("Ask with WithMessages failed: %v", err)
		}

		if response != MockLLMResponse {
			t.Fatalf("expected %s, got %s", MockLLMResponse, response)
		}
	})

	// Test WithMessages with StructuredAsk function using mock
	t.Run("with_structured_ask", func(t *testing.T) {
		os.Setenv("LLM_DEFAULT_MODEL", MockModel)
		MockLLMResponse = `{"answer": "from-messages"}`
		MockLLMError = nil

		type result struct {
			Answer string `json:"answer"`
		}

		messages := minds.Messages{
			{Role: minds.RoleSystem, Content: "Respond with JSON."},
			{Role: minds.RoleUser, Content: "Give me an answer."},
		}

		response, err := StructuredAsk[result](context.Background(), "answer", "This should be ignored", WithMessages(messages))
		if err != nil {
			t.Fatalf("StructuredAsk with WithMessages failed: %v", err)
		}

		if response.Answer != "from-messages" {
			t.Fatalf("expected 'from-messages', got %s", response.Answer)
		}
	})

	// Test WithMessages combined with other options
	t.Run("with_other_options", func(t *testing.T) {
		messages := minds.Messages{
			{Role: minds.RoleUser, Content: "Test message"},
		}

		opts := GetOptionsFromAskOptions(
			WithModel(GPT41),
			WithMaxTokens(500),
			WithMessages(messages),
			WithAPIKey("test-key"),
		)

		if opts.model != GPT41 {
			t.Errorf("model not set correctly: expected %s, got %s", GPT41, opts.model)
		}

		if opts.maxTokens != 500 {
			t.Errorf("maxTokens not set correctly: expected 500, got %d", opts.maxTokens)
		}

		if opts.apiKey != "test-key" {
			t.Errorf("apiKey not set correctly: expected 'test-key', got %s", opts.apiKey)
		}

		if !opts.explicitMessages {
			t.Error("explicitMessages should be true")
		}

		if len(opts.messages) != 1 || opts.messages[0].Content != "Test message" {
			t.Error("messages not set correctly")
		}
	})
}

func TestWithSystemMessageAndPrefillCompatibility(t *testing.T) {
	// Test that WithSystemMessage and WithPrefill still work and convert to messages format
	t.Run("system_message_conversion", func(t *testing.T) {
		o := &options{}

		opt := WithSystemMessage("You are a helpful assistant.")
		opt(o)

		if !o.explicitMessages {
			t.Error("explicitMessages should be true after WithSystemMessage")
		}

		if len(o.messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(o.messages))
		}

		if o.messages[0].Role != minds.RoleSystem {
			t.Errorf("expected system role, got %s", o.messages[0].Role)
		}

		if o.messages[0].Content != "You are a helpful assistant." {
			t.Errorf("system message content incorrect: %s", o.messages[0].Content)
		}
	})

	t.Run("prefill_conversion", func(t *testing.T) {
		o := &options{}

		opt := WithPrefill("Sure, I can help with that.")
		opt(o)

		if !o.explicitMessages {
			t.Error("explicitMessages should be true after WithPrefill")
		}

		if len(o.messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(o.messages))
		}

		if o.messages[0].Role != minds.RoleAssistant {
			t.Errorf("expected assistant role, got %s", o.messages[0].Role)
		}

		if o.messages[0].Content != "Sure, I can help with that." {
			t.Errorf("prefill message content incorrect: %s", o.messages[0].Content)
		}
	})

	t.Run("system_and_prefill_combined", func(t *testing.T) {
		o := &options{}

		// Apply system message first
		systemOpt := WithSystemMessage("You are helpful.")
		systemOpt(o)

		// Then add a user message
		o.messages = append(o.messages, minds.Message{
			Role: minds.RoleUser,
			Content: "Hello",
		})

		// Then apply prefill
		prefillOpt := WithPrefill("Hello! How can I help?")
		prefillOpt(o)

		if len(o.messages) != 3 {
			t.Fatalf("expected 3 messages, got %d", len(o.messages))
		}

		// Check order: system, user, assistant
		if o.messages[0].Role != minds.RoleSystem {
			t.Error("first message should be system")
		}
		if o.messages[1].Role != minds.RoleUser {
			t.Error("second message should be user")
		}
		if o.messages[2].Role != minds.RoleAssistant {
			t.Error("third message should be assistant")
		}
	})
}
