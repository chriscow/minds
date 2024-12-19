package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/chriscow/minds"
	"google.golang.org/api/option"

	"github.com/google/generative-ai-go/genai"
)

const defaultModel = "gemini-1.5-flash"

// const OpenAICompatURL = "https://generativelanguage.googleapis.com/v1beta/openai/"

type Provider struct {
	client  *genai.Client
	options Options
}

// NewProvider creates a new Gemini provider. If no model name is provided, the
// default model is used, currently "gemini-1.5-flash". The default model can be
// overridden by setting the GEMINI_DEFAULT_MODEL environment variable.
func NewProvider(ctx context.Context, opts ...Option) (*Provider, error) {
	options := Options{
		modelName: defaultModel,
		registry:  minds.NewToolRegistry(),
	}

	if os.Getenv("GEMINI_DEFAULT_MODEL") != "" {
		options.modelName = os.Getenv("GEMINI_DEFAULT_MODEL")
	}

	for _, opt := range opts {
		opt(&options)
	}

	if options.apiKey == "" {
		options.apiKey = os.Getenv("GEMINI_API_KEY")
		if options.apiKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY is not set or passed as an option")
		}
	}

	goptions := []option.ClientOption{
		option.WithAPIKey(options.apiKey),
	}
	if options.baseURL != "" {
		goptions = append(goptions, option.WithEndpoint(options.baseURL))
	}

	client, err := genai.NewClient(ctx, goptions...)
	if err != nil {
		return nil, err
	}

	p := Provider{
		client:  client,
		options: options,
	}

	return &p, nil
}

func (p *Provider) Close() {
	p.client.Close()
}

func (p *Provider) ModelName() string {
	return p.options.modelName
}

func (p *Provider) GenerateContent(ctx context.Context, req minds.Request) (minds.Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	model, err := p.getModel()
	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	// TODO: Gemini is not generating the model on the fly
	// The model is created when the client is created
	cs := model.StartChat()

	var sysPrompt *genai.Content
	history := []*genai.Content{}

	if req.Options.ResponseSchema != nil {
		schema, err := convertSchema(req.Options.ResponseSchema.Definition)
		if err != nil {
			return nil, fmt.Errorf("failed to convert schema: %w", err)
		}

		model.ResponseMIMEType = "application/json"
		model.ResponseSchema = schema
	}

	for _, msg := range req.Messages {
		if msg.Role == minds.RoleSystem {
			if sysPrompt == nil {
				sysPrompt = &genai.Content{Parts: []genai.Part{}, Role: "system"}
			}
			part := genai.Text(msg.Content)
			sysPrompt.Parts = append(sysPrompt.Parts, part)
		} else if msg.Role == minds.RoleFunction {
			response := make(map[string]any)
			if err := json.Unmarshal([]byte("test"), &response); err != nil {
				return nil, fmt.Errorf("failed to unmarshal function response: %w", err)
			}
			history = append(history, &genai.Content{
				Parts: []genai.Part{
					genai.FunctionResponse{
						Name:     msg.Name,
						Response: response,
					},
				},
			})
		} else {
			if msg.Role == "" {
				msg.Role = minds.RoleUser
			}
			if msg.Role == minds.RoleAssistant {
				msg.Role = minds.RoleModel
			}
			history = append(history, &genai.Content{
				Parts: []genai.Part{
					genai.Text(msg.Content),
				},
				Role: string(msg.Role),
			})
		}
	}

	if sysPrompt != nil {
		// The Gemini provider typically sets the system prompt with a construction
		// option but to be compatible with OpenAI, we allow the system prompt to be
		// set in the request as well.
		model.SystemInstruction = sysPrompt
	}

	prompt := history[len(history)-1].Parts // The prompt is the last message
	history = history[:len(history)-1]

	cs.History = history

	raw, err := cs.SendMessage(ctx, prompt...)
	if err != nil {
		return nil, err
	}

	resp, err := NewResponse(raw)
	if err != nil {
		return nil, err
	}

	return minds.HandleFunctionCalls(ctx, resp, p.options.registry)
}

func (p *Provider) getModel() (*genai.GenerativeModel, error) {
	model := p.client.GenerativeModel(p.options.modelName)
	model.Temperature = p.options.temperature
	model.MaxOutputTokens = p.options.maxOutputTokens

	if p.options.schema != nil {
		model.ResponseMIMEType = "application/json"
		model.ResponseSchema = p.options.schema
	}

	if p.options.systemPrompt != nil {
		model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(*p.options.systemPrompt)}, Role: "system"}
	}

	// Register any functions provided in options
	for _, f := range p.options.tools {
		if err := p.options.registry.Register(f); err != nil {
			return nil, err
		}
	}

	// Convert functions to Gemini format
	tools := make([]*genai.FunctionDeclaration, 0)
	for _, f := range p.options.registry.List() {
		schema, err := convertSchema(f.Parameters())
		if err != nil {
			return nil, err
		}

		tools = append(tools, &genai.FunctionDeclaration{
			Name:        f.Name(),
			Description: f.Description(),
			Parameters:  schema,
		})
	}

	if len(tools) > 0 {
		model.Tools = []*genai.Tool{{
			FunctionDeclarations: tools,
		}}
	}

	return model, nil
}
