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
	// Functions need to be wrapped with metadata to turn them into a tool
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

	req := minds.Request{
		Messages: minds.Messages{{Role: minds.RoleUser, Content: prompt}},
	}

	withGemini(ctx, lightControl, req)
	withOpenAI(ctx, lightControl, req)
	withDeepSeek(ctx, lightControl, req)
}

func withGemini(ctx context.Context, fn minds.Tool, req minds.Request) {
	provider, err := gemini.NewProvider(ctx, gemini.WithTool(fn))
	if err != nil {
		fmt.Printf("[%s] error: %v", cyan("Gemini"), err)
		return
	}
	defer provider.Close()

	resp, err := provider.GenerateContent(ctx, req)
	if err != nil {
		fmt.Printf("[%s] error: %v", cyan("Gemini"), err)
		return
	}

	printOutput(cyan("Gemini"), resp)
}

func withOpenAI(ctx context.Context, fn minds.Tool, req minds.Request) {
	provider, err := openai.NewProvider(openai.WithTool(fn))
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

	printOutput(green("OpenAI"), resp)
}

func withDeepSeek(ctx context.Context, fn minds.Tool, req minds.Request) {
	baseURl := "https://api.deepseek.com"
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	model := "deepseek-chat"
	provider, err := openai.NewProvider(openai.WithAPIKey(apiKey), openai.WithModel(model), openai.WithTool(fn), openai.WithBaseURL(baseURl))
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

	for _, call := range resp.ToolCalls() {
		fn := call.Function
		fmt.Printf("[%s] Called %s with args: %v\n", name, fn.Name, string(fn.Parameters))
		fmt.Printf("[%s] Result: %v\n", name, string(fn.Result))
	}
}
