package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"

	"github.com/chriscow/minds"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

const (
	OpenAIAPIURLv1 = "https://api.openai.com/v1"
	DeepSeekAPIURL = "https://api.deepseek.com"

	GPT41Nano = "gpt-4.1-nano" // $0.10 input	$0.025 cached input	$0.40 output
	GPT41Mini = "gpt-4.1-mini" // $0.40 input	$0.10 cached input	$1.60 output
	GPT41     = "gpt-4.1"      // $2.00 input	$0.50 cached input	$8.00 output

	GPT4oMini = "gpt-4o-mini" // $0.15 input	$0.075 cached input	$0.60 output

	O4Mini = "o4-mini" // $1.10 input	$0.275 cached input	$4.40 output
	O3Mini = "o3-mini" // $1.10 input	$0.55 cached input	$4.40 output

	// DeepSeek Chat Discounted UTC 16:30-00:30 50% off    10:30 AM – 6:30 PM Mountain Daylight Time (MDT) Mar-Nov
	DeepSeekChat = "deepseek-chat" // $0.27 ($0.135 discouted) input	$0.07 cached input	$1.10 ($0.550 discounted) output
	// DeepSeek Reasoner Discounted UTC 16:30-00:30 75% off (same as chat!) 10:30 AM – 6:30 PM Mountain Daylight Time (MDT) Mar-Nov
	DeepSeekReasoner = "deepseek-reasoner" // $0.55 input $0.14 cached input $2.19 output output

	MockModel = "mock-model"

	maxResponseTokens = 1000
)

var MockLLMResponse = "mock-llm-response"
var MockLLMError error = nil

// Option represents a functional option for configuring LLM requests
type Option func(*options)

// options holds all configurable options for LLM requests
type options struct {
	model     string
	baseURL   string
	apiKey    string
	maxTokens int
	messages  minds.Messages
	response  *openai.ChatCompletionResponse
}

func IsDeepSeekModel(model string) bool {
	return model == DeepSeekChat || model == DeepSeekReasoner
}

func IsOpenAIModel(model string) bool {
	return model == GPT41Nano || model == GPT41Mini || model == GPT41 ||
		model == GPT4oMini || model == O4Mini || model == O3Mini
}

// WithModel returns an Option that sets the model to use
func WithModel(model string) Option {
	return func(o *options) {
		o.model = model
	}
}

func WithMaxTokens(maxTokens int) Option {
	return func(o *options) {
		o.maxTokens = maxTokens
	}
}

func WithBaseURL(baseURL string) Option {
	return func(o *options) {
		o.baseURL = baseURL
	}
}

func WithAPIKey(apiKey string) Option {
	return func(o *options) {
		o.apiKey = apiKey
	}
}

func WithMessages(messages minds.Messages) Option {
	return func(o *options) {
		o.messages = messages
	}
}

func WantCompletionResponse(response *openai.ChatCompletionResponse) Option {
	return func(o *options) {
		o.response = response
	}
}

// getDefaultModel returns the default model from environment variables
func getDefaultModel() string {
	model := os.Getenv("LLM_DEFAULT_MODEL")
	if model == "" {
		model = os.Getenv("OPENAI_DEFAULT_MODEL")
		if model == "" {
			model = GPT41Nano
		}
	}
	return model
}

// Ask sends a prompt to an LLM and returns the response.
// It accepts optional WithModel() to specify a model, otherwise uses default.
func Ask(ctx context.Context, prompt string, opts ...Option) (string, error) {
	// Process options only to determine the model
	o := &options{
		model:     getDefaultModel(),
		maxTokens: maxResponseTokens,
	}

	for _, opt := range opts {
		opt(o)
	}

	// Validate options for conflicts
	if err := validateOptions(o); err != nil {
		return "", err
	}

	switch o.model {
	case GPT41Nano, GPT41Mini, GPT41, GPT4oMini, O4Mini, O3Mini:
		return AskOpenAI(ctx, prompt, opts...)

	case DeepSeekChat, DeepSeekReasoner:
		// For DeepSeek models, we need to ensure the base URL is set correctly
		// Create a new options slice with the DeepSeek base URL
		deepSeekOpts := slices.Clone(opts)
		key := os.Getenv("DEEPSEEK_API_KEY")
		deepSeekOpts = append(deepSeekOpts, WithBaseURL(DeepSeekAPIURL), WithAPIKey(key))
		return AskOpenAI(ctx, prompt, deepSeekOpts...)

	case MockModel:
		return MockLLMResponse, nil
	default:
		return "", fmt.Errorf("unknown model: %s", o.model)
	}
}

// AskOpenAI sends a prompt to OpenAI and returns the response.
// It accepts optional WithModel() to specify a model, otherwise uses default.
func AskOpenAI(ctx context.Context, prompt string, opts ...Option) (string, error) {
	// Process options
	o := &options{
		model:     os.Getenv("OPENAI_DEFAULT_MODEL"),
		baseURL:   os.Getenv("OPENAI_BASE_URL"),
		apiKey:    os.Getenv("OPENAI_API_KEY"),
		maxTokens: maxResponseTokens,
	}

	if o.model == "" {
		o.model = GPT41Nano
	}

	if o.baseURL == "" {
		o.baseURL = OpenAIAPIURLv1
	}

	for _, opt := range opts {
		opt(o)
	}

	// Validate options for conflicts
	if err := validateOptions(o); err != nil {
		return "", err
	}

	if o.apiKey == "" {
		return "", fmt.Errorf("AskOpenAI: API key is not set")
	}

	config := openai.DefaultConfig(o.apiKey)
	config.BaseURL = o.baseURL

	var messages []openai.ChatCompletionMessage

	if len(o.messages) > 0 {
		// Use the provided messages
		messages = convertToOpenAIMessages(o.messages)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		})
	} else {
		messages = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		}
	}

	req := &openai.ChatCompletionRequest{
		Model:     o.model,
		Messages:  messages,
		MaxTokens: o.maxTokens,
	}

	client := openai.NewClientWithConfig(config)
	resp, err := client.CreateChatCompletion(ctx, *req)
	if err != nil {
		llm := "OpenAI"

		if IsDeepSeekModel(o.model) {
			llm = "DeepSeek"
		}

		return "", fmt.Errorf("AskOpenAI: %s API: %w. baseURL:%s maxTokens:%d model:%s key:%s", llm, err, o.baseURL, o.maxTokens, o.model, o.apiKey[0:5]+"...") // Mask API key in error message
	}

	if o.response != nil {
		*o.response = resp
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from LLM API")
	}
	return resp.Choices[0].Message.Content, nil
}

// StructuredAsk sends a prompt to an LLM and returns a structured response of type T.
// It accepts optional WithModel() to specify a model, otherwise uses default.
func StructuredAsk[T any](ctx context.Context, prompt string, opts ...Option) (T, error) {
	// Process options only to determine the model
	o := &options{
		model:     getDefaultModel(),
		maxTokens: maxResponseTokens,
	}

	for _, opt := range opts {
		opt(o)
	}

	// Validate options for conflicts
	if err := validateOptions(o); err != nil {
		var zero T
		return zero, err
	}

	var zero T // Zero value to return in error cases
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	name := t.Name()

	switch o.model {
	case GPT41Nano, GPT41Mini, GPT41, GPT4oMini, O4Mini, O3Mini:
		return StructuredAskOpenAI[T](ctx, name, prompt, opts...)
	case DeepSeekChat, DeepSeekReasoner:
		// For DeepSeek models, we need to ensure the base URL is set correctly
		// Create a new options slice with the DeepSeek base URL
		deepSeekOpts := append([]Option{}, opts...)
		deepSeekOpts = append(deepSeekOpts, WithBaseURL(DeepSeekAPIURL), WithAPIKey(os.Getenv("DEEPSEEK_API_KEY")))
		return StructuredAskOpenAI[T](ctx, name, prompt, deepSeekOpts...)
	case MockModel:
		if err := json.Unmarshal([]byte(MockLLMResponse), &zero); err != nil {
			return zero, fmt.Errorf("failed to unmarshal mock response: %w", err)
		}
		return zero, MockLLMError
	default:
		return zero, fmt.Errorf("unknown model: %s", o.model)
	}
}

// StructuredAskOpenAI sends a prompt to OpenAI and returns a structured response of type T.
// It accepts optional WithModel() to specify a model, otherwise uses default.
func StructuredAskOpenAI[T any](ctx context.Context, typeName, prompt string, opts ...Option) (T, error) {
	var zero T // Zero value to return in error cases

	// Process options
	o := &options{
		model:     os.Getenv("OPENAI_DEFAULT_MODEL"),
		baseURL:   os.Getenv("OPENAI_BASE_URL"),
		apiKey:    os.Getenv("OPENAI_API_KEY"),
		maxTokens: maxResponseTokens,
	}

	if o.model == "" {
		o.model = GPT41Nano
	}

	if o.baseURL == "" {
		o.baseURL = OpenAIAPIURLv1
	}

	for _, opt := range opts {
		opt(o)
	}

	// Validate options for conflicts
	if err := validateOptions(o); err != nil {
		return zero, err
	}

	if o.apiKey == "" {
		return zero, fmt.Errorf("OPENAI_API_KEY is not set")
	}

	config := openai.DefaultConfig(o.apiKey)
	config.BaseURL = o.baseURL

	schema, err := GenerateJSONSchema(zero)
	if err != nil {
		return zero, fmt.Errorf("StructuredAskOpenAI: %w", err)
	}

	responseFormat := openai.ChatCompletionResponseFormat{
		Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
		JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
			Name:   typeName,
			Schema: schema,
			Strict: true,
		},
	}
	if o.model == DeepSeekChat || o.model == DeepSeekReasoner {
		responseFormat.Type = openai.ChatCompletionResponseFormatTypeJSONObject
		responseFormat.JSONSchema = nil
	}

	var messages []openai.ChatCompletionMessage

	if len(o.messages) > 0 {
		// Use the provided messages
		messages = convertToOpenAIMessages(o.messages)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		})
	} else {
		// Build messages from the prompt
		messages = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		}
	}

	req := &openai.ChatCompletionRequest{
		Model:          o.model,
		Messages:       messages,
		MaxTokens:      o.maxTokens,
		ResponseFormat: &responseFormat,
	}

	client := openai.NewClientWithConfig(config)
	resp, err := client.CreateChatCompletion(ctx, *req)
	if err != nil {
		llm := "OpenAI"

		if IsDeepSeekModel(o.model) {
			llm = "DeepSeek"
		}

		return zero, fmt.Errorf("StructuredAskOpenAI: %s API: %w. baseURL:%s maxTokens:%d model:%s key:%s", llm, err, o.baseURL, o.maxTokens, o.model, o.apiKey[0:5]+"...") // Mask API key in error message
	}

	var result T
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
		return zero, fmt.Errorf("StructuredAskOpenAI: failed to unmarshal response: %w. content:%s", err, resp.Choices[0].Message.Content)
	}
	return result, nil
}

func GenerateJSONSchema(v any) (*jsonschema.Definition, error) {
	schema, err := minds.GenerateSchema(v)
	if err != nil {
		return nil, fmt.Errorf("GenerateJSONSchema: %w", err)
	}

	return ConvertSchemaDefinition(schema)
}

func ConvertSchemaDefinition(schema *minds.Definition) (*jsonschema.Definition, error) {
	// a minds.Definition and jsonschema.Definition are exactly the same.  just deep copy everything
	// and return it
	//
	// type Definition struct {
	// 	// Type specifies the data type of the schema.
	// 	Type DataType `json:"type,omitempty"`
	// 	// Description is the description of the schema.
	// 	Description string `json:"description,omitempty"`
	// 	// Enum is used to restrict a value to a fixed set of values. It must be an array with at least
	// 	// one element, where each element is unique. You will probably only use this with strings.
	// 	Enum []string `json:"enum,omitempty"`
	// 	// Properties describes the properties of an object, if the schema type is Object.
	// 	Properties map[string]Definition `json:"properties,omitempty"`
	// 	// Required specifies which properties are required, if the schema type is Object.
	// 	Required []string `json:"required,omitempty"`
	// 	// Items specifies which data type an array contains, if the schema type is Array.
	// 	Items *Definition `json:"items,omitempty"`
	// 	// AdditionalProperties is used to control the handling of properties in an object
	// 	// that are not explicitly defined in the properties section of the schema. example:
	// 	// additionalProperties: true
	// 	// additionalProperties: false
	// 	// additionalProperties: Definition{Type: String}
	// 	AdditionalProperties any `json:"additionalProperties,omitempty"`
	// }

	result := jsonschema.Definition{
		Type:                 jsonschema.DataType(schema.Type),
		Description:          schema.Description,
		Enum:                 schema.Enum,
		Required:             schema.Required,
		AdditionalProperties: schema.AdditionalProperties,
	}

	// Deep copy properties map if it exists
	if schema.Properties != nil {
		result.Properties = make(map[string]jsonschema.Definition)
		for key, prop := range schema.Properties {
			propCopy, err := ConvertSchemaDefinition(&prop)
			if err != nil {
				return nil, fmt.Errorf("ConvertSchemaDefinition: %s: %w", key, err)
			}
			result.Properties[key] = *propCopy
		}
	}

	// Deep copy Items if it exists
	if schema.Items != nil {
		itemsCopy, err := ConvertSchemaDefinition(schema.Items)
		if err != nil {
			return nil, fmt.Errorf("ConvertSchemaDefinition: failed to convert items: %w", err)
		}
		result.Items = itemsCopy
	}

	return &result, nil
}

// For testing purposes
func getOptionsFromAskOptions(opts ...Option) *options {
	o := &options{
		model: getDefaultModel(),
	}

	for _, opt := range opts {
		opt(o)
	}

	// If API key wasn't provided in options, try to get from environment
	if o.apiKey == "" {
		o.apiKey = os.Getenv("OPENAI_API_KEY")
	}

	return o
}

// validateOptions checks for conflicting options when WithMessages is used
func validateOptions(_ *options) error {
	// WithMessages is compatible with all other options since it just provides
	// the message thread. Model and maxTokens are still relevant.
	return nil
}

// convertToOpenAIMessages converts minds.Messages to OpenAI chat completion messages
func convertToOpenAIMessages(mindsMessages minds.Messages) []openai.ChatCompletionMessage {
	var messages []openai.ChatCompletionMessage

	for _, msg := range mindsMessages {
		openaiMsg := openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}

		if msg.Name != "" {
			openaiMsg.Name = msg.Name
		}

		if msg.ToolCallID != "" {
			openaiMsg.ToolCallID = msg.ToolCallID
		}

		// Convert ToolCalls if present
		if len(msg.ToolCalls) > 0 {
			openaiToolCalls := make([]openai.ToolCall, len(msg.ToolCalls))
			for i, tc := range msg.ToolCalls {
				openaiToolCalls[i] = openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolType(tc.Type),
					Function: openai.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: string(tc.Function.Parameters),
					},
				}
			}
			openaiMsg.ToolCalls = openaiToolCalls
		}

		messages = append(messages, openaiMsg)
	}

	return messages
}
