package main

import (
	"context"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/providers/deepseek"
	"github.com/chriscow/minds/providers/gemini"
	"github.com/chriscow/minds/providers/openai"
)

const prompt = `Hello, how are you?`

func main() {
	req := minds.Request{
		Messages: minds.Messages{
			{Content: prompt},
		},
	}

	withGemini(req)
	withOpenAI(req)
	withDeepSeek(req)
}

func withGemini(req minds.Request) {
	ctx := context.Background()
	llm, _ := gemini.NewProvider(ctx)
	resp, _ := llm.GenerateContent(ctx, req)
	println("Gemini: " + resp.String())
}

func withOpenAI(req minds.Request) {
	ctx := context.Background()
	llm, _ := openai.NewProvider()
	resp, _ := llm.GenerateContent(ctx, req)
	println("OpenAI: " + resp.String())
}

func withDeepSeek(req minds.Request) {
	ctx := context.Background()
	llm, _ := deepseek.NewProvider()
	resp, _ := llm.GenerateContent(ctx, req)
	println("DeepSeek: " + resp.String())
}
