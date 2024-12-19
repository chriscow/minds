package calculator

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"
)

const (
	calculatorInputTopic = "calculator.tools.minds.thoughtnet.cloud"
)

// Execute runs the calculator tool with the given input.
func withStarlark(_ context.Context, args []byte) ([]byte, error) {
	var params struct{ Input string }
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	input := params.Input

	mathModule := &starlarkstruct.Module{
		Name: "math",
		Members: starlark.StringDict{
			// rounding functions
			"floor": starlark.NewBuiltin("floor", floor),
			"ceil":  starlark.NewBuiltin("ceil", ceil),
			"round": starlark.NewBuiltin("round", round),

			// arithmetic functions
			"sqrt": starlark.NewBuiltin("sqrt", sqrt),
			"pow":  starlark.NewBuiltin("pow", pow),

			// trigonometric functions
			"sin":   starlark.NewBuiltin("sin", sin),
			"cos":   starlark.NewBuiltin("cos", cos),
			"tan":   starlark.NewBuiltin("tan", tan),
			"asin":  starlark.NewBuiltin("asin", asin),
			"acos":  starlark.NewBuiltin("acos", acos),
			"atan":  starlark.NewBuiltin("atan", atan),
			"atan2": starlark.NewBuiltin("atan2", atan2),
			// constants
			"pi": starlark.Float(math.Pi),
			"e":  starlark.Float(math.E),

			// abs, exp, log, log10
			"abs":   starlark.NewBuiltin("abs", abs),
			"exp":   starlark.NewBuiltin("exp", exp),
			"log":   starlark.NewBuiltin("log", log),
			"log10": starlark.NewBuiltin("log10", log10),
		},
	}

	thread := &starlark.Thread{Name: "main"}
	globals := starlark.StringDict{
		"math": mathModule,
	}

	// Wrap the input in a return expression if it's not an assignment
	wrappedInput := "_ = " + input

	opt := syntax.FileOptions{}
	globals, err := starlark.ExecFileOptions(&opt, thread, "script.star", wrappedInput, globals)
	if err != nil {
		return nil, fmt.Errorf("error from evaluator: %s", err.Error())
	}

	// Get the "_" value from globals
	if result, ok := globals["_"]; ok {
		return []byte(result.String()), nil
	}

	return nil, fmt.Errorf("no result value found in script output")
}

func floor(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &x); err != nil {
		return nil, err
	}

	var val float64
	switch v := x.(type) {
	case starlark.Int:
		return x, nil // integers don't need flooring
	case starlark.Float:
		val = float64(v)
	default:
		return nil, fmt.Errorf("floor: got %s, want number", x.Type())
	}

	return starlark.Float(math.Floor(val)), nil
}

func ceil(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &x); err != nil {
		return nil, err
	}

	var val float64
	switch v := x.(type) {
	case starlark.Int:
		return x, nil // integers don't need ceiling
	case starlark.Float:
		val = float64(v)
	default:
		return nil, fmt.Errorf("ceil: got %s, want number", x.Type())
	}

	return starlark.Float(math.Ceil(val)), nil
}

func round(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &x); err != nil {
		return nil, err
	}

	var val float64
	switch v := x.(type) {
	case starlark.Int:
		return x, nil // integers don't need rounding
	case starlark.Float:
		val = float64(v)
	default:
		return nil, fmt.Errorf("round: got %s, want number", x.Type())
	}

	return starlark.Float(math.Round(val)), nil
}

// sqrt is a Starlark built-in function that calculates the square root of a number
func sqrt(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &x); err != nil {
		return nil, err
	}

	var val float64
	switch v := x.(type) {
	case starlark.Int:
		i, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("integer value is too large")
		}
		if i < 0 {
			return nil, fmt.Errorf("cannot calculate square root of negative number")
		}
		val = float64(i)
	case starlark.Float:
		val = float64(v)
		if val < 0 {
			return nil, fmt.Errorf("cannot calculate square root of negative number")
		}
	default:
		return nil, fmt.Errorf("sqrt: got %s, want number", x.Type())
	}

	return starlark.Float(math.Sqrt(val)), nil
}

func pow(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x, y starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &x, "y", &y); err != nil {
		return nil, err
	}

	// Convert first argument to float64
	var xVal float64
	switch v := x.(type) {
	case starlark.Int:
		i, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("base value is too large")
		}
		xVal = float64(i)
	case starlark.Float:
		xVal = float64(v)
	default:
		return nil, fmt.Errorf("pow: got %s for x, want number", x.Type())
	}

	// Convert second argument to float64
	var yVal float64
	switch v := y.(type) {
	case starlark.Int:
		i, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("exponent value is too large")
		}
		yVal = float64(i)
	case starlark.Float:
		yVal = float64(v)
	default:
		return nil, fmt.Errorf("pow: got %s for y, want number", y.Type())
	}

	// Handle special cases and errors
	if xVal == 0 && yVal < 0 {
		return nil, fmt.Errorf("cannot raise zero to a negative power")
	}

	result := math.Pow(xVal, yVal)

	// Check for special cases in the result
	if math.IsNaN(result) {
		return nil, fmt.Errorf("result is not a number")
	}
	if math.IsInf(result, 0) {
		return nil, fmt.Errorf("result is infinite")
	}

	return starlark.Float(result), nil
}

func sin(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &x); err != nil {
		return nil, err
	}

	var val float64
	switch v := x.(type) {
	case starlark.Int:
		i, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("value is too large")
		}
		val = float64(i)
	case starlark.Float:
		val = float64(v)
	default:
		return nil, fmt.Errorf("sin: got %s, want number", x.Type())
	}

	return starlark.Float(math.Sin(val)), nil
}

func cos(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &x); err != nil {
		return nil, err
	}

	var val float64
	switch v := x.(type) {
	case starlark.Int:
		i, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("value is too large")
		}
		val = float64(i)
	case starlark.Float:
		val = float64(v)
	default:
		return nil, fmt.Errorf("cos: got %s, want number", x.Type())
	}

	return starlark.Float(math.Cos(val)), nil
}

func tan(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &x); err != nil {
		return nil, err
	}

	var val float64
	switch v := x.(type) {
	case starlark.Int:
		i, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("value is too large")
		}
		val = float64(i)
	case starlark.Float:
		val = float64(v)
	default:
		return nil, fmt.Errorf("tan: got %s, want number", x.Type())
	}

	return starlark.Float(math.Tan(val)), nil
}

func asin(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &x); err != nil {
		return nil, err
	}

	var val float64
	switch v := x.(type) {
	case starlark.Int:
		i, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("value is too large")
		}
		val = float64(i)
	case starlark.Float:
		val = float64(v)
	default:
		return nil, fmt.Errorf("asin: got %s, want number", x.Type())
	}

	if val < -1 || val > 1 {
		return nil, fmt.Errorf("asin: argument out of domain")
	}

	return starlark.Float(math.Asin(val)), nil
}

func acos(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &x); err != nil {
		return nil, err
	}

	var val float64
	switch v := x.(type) {
	case starlark.Int:
		i, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("value is too large")
		}
		val = float64(i)
	case starlark.Float:
		val = float64(v)
	default:
		return nil, fmt.Errorf("acos: got %s, want number", x.Type())
	}

	if val < -1 || val > 1 {
		return nil, fmt.Errorf("acos: argument out of domain")
	}

	return starlark.Float(math.Acos(val)), nil
}

func atan(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &x); err != nil {
		return nil, err
	}

	var val float64
	switch v := x.(type) {
	case starlark.Int:
		i, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("value is too large")
		}
		val = float64(i)
	case starlark.Float:
		val = float64(v)
	default:
		return nil, fmt.Errorf("atan: got %s, want number", x.Type())
	}

	return starlark.Float(math.Atan(val)), nil
}

func atan2(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var y, x starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "y", &y, "x", &x); err != nil {
		return nil, err
	}

	var yVal, xVal float64

	switch v := y.(type) {
	case starlark.Int:
		i, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("y value is too large")
		}
		yVal = float64(i)
	case starlark.Float:
		yVal = float64(v)
	default:
		return nil, fmt.Errorf("atan2: got %s for y, want number", y.Type())
	}

	switch v := x.(type) {
	case starlark.Int:
		i, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("x value is too large")
		}
		xVal = float64(i)
	case starlark.Float:
		xVal = float64(v)
	default:
		return nil, fmt.Errorf("atan2: got %s for x, want number", x.Type())
	}

	return starlark.Float(math.Atan2(yVal, xVal)), nil
}

func abs(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &x); err != nil {
		return nil, err
	}

	switch v := x.(type) {
	case starlark.Int:
		i, ok := v.Int64()
		if ok && i < 0 {
			return starlark.MakeInt(-int(i)), nil
		}
		return x, nil
	case starlark.Float:
		if v < 0 {
			return starlark.Float(-v), nil
		}
		return x, nil
	default:
		return nil, fmt.Errorf("abs: got %s, want number", x.Type())
	}
}

func exp(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &x); err != nil {
		return nil, err
	}

	var val float64
	switch v := x.(type) {
	case starlark.Int:
		i, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("value is too large")
		}
		val = float64(i)
	case starlark.Float:
		val = float64(v)
	default:
		return nil, fmt.Errorf("exp: got %s, want number", x.Type())
	}

	return starlark.Float(math.Exp(val)), nil
}

func log(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x, base starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &x, "base?", &base); err != nil {
		return nil, err
	}

	var xVal float64
	switch v := x.(type) {
	case starlark.Int:
		i, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("value is too large")
		}
		xVal = float64(i)
	case starlark.Float:
		xVal = float64(v)
	default:
		return nil, fmt.Errorf("log: got %s, want number", x.Type())
	}

	if base == nil {
		return starlark.Float(math.Log(xVal)), nil
	}

	var baseVal float64
	switch v := base.(type) {
	case starlark.Int:
		i, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("base value is too large")
		}
		baseVal = float64(i)
	case starlark.Float:
		baseVal = float64(v)
	default:
		return nil, fmt.Errorf("log: got %s, want number", base.Type())
	}

	return starlark.Float(math.Log(xVal) / math.Log(baseVal)), nil
}

func log10(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &x); err != nil {
		return nil, err
	}

	var val float64
	switch v := x.(type) {
	case starlark.Int:
		i, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("value is too large")
		}
		val = float64(i)
	case starlark.Float:
		val = float64(v)
	default:
		return nil, fmt.Errorf("log10: got %s, want number", x.Type())
	}

	return starlark.Float(math.Log10(val)), nil
}
