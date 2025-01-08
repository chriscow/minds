module github.com/chriscow/minds/providers/openai

go 1.18

replace github.com/chriscow/minds => ../../

require (
	github.com/chriscow/minds v0.0.0-00010101000000-000000000000
	github.com/matryer/is v1.4.1
	github.com/sashabaranov/go-openai v1.36.0
	github.com/tiktoken-go/tokenizer v0.2.1
)

require (
	github.com/dlclark/regexp2 v1.9.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
