package openai

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/chriscow/minds"
	"github.com/chriscow/minds/handlers"
	"github.com/matryer/is"
)

// TestSummarize tests the Summerize handler using the OpenAI provider. This
// test lives in the openai module to avoid dragging in dependencies into the
// minds module.
func TestSummarize(t *testing.T) {
	is := is.New(t)

	fmt.Println(os.Getenv("OPENAI_API_KEY"))
	apiToken := os.Getenv("OPENAI_API_KEY")
	if apiToken == "" {
		t.Skip("Skipping integration test against OpenAI API. Set OPENAI_API_KEY environment variable to enable it.")
	}

	llm, err := NewProvider(WithAPIKey(apiToken))
	is.NoErr(err)

	systemMsg := "you are a helpful summerization assistant"

	summerizer := handlers.Summerize(llm, systemMsg)
	tc := minds.NewThreadContext(context.Background()).WithMessages(minds.Messages{
		{Role: minds.RoleSystem, Content: systemMsg},
		{Role: minds.RoleUser, Content: "What is the meaning of life?"},
		{Role: minds.RoleAssistant, Content: `
The meaning of life is a deeply personal and philosophical question that has been explored for centuries by thinkers, religions, and individuals. Some common perspectives include:

- **Philosophical**: To seek knowledge, understanding, or personal fulfillment.
- **Religious/Spiritual**: To serve a higher purpose, connect with the divine, or achieve spiritual enlightenment.
- **Biological**: To survive and propagate the species.
- **Existential**: To create meaning through personal choices and actions in an otherwise neutral universe.
- **Hedonistic**: To pursue happiness and minimize suffering.

Ultimately, the meaning of life is what you define it to be, based on your beliefs, values, and experiences. What feels meaningful to you?`,
		},
	})

	result, err := summerizer.HandleThread(tc, nil)
	is.NoErr(err)
	msgOut := result.Messages()

	is.True(len(msgOut) == 1)
	is.True(msgOut[0].Role == minds.RoleSystem)
	is.True(len(msgOut[0].Content) > len(systemMsg))
}
