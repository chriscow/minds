package minds

import (
	"context"
	"fmt"
)

type ToolType string

const (
	ToolTypeFunction ToolType = "function"
)

type ToolCall struct {
	ID       string       `json:"id,omitempty"`
	Type     string       `json:"type,omitempty"`
	Function FunctionCall `json:"function,omitempty"`
}

// Tool is an interface for a tool that can be executed by an LLM. It is similar
// to a function in that it takes input and produces output, but it can be more
// complex than a simple function and doesn't require a wrapper.
type Tool interface {
	Type() string
	Name() string
	Description() string
	Parameters() Definition
	Call(context.Context, []byte) ([]byte, error)
}

type ToolRegistry interface {
	// Register adds a new function to the registry
	Register(t Tool) error
	// Lookup retrieves a function by name
	Lookup(name string) (Tool, bool)
	// List returns all registered functions
	List() []Tool
}

func NewToolRegistry() ToolRegistry {
	return &toolRegistry{
		tools: make(map[string]Tool),
	}
}

type toolRegistry struct {
	tools map[string]Tool
}

func (t *toolRegistry) Register(tool Tool) error {
	if _, exists := t.tools[tool.Name()]; exists {
		return fmt.Errorf("tool %s already registered", tool.Name())
	}
	t.tools[tool.Name()] = tool
	return nil
}

func (t *toolRegistry) Lookup(name string) (Tool, bool) {
	tool, ok := t.tools[name]
	return tool, ok
}

func (t *toolRegistry) List() []Tool {
	tools := make([]Tool, 0, len(t.tools))
	for _, tool := range t.tools {
		tools = append(tools, tool)
	}
	return tools
}
