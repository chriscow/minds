package calculator

import (
	"fmt"

	"github.com/chriscow/minds"
)

type Syntax string

const (
	Starlark Syntax = "starlark"
	Lua      Syntax = "lua"
)

func NewCalculator(syntax Syntax) (minds.Tool, error) {
	var fn minds.CallableFunc
	var runtime string
	switch syntax {
	case Starlark:
		fn = withStarlark
		runtime = "starlark"
	case Lua:
		fn = withLua
		runtime = "lua"
	default:
		return nil, fmt.Errorf("unsupported syntax: %s", syntax)
	}

	return minds.WrapFunction(
		"calculator",
		fmt.Sprintf(`Useful for getting the result of a math expression. The input to this 
		tool should be a valid mathematical expression that could be executed by 
		a %s evaluator.`, runtime),
		struct {
			Input string `json:"input" description:"The mathematical expression to evaluate"`
		}{},
		fn,
	)
}
