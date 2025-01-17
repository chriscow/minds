package minds

import (
	"context"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestThreadContext(t *testing.T) {
	is := is.New(t)

	t.Run("NewThreadContext", func(t *testing.T) {
		is := is.New(t)
		ctx := context.Background()
		tc := NewThreadContext(ctx)

		is.True(tc != nil)
		is.Equal(ctx, tc.Context())
		is.True(tc.UUID() != "")
		is.Equal(len(tc.Messages()), 0)
		is.Equal(len(tc.Metadata()), 0)
	})

	t.Run("AppendMessage", func(t *testing.T) {
		is := is.New(t)
		tc := NewThreadContext(context.Background())
		msg := Message{Content: "test"}

		is.True(len(tc.Messages()) == 0)

		tc.AppendMessages(msg)
		msgs := tc.Messages()

		is.Equal(len(msgs), 1)
		is.Equal(msgs[0].Content, msg.Content)
	})

	t.Run("WithContext", func(t *testing.T) {
		is := is.New(t)
		oldCtx := context.Background()
		tc := NewThreadContext(oldCtx)

		newCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		newTc := tc.WithContext(newCtx)

		is.True(tc != newTc)
		is.Equal(newTc.Context(), newCtx)
		is.Equal(newTc.UUID(), tc.UUID())
		is.Equal(newTc.Messages(), tc.Messages())
		is.Equal(newTc.Metadata(), tc.Metadata())
	})

	t.Run("WithUUID", func(t *testing.T) {
		is := is.New(t)
		tc := NewThreadContext(context.Background())
		newUUID := "test-uuid"

		newTc := tc.WithUUID(newUUID)

		is.True(tc != newTc)
		is.Equal(newTc.UUID(), newUUID)
		is.Equal(newTc.Context(), tc.Context())
		is.Equal(newTc.Messages(), tc.Messages())
		is.Equal(newTc.Metadata(), tc.Metadata())
	})

	t.Run("WithMessages", func(t *testing.T) {
		is := is.New(t)
		tc := NewThreadContext(context.Background())
		msgs := Messages{{Content: "test1"}, {Content: "test2"}}

		newTc := tc.WithMessages(msgs...)

		is.True(tc != newTc)
		resultMsgs := newTc.Messages()
		is.Equal(len(resultMsgs), len(msgs))
		for i := range msgs {
			is.Equal(resultMsgs[i].Content, msgs[i].Content)
		}
		is.Equal(newTc.Context(), tc.Context())
		is.Equal(newTc.UUID(), tc.UUID())
		is.Equal(newTc.Metadata(), tc.Metadata())
	})

	t.Run("WithMetadata", func(t *testing.T) {
		is := is.New(t)
		tc := NewThreadContext(context.Background())
		metadata := Metadata{"key": "value"}

		newTc := tc.WithMetadata(metadata)

		is.True(tc != newTc)
		is.Equal(newTc.Metadata(), metadata)
		is.Equal(newTc.Context(), tc.Context())
		is.Equal(newTc.UUID(), tc.UUID())
		is.Equal(newTc.Messages(), tc.Messages())
	})

	t.Run("Concurrency", func(t *testing.T) {
		is := is.New(t)
		tc := NewThreadContext(context.Background())
		done := make(chan bool)

		go func() {
			for i := 0; i < 100; i++ {
				tc.AppendMessages(Message{Content: "test"})
			}
			done <- true
		}()

		go func() {
			var msgs Messages
			for i := 0; i < 100; i++ {
				msgs = tc.Messages()
			}
			is.True(len(msgs) >= 0) // Verify we can read messages
			done <- true
		}()

		<-done
		<-done

		finalMsgs := tc.Messages()
		is.True(len(finalMsgs) > 0) // Verify messages were actually appended
	})
}
