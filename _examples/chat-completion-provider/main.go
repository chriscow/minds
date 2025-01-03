package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/providers/gemini"
	"github.com/chriscow/minds/providers/openai"
	"github.com/fatih/color"
)

const prompt = `Hello, how are you?`

var (
	cyan   = color.New(color.FgCyan).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	purple = color.New(color.FgHiMagenta).SprintFunc()
)

func main() {
	req := minds.Request{Messages: minds.Messages{{Role: minds.RoleUser, Content: prompt}}}
	ctx := context.Background()
	withGemini(ctx, req)
	withOpenAI(ctx, req)
	withDeepSeek(ctx, req)
}

func withGemini(ctx context.Context, req minds.Request) {
	llm, err := gemini.NewProvider(ctx)
	if err != nil {
		fmt.Printf("[%s] error: %v", cyan("Gemini"), err)
		return
	}
	resp, _ := llm.GenerateContent(ctx, req)
	fmt.Printf("[%s] %s", cyan("Gemini"), resp.String())
}

func withOpenAI(ctx context.Context, req minds.Request) {
	llm, err := openai.NewProvider()
	if err != nil {
		fmt.Printf("[%s] error: %v", green("OpenAI"), err)
		return
	}
	resp, _ := llm.GenerateContent(ctx, req)
	fmt.Printf("[%s] %s\n", green("OpenAI"), resp.String())
}

func withDeepSeek(ctx context.Context, req minds.Request) {
	baseURl := "https://api.deepseek.com"
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	model := "deepseek-chat"
	llm, err := openai.NewProvider(
		openai.WithAPIKey(apiKey),
		openai.WithModel(model),
		openai.WithBaseURL(baseURl),
	)
	if err != nil {
		fmt.Printf("[%s] error: %v", purple("DeepSeek"), err)
		return
	}
	resp, _ := llm.GenerateContent(ctx, req)
	fmt.Printf("[%s] %s\n", purple("DeepSeek"), resp.String())
}
