package calculator

import (
	"context"
	"testing"

	"github.com/matryer/is"
)

func TestCalculatorLua(t *testing.T) {
	is := is.New(t)

	t.Run("basic lua addition", func(t *testing.T) {
		calc, err := NewCalculator(Lua)
		is.NoErr(err)
		result, err := calc.Call(context.Background(), []byte("1 + 2"))
		is.NoErr(err)
		is.Equal(result, "3")
	})

	t.Run("basic lua subtraction", func(t *testing.T) {
		calc, err := NewCalculator(Lua)
		is.NoErr(err)
		result, err := calc.Call(context.Background(), []byte("2 - 1"))
		is.NoErr(err)
		is.Equal(result, "1")
	})

	t.Run("basic lua multiplication", func(t *testing.T) {
		calc, err := NewCalculator(Lua)
		is.NoErr(err)
		result, err := calc.Call(context.Background(), []byte("2 * 3"))
		is.NoErr(err)
		is.Equal(result, "6")
	})

	t.Run("basic lua division", func(t *testing.T) {
		calc, err := NewCalculator(Lua)
		is.NoErr(err)
		result, err := calc.Call(context.Background(), []byte("6 / 2"))
		is.NoErr(err)
		is.Equal(result, "3") // Lua numbers don't show decimal for whole numbers
	})

	t.Run("basic lua modulo", func(t *testing.T) {
		calc, err := NewCalculator(Lua)
		is.NoErr(err)
		result, err := calc.Call(context.Background(), []byte("7 % 3")) // In Lua, modulo is also %
		is.NoErr(err)
		is.Equal(result, "1")
	})

	t.Run("basic lua exponentiation", func(t *testing.T) {
		calc, err := NewCalculator(Lua)
		is.NoErr(err)
		// In Lua, we can use either math.pow() or the ^ operator
		script := "math.sqrt(16) + math.pow(2, 3)"
		result, err := calc.Call(context.Background(), []byte(script))
		is.NoErr(err)
		is.Equal(result, "12")
	})

	// Additional Lua-specific tests
	t.Run("lua power operator", func(t *testing.T) {
		calc, err := NewCalculator(Lua)
		is.NoErr(err)
		result, err := calc.Call(context.Background(), []byte("2^3")) // Lua's built-in power operator
		is.NoErr(err)
		is.Equal(result, "8")
	})

	t.Run("lua number formatting", func(t *testing.T) {
		calc, err := NewCalculator(Lua)
		is.NoErr(err)
		result, err := calc.Call(context.Background(), []byte("22/7")) // Will show decimal places
		is.NoErr(err)
		is.Equal(result[:4], "3.14") // Check first 4 characters for approximate pi
	})

	t.Run("lua math constants", func(t *testing.T) {
		calc, err := NewCalculator(Lua)
		is.NoErr(err)
		result, err := calc.Call(context.Background(), []byte("math.pi"))
		is.NoErr(err)
		is.Equal(result[:4], "3.14") // Check first 4 characters for pi
	})
}
