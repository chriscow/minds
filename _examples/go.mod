module github.com/chriscow/minds-examples

go 1.22.7

toolchain go1.23.4

// replace github.com/chriscow/minds => ../

// replace github.com/chriscow/minds/providers/openai => ../providers/openai

// replace github.com/chriscow/minds/providers/gemini => ../providers/gemini

// replace github.com/chriscow/minds/providers/deepseek => ../providers/deepseek

// replace github.com/chriscow/minds/tools => ../tools

require (
	github.com/chriscow/minds v0.0.3
	github.com/chriscow/minds/providers/gemini v0.0.3
	github.com/chriscow/minds/providers/openai v0.0.3
	github.com/chriscow/minds/tools v0.0.1
	github.com/fatih/color v1.18.0
	golang.org/x/time v0.8.0
)

require (
	cloud.google.com/go v0.118.0 // indirect
	cloud.google.com/go/ai v0.9.0 // indirect
	cloud.google.com/go/auth v0.13.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.6 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	cloud.google.com/go/longrunning v0.6.4 // indirect
	github.com/ThreeDotsLabs/watermill v1.4.1 // indirect
	github.com/dlclark/regexp2 v1.11.4 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/generative-ai-go v0.19.0 // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/lithammer/shortuuid/v3 v3.0.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sashabaranov/go-openai v1.36.1 // indirect
	github.com/tiktoken-go/tokenizer v0.2.1 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.58.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.58.0 // indirect
	go.opentelemetry.io/otel v1.33.0 // indirect
	go.opentelemetry.io/otel/metric v1.33.0 // indirect
	go.opentelemetry.io/otel/trace v1.33.0 // indirect
	go.starlark.net v0.0.0-20241125201518-c05ff208a98f // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/oauth2 v0.24.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/api v0.214.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250102185135-69823020774d // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250102185135-69823020774d // indirect
	google.golang.org/grpc v1.69.2 // indirect
	google.golang.org/protobuf v1.36.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
