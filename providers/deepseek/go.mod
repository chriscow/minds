module github.com/chriscow/minds/providers/openai

go 1.18

replace github.com/chriscow/minds => ../../

require (
	github.com/chriscow/minds v0.0.2
	github.com/hashicorp/go-retryablehttp v0.7.7
	github.com/matryer/is v1.4.1
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
