package extensions

import (
	"context"
	"testing"

	"github.com/chriscow/minds"
	"github.com/matryer/is"
)

func TestLuaCondition(t *testing.T) {
	is := is.New(t)

	t.Run("simple boolean return", func(t *testing.T) {
		is := is.New(t)

		cond := LuaCondition{
			Script: "return true",
		}

		tc := minds.NewThreadContext(context.Background())

		result, err := cond.Evaluate(tc)
		is.NoErr(err)
		is.True(result)
	})

	t.Run("complex logic with message content", func(t *testing.T) {
		is := is.New(t)

		cond := LuaCondition{
			Script: `
				local msg = last_message
				return string.match(msg, "important") ~= nil
			`,
		}
		tc := minds.NewThreadContext(context.Background()).
			WithMessages(minds.Message{
				Role: minds.RoleUser, Content: "This is not important",
			})

		result, err := cond.Evaluate(tc)
		is.NoErr(err)
		is.True(result)
	})

	t.Run("invalid script", func(t *testing.T) {
		is := is.New(t)

		cond := LuaCondition{
			Script: "invalid lua code",
		}

		tc := minds.NewThreadContext(context.Background())

		_, err := cond.Evaluate(tc)
		is.True(err != nil) // Should return an error
	})

	t.Run("non-boolean return", func(t *testing.T) {
		is := is.New(t)

		cond := LuaCondition{
			Script: "return 'string'",
		}

		tc := minds.NewThreadContext(context.Background())
		_, err := cond.Evaluate(tc)
		is.True(err != nil) // Should return an error
	})
}
