package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/chriscow/minds"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

const (
	GPT41Nano = "gpt-4.1-nano" // $0.10 input	$0.025 cached output	$0.40 output
	GPT41Mini = "gpt-4.1-mini" // $0.40 input	$0.10 cached output		$1.60 output
	GPT41     = "gpt-4.1"      // $2.00 input	$0.50 cached output		$8.00 output

	GPT4oMini = "gpt-4o-mini" // $0.15 input	$0.075 cached output	$0.60 output

	O4Mini = "o4-mini" // $1.10 input	$0.275 cached output	$4.40 output
	O3Mini = "o3-mini" // $1.10 input	$0.55 cached output		$4.40 output

	MockModel = "mock-model"

	maxResponseTokens = 1000
)

var MockLLMResponse = "mock-llm-response"
var MockLLMError error = nil

// Ask sends a prompt to an LLM and returns the response. If the
// LLM_DEFAULT_MODEL is not set, it will use GPT41Nano.
func Ask(ctx context.Context, prompt string) (string, error) {
	model := os.Getenv("LLM_DEFAULT_MODEL")
	if model == "" {
		model = os.Getenv("OPENAI_DEFAULT_MODEL")
		if model == "" {
			model = GPT41Nano
		}
	}
	return AskModel(ctx, prompt, model)
}

// AskModel sends a prompt to an LLM and returns the response using a specific model.
// The LLM provider is determined by the model variable.
func AskModel(ctx context.Context, prompt string, model string) (string, error) {
	switch model {
	case GPT41Nano, GPT41Mini, GPT41, GPT4oMini, O4Mini, O3Mini:
		return AskOpenAIModel(ctx, prompt, model)
	case MockModel:
		return MockLLMResponse, nil
	default:
		return "", fmt.Errorf("unknown model: %s", model)
	}
}

func AskOpenAI(ctx context.Context, prompt string) (string, error) {
	model := os.Getenv("OPENAI_DEFAULT_MODEL")
	if model == "" {
		model = GPT41Nano
	}
	return AskOpenAIModel(ctx, prompt, model)
}

func AskOpenAIModel(ctx context.Context, prompt string, model string) (string, error) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		return "", fmt.Errorf("OPENAI_API_KEY is not set")
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     model,
		Messages:  []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: prompt}},
		MaxTokens: maxResponseTokens,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get response from OpenAI API: %w", err)
	}
	return resp.Choices[0].Message.Content, nil
}

func StructuredAsk[T any](ctx context.Context, name, prompt string) (T, error) {
	model := os.Getenv("LLM_DEFAULT_MODEL")
	if model == "" {
		model = os.Getenv("OPENAI_DEFAULT_MODEL")
		if model == "" {
			model = GPT41Nano
		}
	}

	return StructuredAskModel[T](ctx, name, prompt, model)
}

func StructuredAskModel[T any](ctx context.Context, name, prompt, model string) (T, error) {
	var zero T // Zero value to return in error cases
	switch model {
	case GPT41Nano, GPT41Mini, GPT41, GPT4oMini, O4Mini, O3Mini:
		return StructuredAskOpenAIModel[T](ctx, name, prompt, model)
	case MockModel:
		if err := json.Unmarshal([]byte(MockLLMResponse), &zero); err != nil {
			return zero, fmt.Errorf("failed to unmarshal mock response: %w", err)
		}
		return zero, MockLLMError
	default:
		return zero, fmt.Errorf("unknown model: %s", model)
	}
}

func StructuredAskOpenAIModel[T any](ctx context.Context, name, prompt, model string) (T, error) {
	var zero T // Zero value to return in error cases

	if os.Getenv("OPENAI_API_KEY") == "" {
		return zero, fmt.Errorf("OPENAI_API_KEY is not set")
	}

	schema, err := GenerateJSONSchema(zero)
	if err != nil {
		return zero, fmt.Errorf("failed to generate schema: %w", err)
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     model,
		Messages:  []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: prompt}},
		MaxTokens: maxResponseTokens,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   name,
				Schema: schema,
				Strict: true,
			},
		},
	})
	if err != nil {
		return zero, fmt.Errorf("failed to get response from OpenAI API: %w", err)
	}

	var result T
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
		return zero, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return result, nil
}

func GenerateJSONSchema(v any) (*jsonschema.Definition, error) {
	schema, err := minds.GenerateSchema(v)
	if err != nil {
		return nil, fmt.Errorf("failed to generate schema: %w", err)
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
				return nil, fmt.Errorf("failed to convert property %s: %w", key, err)
			}
			result.Properties[key] = *propCopy
		}
	}

	// Deep copy Items if it exists
	if schema.Items != nil {
		itemsCopy, err := ConvertSchemaDefinition(schema.Items)
		if err != nil {
			return nil, fmt.Errorf("failed to convert items: %w", err)
		}
		result.Items = itemsCopy
	}

	return &result, nil
}
