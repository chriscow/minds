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

const prompt = "Generate sample data for a person in JSON format"

var (
	cyan   = color.New(color.FgCyan).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	purple = color.New(color.FgHiMagenta).SprintFunc()
)

type SampleData struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// This example demonstrates how to generate structured output from an LLM.
// The structured output is defined by the SampleData struct above.
// The example uses two different LLM providers, Gemini and OpenAI, to generate the same output.
func main() {
	ctx := context.Background()
	req := minds.Request{Messages: minds.Messages{{Role: minds.RoleUser, Content: prompt}}}
	withGemini(ctx, req)
	withOpenAI(ctx, req)
	withDeepSeek(ctx, req)
}

func withGemini(ctx context.Context, req minds.Request) {
	schema, _ := gemini.GenerateSchema(SampleData{})
	llm, err := gemini.NewProvider(ctx, gemini.WithResponseSchema(schema))
	if err != nil {
		fmt.Printf("[%s] error: %v", cyan("Gemini"), err)
		return
	}

	resp, err := llm.GenerateContent(ctx, req)
	if err != nil {
		fmt.Printf("[%s] error: %v", cyan("Gemini"), err)
		return
	}
	printResult(cyan("Gemini"), resp)
}

func withOpenAI(ctx context.Context, req minds.Request) {
	schema, _ := minds.NewResponseSchema("SampleData", "some sample data", SampleData{})
	llm, err := openai.NewProvider(openai.WithResponseSchema(*schema))
	if err != nil {
		panic(err)
	}

	resp, err := llm.GenerateContent(ctx, req)
	if err != nil {
		fmt.Printf("[%s] error: %v", green("OpenAI"), err)
		return
	}

	printResult(green("OpenAI"), resp)
}

func withDeepSeek(ctx context.Context, req minds.Request) {
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Printf("[%s] %s\n", purple("DeepSeek"), yellow("WARNING: DeepSeek does not support structured output in this example"))

	baseURl := "https://api.deepseek.com"
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	model := "deepseek-chat"
	schema, _ := minds.NewResponseSchema("SampleData", "some sample data", SampleData{})

	llm, err := openai.NewProvider(
		openai.WithAPIKey(apiKey),
		openai.WithModel(model),
		openai.WithBaseURL(baseURl),
		openai.WithResponseSchema(*schema),
	)
	if err != nil {
		fmt.Printf("[%s] error: %v", purple("DeepSeek"), err)
		return
	}

	resp, err := llm.GenerateContent(ctx, req)
	if err != nil {
		fmt.Printf("[%s] error: %v", purple("DeepSeek"), err)
		return
	}

	printResult("DeepSeek", resp)
}

func printResult(name string, resp minds.Response) {
	var result SampleData
	json.Unmarshal([]byte(resp.String()), &result)

	fmt.Printf("[%s] Name: %s, Age: %d\n", name, result.Name, result.Age)
}
