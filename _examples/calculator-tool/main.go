package main

import (
	"context"
	"fmt"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/providers/deepseek"
	"github.com/chriscow/minds/providers/gemini"
	"github.com/chriscow/minds/providers/openai"
	"github.com/chriscow/minds/tools/calculator"
	"github.com/fatih/color"
)

const prompt = "calculate 3+7*4"

var (
	cyan   = color.New(color.FgCyan).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	purple = color.New(color.FgHiMagenta).SprintFunc()
)

func main() {
	ctx := context.Background()

	registry := minds.NewToolRegistry()
	calc, _ := calculator.NewCalculator(calculator.Starlark)
	registry.Register(calc)

	// withGemini(ctx, registry)
	// withOpenAI(ctx, registry)
	withDeepSeek(ctx, registry)

	registry = minds.NewToolRegistry()
	calc, _ = calculator.NewCalculator(calculator.Lua)
	registry.Register(calc)

	// withGemini(ctx, registry)
	// withOpenAI(ctx, registry)
	withDeepSeek(ctx, registry)
}

func withGemini(ctx context.Context, registry minds.ToolRegistry) {
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

	printOutput(cyan("Gemini"), resp)
}

func withOpenAI(ctx context.Context, registry minds.ToolRegistry) {
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

	printOutput(green("OpenAI"), resp)
}

func withDeepSeek(ctx context.Context, registry minds.ToolRegistry) {
	provider, err := deepseek.NewProvider(deepseek.WithToolRegistry(registry))
	if err != nil {
		panic(err)
	}
	defer provider.Close()

	req := minds.Request{
		Messages: minds.Messages{
			{Role: minds.RoleUser, Content: prompt},
		},
	}
	resp, err := provider.GenerateContent(ctx, req)
	if err != nil {
		panic(err)
	}

	printOutput(purple("DeepSeek"), resp)
}

func printOutput(name string, resp minds.Response) {
	//
	// We should get a function call response
	//
	switch resp.Type() {
	case minds.ResponseTypeText:
		text, _ := resp.Text()
		fmt.Printf("[%s] Unexpected response: %s\n", name, text)

	case minds.ResponseTypeToolCall:
		calls, _ := resp.ToolCalls()
		for _, call := range calls {
			fn := call.Function
			fmt.Printf("[%s] Called %s with args: %v\n", name, fn.Name, string(fn.Parameters))
			fmt.Printf("[%s] Result: %v\n", name, string(fn.Result))
		}

	default:
		fmt.Printf("[%s] Unknown response type: %v\n", name, resp.Type())
	}
}
