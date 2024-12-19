package main

import (
	"context"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/providers/gemini"
	"github.com/chriscow/minds/providers/openai"
)

const prompt = `Hello, how are you?`

func main() {
	req := minds.Request{
		Messages: minds.Messages{
			{
				Content: prompt,
			},
		},
	}

	withGemini(req)
	withOpenAI(req)
}

func withGemini(req minds.Request) {
	ctx := context.Background()
	llm, _ := gemini.NewProvider(ctx)
	resp, _ := llm.GenerateContent(ctx, req)
	println(resp.String())
}

func withOpenAI(req minds.Request) {
	ctx := context.Background()
	llm, _ := openai.NewProvider()
	resp, _ := llm.GenerateContent(ctx, req)
	println(resp.String())
}
