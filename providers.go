package minds

import "context"

type ContentGenerator interface {
	ModelName() string
	GenerateContent(context.Context, Request) (Response, error)
	Close()
}

type Embedder interface {
	CreateEmbeddings(model string, input []string) ([][]float32, error)
}

type KVStore interface {
	Save(ctx context.Context, key []byte, value []byte) error
	Load(ctx context.Context, key []byte) ([]byte, error)
}

// TokenCounter defines how to count tokens for different models
type TokenCounter interface {
	CountTokens(text string) (int, error)
}
