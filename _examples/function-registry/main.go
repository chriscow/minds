package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/providers/gemini"
	"github.com/chriscow/minds/providers/openai"
	"github.com/fatih/color"
)

const prompt = `Make the room cozy and warm`

var (
	cyan   = color.New(color.FgCyan).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	purple = color.New(color.FgHiMagenta).SprintFunc()
)

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

	req := minds.Request{
		Messages: minds.Messages{{Role: minds.RoleUser, Content: prompt}},
	}

	withGemini(ctx, registry, req)
	withOpenAI(ctx, registry, req)
	withDeepSeek(ctx, registry, req)
}

func withGemini(ctx context.Context, registry minds.ToolRegistry, req minds.Request) {
	provider, err := gemini.NewProvider(ctx, gemini.WithToolRegistry(registry))
	if err != nil {
		panic(err)
	}
	defer provider.Close()

	resp, err := provider.GenerateContent(ctx, req)
	if err != nil {

		panic(err)
	}

	printOutput(cyan("Gemini"), resp)
}

func withOpenAI(ctx context.Context, registry minds.ToolRegistry, req minds.Request) {
	provider, err := openai.NewProvider(openai.WithToolRegistry(registry))
	if err != nil {
		panic(err)
	}
	defer provider.Close()

	resp, err := provider.GenerateContent(ctx, req)
	if err != nil {
		panic(err)
	}

	printOutput(green("OpenAI"), resp)
}

func withDeepSeek(ctx context.Context, registry minds.ToolRegistry, req minds.Request) {
	baseURl := "https://api.deepseek.com"
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	model := "deepseek-chat"
	provider, err := openai.NewProvider(
		openai.WithAPIKey(apiKey),
		openai.WithModel(model),
		openai.WithToolRegistry(registry),
		openai.WithBaseURL(baseURl),
	)
	if err != nil {
		fmt.Printf("[%s] error: %v", purple("DeepSeek"), err)
		return
	}
	defer provider.Close()

	resp, err := provider.GenerateContent(ctx, req)
	if err != nil {
		fmt.Printf("[%s] error: %v", purple("DeepSeek"), err)
		return
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
		fmt.Println("Unexpected response:", text)

	case minds.ResponseTypeToolCall:
		calls, _ := resp.ToolCalls()
		for _, call := range calls {
			fn := call.Function
			fmt.Printf("[%s] Called %s with args: %v\n", name, fn.Name, string(fn.Parameters))
			fmt.Printf("[%s] Result: %v\n", name, string(fn.Result))
		}

	default:
		fmt.Println("[%s] Unknown response type: %v", name, resp.Type())
	}
}
