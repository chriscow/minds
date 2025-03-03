package summarizer

type Option func(*Options)

type Options struct {
	Prompt string
}

func WithPrompt(prompt string) Option {
	return func(o *Options) {
		o.Prompt = prompt
	}
}
