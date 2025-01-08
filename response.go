package minds

// ResponseType indicates what kind of response we received
type ResponseType int

const (
	ResponseTypeUnknown ResponseType = iota
	ResponseTypeText
	ResponseTypeToolCall
)

type ResponseSchema struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Definition  Definition `json:"schema"`
}

func NewResponseSchema(name, desc string, v any) (*ResponseSchema, error) {
	def, err := GenerateSchema(v)
	if err != nil {
		return nil, err
	}

	return &ResponseSchema{
		Name:        name,
		Description: desc,
		Definition:  *def,
	}, nil
}

type Response interface {
	// String returns a string representation of the response
	String() string

	// ToolCall returns the tool call details if this is a tool call response.
	ToolCalls() []ToolCall
}

type ResponseHandler func(resp Response) error

func (h ResponseHandler) HandleResponse(resp Response) error {
	return h(resp)
}
