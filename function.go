// minds/interfaces.go

package minds

import (
	"context"
	"fmt"
	"reflect"
)

type FunctionCall struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	// call function with arguments in JSON format
	Parameters []byte `json:"parameters,omitempty"`
	Result     []byte `json:"result,omitempty"`
}

type CallableFunc func(context.Context, []byte) ([]byte, error)

// functionWrapper provides a convenient way to wrap Go functions with metadata
type functionWrapper struct {
	name        string
	description string
	argsSchema  Definition
	impl        CallableFunc // The actual function to call
}

func WrapFunction(name, description string, args interface{}, fn CallableFunc) (Tool, error) {
	// Validate that fn is actually a function
	fnType := reflect.TypeOf(args)

	// Generate parameter schema from the function's first parameter
	params, err := GenerateSchema(reflect.New(fnType).Interface())
	if err != nil {
		return nil, err
	}

	if !isValidToolName(name) {
		return nil, fmt.Errorf("invalid function name: `%s`. Must start with a letter or an underscore. Must be alphameric (a-z, A-Z, 0-9), underscores (_), dots (.) or dashes (-), with a maximum length of 64", name)
	}

	return &functionWrapper{
		name:        name,
		description: description,
		argsSchema:  *params,
		impl:        fn,
	}, nil
}

func isValidToolName(name string) bool {
	if len(name) == 0 {
		return false
	}

	if len(name) > 64 {
		return false
	}

	if !isAlphaNumeric(rune(name[0])) {
		return false
	}

	for _, c := range name {
		if !isAlphaNumeric(c) && c != '_' && c != '.' && c != '-' {
			return false
		}
	}

	return true
}

func isAlphaNumeric(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func (f *functionWrapper) Type() string           { return "function" }
func (f *functionWrapper) Name() string           { return f.name }
func (f *functionWrapper) Description() string    { return f.description }
func (f *functionWrapper) Parameters() Definition { return f.argsSchema }

func (f *functionWrapper) Call(ctx context.Context, params []byte) ([]byte, error) {
	return f.impl(ctx, params)
}

func HandleFunctionCalls(ctx context.Context, resp Response, registry ToolRegistry) (Response, error) {

	if calls, ok := resp.ToolCalls(); ok {
		for i, call := range calls {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			fn := call.Function
			f, ok := registry.Lookup(fn.Name)
			if !ok {
				// The tool was not found. Return a string telling the LLM what tools are available
				tools := registry.List()
				names := make([]string, 0, len(tools))
				for _, tool := range tools {
					names = append(names, tool.Name())
				}

				calls[i].Function.Result = []byte(fmt.Sprintf("ERROR: `%s` is not a valid tool name. The available tools are: %s", fn.Name, names))
				continue
			}

			result, err := f.Call(ctx, fn.Parameters)
			if err != nil {
				calls[i].Function.Result = []byte(fmt.Sprintf("ERROR: Tool `%s` failed: %v", fn.Name, err))
				continue
			}

			calls[i].Function.Result = result
		}
		return NewToolCallResponse(calls)
	}

	return resp, nil
}
