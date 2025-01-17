package middleware

import (
	"github.com/chriscow/minds"
)

// Base middleware type that implements Middleware
type middleware struct {
	name string
	fn   func(minds.ThreadHandler) minds.ThreadHandler
}

func (m *middleware) Wrap(next minds.ThreadHandler) minds.ThreadHandler {
	return m.fn(next)
}
