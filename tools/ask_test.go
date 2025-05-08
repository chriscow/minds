package tools

import (
	"context"
	"errors"
	"os"
	"testing"
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
