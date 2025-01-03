package openai

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/chriscow/minds"

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

	config := openai.DefaultConfig(options.apiKey)
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

	resp, err := NewResponse(raw)
	if err != nil {
		return nil, err
	}

	return minds.HandleFunctionCalls(ctx, resp, p.options.registry)
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

	for _, msg := range req.Messages {
		if msg.Role == "" {
			msg.Role = minds.RoleUser
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
		request.ToolChoice = "auto"
	}

	return request, nil
}
