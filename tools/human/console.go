package human

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chriscow/minds"
)

type Input struct {
	Question string `json:"question" description:"The question for the human to answer"`
}

func NewConsoleInput() (minds.Tool, error) {
	return minds.WrapFunction(
		"human",
		`Useful for getting input from a human. This tool can ask a human a question and returns the answer.`,
		Input{},
		human_console,
	)
}

func human_console(_ context.Context, args []byte) ([]byte, error) {
	var params Input
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	var answer string

	fmt.Printf("\n\n%s > ", params.Question)
	fmt.Scanln(&answer)

	return []byte(answer), nil
}
