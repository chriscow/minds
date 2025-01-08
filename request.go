package minds

type RequestOptions struct {
	ModelName       *string
	Temperature     *float32
	MaxOutputTokens *int
	ResponseSchema  *ResponseSchema
	ToolRegistry    ToolRegistry
	ToolChoice      string
}

type RequestOption func(*RequestOptions)

type Request struct {
	Options  RequestOptions
	Messages Messages `json:"messages"`
}

func NewRequest(messages Messages, opts ...RequestOption) Request {
	options := RequestOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	return Request{
		Options:  options,
		Messages: messages,
	}
}

func WithModel(model string) RequestOption {
	return func(o *RequestOptions) {
		o.ModelName = &model
	}
}

func WithTemperature(temperature float32) RequestOption {
	return func(o *RequestOptions) {
		o.Temperature = &temperature
	}
}

func WithMaxOutputTokens(tokens int) RequestOption {
	return func(o *RequestOptions) {
		o.MaxOutputTokens = &tokens
	}
}

func WithResponseSchema(schema ResponseSchema) RequestOption {
	return func(o *RequestOptions) {
		o.ResponseSchema = &schema
	}
}

func (r Request) TokenCount(tokenizer TokenCounter) (int, error) {
	total := 0
	for _, msg := range r.Messages {
		count, err := tokenizer.CountTokens(msg.Content)
		if err != nil {
			return 0, err
		}
		total += count
	}

	return total, nil
}
