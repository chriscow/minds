package calculator

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	lua "github.com/yuin/gopher-lua"
)

// Execute runs the calculatorLua tool with the given input.
func withLua(_ context.Context, args []byte) ([]byte, error) {
	var params struct{ Input string }
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	input := params.Input

	L := lua.NewState()
	defer L.Close()

	// Open Lua standard libraries to get access to basic operators
	L.OpenLibs()

	// Register math functions
	mathLib := map[string]lua.LGFunction{
		"floor": func(L *lua.LState) int {
			x := L.CheckNumber(1)
			L.Push(lua.LNumber(math.Floor(float64(x))))
			return 1
		},
		"ceil": func(L *lua.LState) int {
			x := L.CheckNumber(1)
			L.Push(lua.LNumber(math.Ceil(float64(x))))
			return 1
		},
		"round": func(L *lua.LState) int {
			x := L.CheckNumber(1)
			L.Push(lua.LNumber(math.Round(float64(x))))
			return 1
		},
		"sqrt": func(L *lua.LState) int {
			x := L.CheckNumber(1)
			if float64(x) < 0 {
				L.RaiseError("cannot calculate square root of negative number")
				return 0
			}
			L.Push(lua.LNumber(math.Sqrt(float64(x))))
			return 1
		},
		"pow": func(L *lua.LState) int {
			x := L.CheckNumber(1)
			y := L.CheckNumber(2)
			result := math.Pow(float64(x), float64(y))
			if math.IsNaN(result) || math.IsInf(result, 0) {
				L.RaiseError("invalid power operation")
				return 0
			}
			L.Push(lua.LNumber(result))
			return 1
		},
		"sin": func(L *lua.LState) int {
			x := L.CheckNumber(1)
			L.Push(lua.LNumber(math.Sin(float64(x))))
			return 1
		},
		"cos": func(L *lua.LState) int {
			x := L.CheckNumber(1)
			L.Push(lua.LNumber(math.Cos(float64(x))))
			return 1
		},
		"tan": func(L *lua.LState) int {
			x := L.CheckNumber(1)
			L.Push(lua.LNumber(math.Tan(float64(x))))
			return 1
		},
		"asin": func(L *lua.LState) int {
			x := L.CheckNumber(1)
			if float64(x) < -1 || float64(x) > 1 {
				L.RaiseError("asin: argument out of domain")
				return 0
			}
			L.Push(lua.LNumber(math.Asin(float64(x))))
			return 1
		},
		"acos": func(L *lua.LState) int {
			x := L.CheckNumber(1)
			if float64(x) < -1 || float64(x) > 1 {
				L.RaiseError("acos: argument out of domain")
				return 0
			}
			L.Push(lua.LNumber(math.Acos(float64(x))))
			return 1
		},
		"atan": func(L *lua.LState) int {
			x := L.CheckNumber(1)
			L.Push(lua.LNumber(math.Atan(float64(x))))
			return 1
		},
		"atan2": func(L *lua.LState) int {
			y := L.CheckNumber(1)
			x := L.CheckNumber(2)
			L.Push(lua.LNumber(math.Atan2(float64(y), float64(x))))
			return 1
		},
		"abs": func(L *lua.LState) int {
			x := L.CheckNumber(1)
			L.Push(lua.LNumber(math.Abs(float64(x))))
			return 1
		},
		"exp": func(L *lua.LState) int {
			x := L.CheckNumber(1)
			L.Push(lua.LNumber(math.Exp(float64(x))))
			return 1
		},
		"log": func(L *lua.LState) int {
			x := L.CheckNumber(1)
			base := L.OptNumber(2, math.E)
			result := math.Log(float64(x))
			if base != math.E {
				result = result / math.Log(float64(base))
			}
			L.Push(lua.LNumber(result))
			return 1
		},
		"log10": func(L *lua.LState) int {
			x := L.CheckNumber(1)
			L.Push(lua.LNumber(math.Log10(float64(x))))
			return 1
		},
	}

	// Create math table
	mathTab := L.NewTable()
	for name, fn := range mathLib {
		L.SetField(mathTab, name, L.NewFunction(fn))
	}

	// Add constants
	L.SetField(mathTab, "pi", lua.LNumber(math.Pi))
	L.SetField(mathTab, "e", lua.LNumber(math.E))

	// Register the math library
	L.SetGlobal("math", mathTab)

	wrappedInput := fmt.Sprintf("return %s", input)

	if err := L.DoString(wrappedInput); err != nil {
		return nil, fmt.Errorf("error from evaluator: %s", err.Error())
	}

	// Get the result from the stack
	if L.GetTop() > 0 {
		result := L.Get(-1)
		L.Pop(1)
		return []byte(result.String()), nil
	}

	return nil, fmt.Errorf("no result value found in script output")
}
