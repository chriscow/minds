package handlers

import (
	"fmt"

	"github.com/chriscow/minds"
)

type sequential struct {
	name         string
	handlers     []minds.ThreadHandler
	finalHandler minds.ThreadHandler
	middlware    minds.ThreadHandler
}

// Sequential creates a ThreadHandler that executes a series of handlers in sequence.
// If any handler returns an error, the sequence stops, and the error is returned.
// You can add handlers to the sequence after construction using AddHandler or AddHandlers.
// Optionally, you can specify a handler as middleware with the `Use` method.
//
// Example:
//
//	this := DoThisHandler()
//	that := DoThatHandler()
//	llm := LLMHandler()
//
//	// Sequentially run `this` -> `llm` -> `that` handlers
//	seq := Sequential("example", this, llm, that)
//	finalThread, err := seq.HandleThread(initialThread, nil)
//	if err != nil {
//	    log.Fatalf("Error handling thread: %v", err)
//	}
//	fmt.Println("Final thread:", finalThread.Messaages.Last().Content)
func Sequential(name string, handlers ...minds.ThreadHandler) *sequential {
	return &sequential{handlers: handlers, name: name}
}

func (s *sequential) AddHandler(handler minds.ThreadHandler) {
	s.handlers = append(s.handlers, handler)
}

func (s *sequential) AddHandlers(handlers ...minds.ThreadHandler) {
	s.handlers = append(s.handlers, handlers...)
}

func (s *sequential) Use(handler minds.ThreadHandler) {
	s.middlware = handler
}

func (s *sequential) String() string {
	return fmt.Sprintf("Sequential: %s", s.name)
}

// HandleThread processes the thread context by executing the sequence of handlers.
// If any handler returns an error, the sequence stops, and the error is returned.
// If a final handler is provided, it is executed after the sequence completes.
func (s *sequential) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	final := s.finalHandler
	if final == nil {
		final = next
	}

	for _, handler := range s.handlers {
		// Wrap the handler execution with middleware, if present
		if s.middlware != nil {
			var err error
			tc, err = s.middlware.HandleThread(tc, handler)
			if err != nil {
				return tc, err
			}
		} else {
			var err error
			tc, err = handler.HandleThread(tc, nil)
			if err != nil {
				return tc, err
			}
		}
	}

	if final != nil {
		return final.HandleThread(tc, nil)
	}

	return tc, nil
}
