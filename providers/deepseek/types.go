package deepseek

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chriscow/minds"
)

type FinishReason string
type ToolChoice string

const (
	ToolChoiceNone     ToolChoice = "none"
	ToolChoiceAuto     ToolChoice = "auto"
	ToolChoiceRequired ToolChoice = "required"

	FinishReasonStop      FinishReason = "stop"
	FinishReasonLength    FinishReason = "length"
	FinishReasonToolCalls FinishReason = "tool_calls"
)

type ChatCompletionRequest struct {
	Messages         minds.Messages   `json:"messages"`
	Model            string           `json:"model"`
	FrequencyPenalty *int             `json:"frequency_penalty"`
	MaxTokens        *int             `json:"max_tokens"`
	PresencePenalty  *float32         `json:"presence_penalty"`
	ResponseFormat   *ResponseFormat  `json:"response_format"`
	Stop             []string         `json:"stop"`
	Stream           bool             `json:"stream"`
	StreamOptions    *StreamUsage     `json:"stream_options"`
	Temperature      *float32         `json:"temperature"`
	TopP             *float32         `json:"top_p"`
	Tools            []minds.ToolCall `json:"tools"`
	ToolChoice       *ToolChoice      `json:"tool_choice"`
	Logprobs         *bool            `json:"logprobs"`
	TopLogprobs      *int             `json:"top_logprobs"`
}

type ResponseFormat struct {
	Type string `json:"type"` // can be "text" or "json_object" default: "text"
}

type StreamUsage struct {
	IncludeUsage bool `json:"include_usage"`
}

type ChatCompletionResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
	SystemFingerprint string   `json:"system_fingerprint"`
}

type Choice struct {
	Index        int           `json:"index"`
	Message      minds.Message `json:"message"`
	Logprobs     *int          `json:"logprobs"`
	FinishReason FinishReason  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens          int `json:"prompt_tokens"`
	CompletionTokens      int `json:"completion_tokens"`
	TotalTokens           int `json:"total_tokens"`
	PromptCacheHitTokens  int `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens int `json:"prompt_cache_miss_tokens"`
}

func (r ChatCompletionResponse) ToMessages() (minds.Messages, error) {
	switch r.Type() {
	case minds.ResponseTypeText:
		return minds.Messages{
			{
				Role:    minds.Role(r.Choices[0].Message.Role),
				Content: r.Choices[0].Message.Content,
			},
		}, nil
	case minds.ResponseTypeToolCall:
		calls, _ := r.ToolCalls()
		resp, err := minds.NewToolCallResponse(calls)
		if err != nil {
			return nil, fmt.Errorf("failed to create tool call response: %w", err)
		}

		return resp.ToMessages()
	default:
		return nil, fmt.Errorf("unknown response type")
	}
}

func (r ChatCompletionResponse) String() string {
	responseType := r.Type()

	switch responseType {
	case minds.ResponseTypeText:
		text, _ := r.Text()
		return text
	case minds.ResponseTypeToolCall:
		tools, _ := r.ToolCalls()
		names := make([]string, 0, len(tools))
		for _, tool := range tools {
			names = append(names, tool.Function.Name)
		}

		return strings.Join(names, ", ")
	default:
		return "Unknown response type"
	}
}

func (r ChatCompletionResponse) Type() minds.ResponseType {
	if r.Choices[0].FinishReason == FinishReasonToolCalls {
		return minds.ResponseTypeToolCall
	}

	if r.Choices[0].FinishReason == FinishReasonStop || r.Choices[0].FinishReason == FinishReasonLength {
		return minds.ResponseTypeText
	}

	return minds.ResponseTypeUnknown
}

func (r ChatCompletionResponse) Text() (string, bool) {
	if r.Type() != minds.ResponseTypeText {
		return "", false
	}

	return r.Choices[0].Message.Content, true
}

func (r ChatCompletionResponse) ToolCalls() ([]minds.ToolCall, bool) {
	if r.Type() != minds.ResponseTypeToolCall {
		return nil, false
	}

	infos := make([]minds.ToolCall, 0)
	for _, tool := range r.Choices[0].Message.ToolCalls {
		name := tool.Function.Name

		params, err := json.Marshal(tool.Function.Parameters)
		if err != nil {
			return nil, false
		}

		infos = append(infos, minds.ToolCall{
			Function: minds.FunctionCall{
				Name:       name,
				Parameters: params,
			},
		})
	}

	return infos, true
}
