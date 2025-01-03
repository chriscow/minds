package calculator

import (
	"context"
	"testing"

	"github.com/matryer/is"
)

func TestCalculator(t *testing.T) {
	is := is.New(t)

	t.Run("basic python addition", func(t *testing.T) {
		calc, _ := NewCalculator(Starlark)
		result, err := calc.Call(context.Background(), []byte(`{"input":"1 + 2"}`))
		is.NoErr(err)
		is.Equal(string(result), "3")
	})

	t.Run("basic python subtraction", func(t *testing.T) {
		calc, _ := NewCalculator(Starlark)
		result, err := calc.Call(context.Background(), []byte(`{"input":"2 - 1"}`))
		is.NoErr(err)
		is.Equal(string(result), "1")
	})

	t.Run("basic python multiplication", func(t *testing.T) {
		calc, _ := NewCalculator(Starlark)
		result, err := calc.Call(context.Background(), []byte(`{"input":"2 * 3"}`))
		is.NoErr(err)
		is.Equal(string(result), "6")
	})

	t.Run("basic python division", func(t *testing.T) {
		calc, _ := NewCalculator(Starlark)
		result, err := calc.Call(context.Background(), []byte(`{"input":"6 / 2"}`))
		is.NoErr(err)
		is.Equal(string(result), "3.0")
	})

	t.Run("basic python modulo", func(t *testing.T) {
		calc, _ := NewCalculator(Starlark)
		result, err := calc.Call(context.Background(), []byte(`{"input":"7 % 3"}`))
		is.NoErr(err)
		is.Equal(string(result), "1")
	})

	t.Run("basic python exponentiation", func(t *testing.T) {
		calc, _ := NewCalculator(Starlark)
		script := `{"input":"math.sqrt(16) + math.pow(2, 3)"}`
		result, err := calc.Call(context.Background(), []byte(script))
		is.NoErr(err)
		is.Equal(string(result), "12.0")
	})
}
