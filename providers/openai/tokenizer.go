package openai

import (
	"github.com/chriscow/minds"

	"github.com/tiktoken-go/tokenizer"
)

// Tokenizer implements the TokenCounter interface for the OpenAI provider.
type Tokenizer struct {
	codec tokenizer.Codec
}

func NewTokenizer(modelName string) (minds.TokenCounter, error) {
	codec, err := tokenizer.ForModel(tokenizer.Model(modelName))
	if err != nil {
		return nil, err
	}

	return Tokenizer{codec: codec}, nil
}

func (t Tokenizer) CountTokens(text string) (int, error) {
	ids, _, err := t.codec.Encode(text)
	if err != nil {
		return 0, err
	}

	return len(ids), nil
}
