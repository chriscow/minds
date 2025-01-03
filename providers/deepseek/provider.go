package deepseek

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/chriscow/minds"
	"github.com/hashicorp/go-retryablehttp"
)

type Provider struct {
	client  *retryablehttp.Client
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

	// config := openai.DefaultConfig(options.apiKey)
	if options.baseURL == "" {
		options.baseURL = baseUrl
	}

	// client := openai.NewClientWithConfig(config)

	p := Provider{
		client:  options.client,
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

	resp, err := p.createChatCompletion(ctx, request)
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
		options.apiKey = os.Getenv(keyEnvVar)
		if options.apiKey == "" {
			return options, errors.New(keyEnvVar + " is not set or passed as an option")
		}
	}

	if options.client == nil {
		options.client = retryablehttp.NewClient()
	}

	// Register any functions provided in options
	for _, f := range options.tools {
		if err := options.registry.Register(f); err != nil {
			return options, err
		}
	}

	return options, nil
}

func (p *Provider) prepareRequest(req minds.Request) (ChatCompletionRequest, error) {
	modelName := p.options.modelName
	if req.Options.ModelName != nil {
		modelName = *req.Options.ModelName
	}

	request := ChatCompletionRequest{
		Model: modelName,
	}

	if p.options.temperature != nil {
		request.Temperature = p.options.temperature
	}
	if req.Options.Temperature != nil {
		request.Temperature = req.Options.Temperature
	}

	if p.options.maxOutputTokens != nil {
		request.MaxTokens = p.options.maxOutputTokens
	}
	if req.Options.MaxOutputTokens != nil {
		request.MaxTokens = req.Options.MaxOutputTokens
	}

	var responseSchema *minds.ResponseSchema
	if p.options.schema != nil {
		responseSchema = p.options.schema
	}
	if req.Options.ResponseSchema != nil {
		responseSchema = req.Options.ResponseSchema
	}

	var tools []minds.Tool
	if req.Options.ToolRegistry != nil {
		tools = req.Options.ToolRegistry.List()
	} else {
		tools = p.options.registry.List()
	}

	toolCalls := make([]minds.ToolCall, 0)
	for _, f := range tools {
		schema := f.Parameters()
		params, err := schema.MarshalJSON()
		if err != nil {
			return ChatCompletionRequest{}, nil
		}
		toolCalls = append(toolCalls, minds.ToolCall{
			Type: "function",
			Function: minds.FunctionCall{
				Name:        f.Name(),
				Description: f.Description(),
				Parameters:  params,
			},
		})
	}
	request.Tools = toolCalls

	if responseSchema != nil {
		request.ResponseFormat = &ResponseFormat{"json_object"}
		messages := request.Messages.Copy()
		last := messages.Last()
		schemaJSON, err := responseSchema.Definition.MarshalJSON()
		if err != nil {
			return ChatCompletionRequest{}, fmt.Errorf("failed to marshal response schema: %w", err)
		}
		if last != nil {
			last.Content += "Respond using the following JSON schema:\n\n" + string(schemaJSON)
		}
		request.Messages = messages
	} else {
		request.ResponseFormat = &ResponseFormat{"text"}
	}

	request.Messages = req.Messages.Copy()

	return request, nil
}

func (p *Provider) createChatCompletion(ctx context.Context, ccr ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if len(ccr.Messages) == 0 {
		return nil, errors.New("no messages provided")
	}

	payload, err := json.Marshal(ccr)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	println(string(payload))

	req, err := retryablehttp.NewRequestWithContext(ctx, "POST", p.options.baseURL, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+p.options.apiKey)

	res, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	println(string(body))

	var response ChatCompletionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}
