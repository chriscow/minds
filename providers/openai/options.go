package openai

import "github.com/chriscow/minds"

type Options struct {
	apiKey          string
	baseURL         string
	modelName       string
	temperature     *float32
	maxOutputTokens *int
	schema          *minds.ResponseSchema
	tools           []minds.Tool
	registry        minds.ToolRegistry
	systemPrompt    *string
}

type Option func(*Options)

func WithAPIKey(key string) Option {
	return func(o *Options) {
		o.apiKey = key
	}
}

func WithBaseURL(url string) Option {
	return func(o *Options) {
		o.baseURL = url
	}
}

func WithModel(model string) Option {
	return func(o *Options) {
		o.modelName = model
	}
}

func WithTemperature(temperature float32) Option {
	return func(o *Options) {
		o.temperature = &temperature
	}
}

func WithMaxOutputTokens(tokens int) Option {
	return func(o *Options) {
		o.maxOutputTokens = &tokens
	}
}

func WithResponseSchema(schema minds.ResponseSchema) Option {
	return func(o *Options) {
		o.schema = &schema
	}
}

func WithTool(fn minds.Tool) Option {
	return func(o *Options) {
		if o.tools == nil {
			o.tools = make([]minds.Tool, 0)
		}

		o.tools = append(o.tools, fn)
	}
}

func WithToolRegistry(registry minds.ToolRegistry) Option {
	return func(o *Options) {
		if o.registry != nil && len(o.registry.List()) > 0 {
			panic("cannot set registry when functions are present in existing registry")
		}

		o.registry = registry
	}
}

func WithSystemPrompt(prompt string) Option {
	return func(o *Options) {
		o.systemPrompt = &prompt
	}
}
