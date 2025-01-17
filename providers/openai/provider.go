package openai

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/chriscow/minds"
	retryablehttp "github.com/hashicorp/go-retryablehttp"

	"github.com/sashabaranov/go-openai"
)

const (
	defaultModel = "gpt-4o-mini"
)

type Provider struct {
	client  *openai.Client
	options Options
}

// NewProvider creates a new OpenAI provider. If no model name is provided, the
// default model is used, currently "gpt-4o-mini". The default model can be
// overridden by setting the OPENAI_DEFAULT_MODEL environment variable.
func NewProvider(opts ...Option) (*Provider, error) {
	options, err := setupOptions(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to set up options: %w", err)
	}

	if options.httpClient == nil {
		client := retryablehttp.NewClient()
		client.RetryMax = 10
		options.httpClient = client.StandardClient()
	}

	config := openai.DefaultConfig(options.apiKey)
	config.HTTPClient = options.httpClient
	if options.baseURL != "" {
		config.BaseURL = options.baseURL
	}

	client := openai.NewClientWithConfig(config)

	p := Provider{
		client:  client,
		options: options,
	}

	return &p, nil
}

func (p *Provider) ModelName() string {
	return p.options.modelName
}

func (p *Provider) Close() {
	// nothing to do
}

func (p *Provider) GenerateContent(ctx context.Context, req minds.Request) (minds.Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	request, err := p.prepareRequest(req)
	if err != nil {
		return nil, err
	}

	raw, err := p.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, err
	}

	calls := make([]minds.ToolCall, 0)
	for _, call := range raw.Choices[0].Message.ToolCalls {
		if call.Type != "function" {
			continue
		}

		calls = append(calls, minds.ToolCall{
			ID: call.ID,
			Function: minds.FunctionCall{
				Name:       call.Function.Name,
				Parameters: []byte(call.Function.Arguments),
			},
		})
	}

	calls, err = minds.HandleFunctionCalls(ctx, calls, p.options.registry)
	if err != nil {
		return nil, err
	}

	return NewResponse(raw, calls)
}

func setupOptions(opts ...Option) (Options, error) {
	options := Options{
		modelName: defaultModel,
		registry:  minds.NewToolRegistry(),
	}
	for _, opt := range opts {
		opt(&options)
	}

	if options.apiKey == "" {
		options.apiKey = os.Getenv("OPENAI_API_KEY")
		if options.apiKey == "" {
			return options, errors.New("OPENAI_API_KEY is not set or passed as an option")
		}
	}

	if os.Getenv("OPENAI_DEFAULT_MODEL") != "" {
		options.modelName = os.Getenv("OPENAI_DEFAULT_MODEL")
	}

	// Register any functions provided in options
	for _, f := range options.tools {
		if err := options.registry.Register(f); err != nil {
			return options, err
		}
	}

	return options, nil
}

func (p *Provider) prepareRequest(req minds.Request) (openai.ChatCompletionRequest, error) {
	// Convert functions to OpenAI format
	tools := make([]openai.Tool, 0)
	for _, f := range p.options.registry.List() {
		schema := f.Parameters()
		tools = append(tools, openai.Tool{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        f.Name(),
				Description: f.Description(),
				Strict:      true,
				Parameters:  schema,
			},
		})
	}

	modelName := p.options.modelName
	if req.Options.ModelName != nil {
		modelName = *req.Options.ModelName
	}

	request := openai.ChatCompletionRequest{
		Model: modelName,
	}

	if p.options.temperature != nil {
		request.Temperature = *p.options.temperature
	}
	if req.Options.Temperature != nil {
		request.Temperature = *req.Options.Temperature
	}

	if p.options.maxOutputTokens != nil {
		request.MaxCompletionTokens = *p.options.maxOutputTokens
	}
	if req.Options.MaxOutputTokens != nil {
		request.MaxCompletionTokens = *req.Options.MaxOutputTokens
	}

	var responseSchema *minds.ResponseSchema
	if p.options.schema != nil {
		responseSchema = p.options.schema
	}
	if req.Options.ResponseSchema != nil {
		responseSchema = req.Options.ResponseSchema
	}

	if responseSchema != nil {
		request.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:        responseSchema.Name,
				Description: responseSchema.Description,
				Schema:      &responseSchema.Definition,
				Strict:      true,
			},
		}
	} else {
		request.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeText,
		}
	}

	if p.options.systemPrompt != nil {
		request.Messages = append([]openai.ChatCompletionMessage{
			{
				Role:    string(minds.RoleSystem),
				Content: *p.options.systemPrompt,
			},
		}, request.Messages...)
	}

	for i, msg := range req.Messages {
		if msg.Role == "" {
			req.Messages[i].Role = minds.RoleUser
		}

		if msg.Role == minds.RoleModel {
			req.Messages[i].Role = minds.RoleAssistant
		}

		switch msg.Role {
		case minds.RoleModel:
			request.Messages[i].Role = string(minds.RoleAssistant)
		}

		calls := make([]openai.ToolCall, len(msg.ToolCalls))
		for _, call := range msg.ToolCalls {

			calls = append(calls, openai.ToolCall{
				ID:       call.ID,
				Type:     "function",
				Function: openai.FunctionCall{Name: call.Function.Name, Arguments: string(call.Function.Parameters)},
			})
		}

		request.Messages = append(request.Messages, openai.ChatCompletionMessage{
			Role:      string(msg.Role),
			Name:      msg.Name,
			Content:   msg.Content,
			ToolCalls: calls,
		})
	}

	if len(tools) > 0 {
		request.Tools = tools
		if req.Options.ToolChoice != "" {
			request.ToolChoice = req.Options.ToolChoice
		} else {
			request.ToolChoice = "auto"
		}
	}

	return request, nil
}
