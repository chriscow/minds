package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/gemini"
	"github.com/chriscow/minds/openai"
)

const prompt = "Generate sample data for a person in JSON format"

type SampleData struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// This example demonstrates how to generate structured output from an LLM.
// The structured output is defined by the SampleData struct above.
// The example uses two different LLM providers, Gemini and OpenAI, to generate the same output.
func main() {
	ctx := context.Background()

	resp := withGemini(ctx)
	printResult(resp)

	resp = withOpenAI(ctx)
	printResult(resp)
}

func withGemini(ctx context.Context) minds.Response {
	schema, _ := gemini.GenerateSchema(SampleData{})
	llm, err := gemini.NewProvider(ctx, gemini.WithResponseSchema(schema))
	if err != nil {
		panic(err)
	}

	req := minds.Request{
		Messages: minds.Messages{
			{
				Content: prompt,
			},
		},
	}

	resp, err := llm.GenerateContent(ctx, req)
	if err != nil {
		panic(err)
	}
	return resp
}

func withOpenAI(ctx context.Context) minds.Response {
	schema, _ := minds.NewResponseSchema("SampleData", "some sample data", SampleData{})
	llm, err := openai.NewProvider(openai.WithResponseSchema(*schema))
	if err != nil {
		panic(err)
	}

	req := minds.Request{
		Messages: minds.Messages{
			{
				Content: prompt,
			},
		},
	}

	resp, err := llm.GenerateContent(ctx, req)
	if err != nil {
		panic(err)
	}

	return resp
}

func printResult(resp minds.Response) {
	var result SampleData
	json.Unmarshal([]byte(resp.String()), &result)

	fmt.Printf("Name: %s, Age: %d\n", result.Name, result.Age)
}
