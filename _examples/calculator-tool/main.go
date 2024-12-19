package main

import (
	"context"
	"fmt"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/gemini"
	"github.com/chriscow/minds/openai"
	"github.com/chriscow/minds/tools/calculator"
)

const prompt = "calculate 3+7*4"

func main() {
	ctx := context.Background()

	registry := minds.NewToolRegistry()
	calc, _ := calculator.NewCalculator(calculator.Starlark)
	registry.Register(calc)

	resp := withGemini(ctx, registry)
	printOutput(resp)

	resp = withOpenAI(ctx, registry)
	printOutput(resp)

	registry = minds.NewToolRegistry()
	calc, _ = calculator.NewCalculator(calculator.Lua)
	registry.Register(calc)

	resp = withGemini(ctx, registry)
	printOutput(resp)

	resp = withOpenAI(ctx, registry)
	printOutput(resp)
}

func withGemini(ctx context.Context, registry minds.ToolRegistry) minds.Response {
	provider, err := gemini.NewProvider(ctx, gemini.WithToolRegistry(registry))
	if err != nil {
		panic(err)
	}
	defer provider.Close()

	req := minds.Request{
		Messages: minds.Messages{
			{Content: prompt},
		},
	}

	resp, err := provider.GenerateContent(ctx, req)
	if err != nil {
		panic(err)
	}

	return resp
}

func withOpenAI(ctx context.Context, registry minds.ToolRegistry) minds.Response {
	provider, err := openai.NewProvider(openai.WithToolRegistry(registry))
	if err != nil {
		panic(err)
	}
	defer provider.Close()

	req := minds.Request{
		Messages: minds.Messages{
			{Content: prompt},
		},
	}
	resp, err := provider.GenerateContent(ctx, req)
	if err != nil {
		panic(err)
	}

	return resp
}

func printOutput(resp minds.Response) {
	//
	// We should get a function call response
	//
	switch resp.Type() {
	case minds.ResponseTypeText:
		text, _ := resp.Text()
		fmt.Println("Unexpected response:", text)

	case minds.ResponseTypeToolCall:
		calls, _ := resp.ToolCalls()
		for _, call := range calls {
			fn := call.Function
			fmt.Printf("Called %s with args: %v\n", fn.Name, string(fn.Arguments))
			fmt.Printf("Result: %v\n", string(fn.Result))
		}

	default:
		fmt.Println("Unknown response type: %v", resp.Type())
	}
}
