package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/providers/gemini"
	"github.com/chriscow/minds/providers/openai"
)

const prompt = `Make the room cozy and warm`

// Function calling requires a struct to define the parameters
type LightControlParams struct {
	Brightness int    `json:"brightness" description:"Light level from 0 to 100"`
	ColorTemp  string `json:"colorTemp" description:"Color temperature (daylight/cool/warm)"`
}

func controlLight(_ context.Context, args []byte) ([]byte, error) {
	var params LightControlParams
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	result := map[string]any{
		"brightness": params.Brightness,
		"colorTemp":  params.ColorTemp,
	}

	return json.Marshal(result)
}

func main() {
	ctx := context.Background()

	//
	// Functions need to be wrapped with metadata
	//
	lightControl, err := minds.WrapFunction(
		"control_light", // Google recommends using snake_case for function names with Gemini
		"Set the brightness and color temperature of a room light",
		LightControlParams{},
		controlLight,
	)
	if err != nil {
		panic(err)
	}

	//
	// You can optionally create a function registry to manage multiple
	// functions and pass it to the provider. This gives you the flexibility to
	// use the same functions across different providers or to provide your own
	// registry implementation.
	//
	// The providers basically do this same thing by default if you don't
	// provide a registry.
	//
	registry := minds.NewToolRegistry()
	registry.Register(lightControl)

	resp := withGemini(ctx, registry)
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
			fmt.Printf("Result: %v\n", string(call.Function.Result))
		}

	default:
		fmt.Println("Unknown response type: %v", resp.Type())
	}
}
