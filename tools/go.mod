module github.com/chriscow/minds/tools

go 1.21

toolchain go1.22.4

// replace github.com/chriscow/minds => ../

// replace github.com/chriscow/minds/providers/openai => ../providers/openai

require (
	github.com/matryer/is v1.4.1
	github.com/yuin/gopher-lua v1.1.1
	go.starlark.net v0.0.0-20241125201518-c05ff208a98f
)

require (
	github.com/chriscow/minds v0.0.4 // indirect
	github.com/chriscow/minds/providers/openai v0.0.4 // indirect
	github.com/dlclark/regexp2 v1.11.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/sashabaranov/go-openai v1.36.1 // indirect
	github.com/tiktoken-go/tokenizer v0.2.1 // indirect
	golang.org/x/sys v0.24.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
