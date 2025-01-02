package handlers

import (
	"fmt"

	"github.com/chriscow/minds"
)

type cycler struct {
	name       string
	handlers   []minds.ThreadHandler
	middleware minds.ThreadHandler
	iterations int // number of times to cycle, 0 means infinite
}

func Cycle(name string, iterations int, handlers ...minds.ThreadHandler) *cycler {
	return &cycler{
		name:       name,
		handlers:   handlers,
		iterations: iterations,
	}
}

func (c *cycler) Use(handler minds.ThreadHandler) {
	c.middleware = handler
}

func (c *cycler) String() string {
	return fmt.Sprintf("Cycle: %s", c.name)
}

func (c *cycler) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	count := 0
	for c.iterations == 0 || count < c.iterations {
		for _, handler := range c.handlers {
			if c.middleware != nil {
				var err error
				tc, err = c.middleware.HandleThread(tc, handler)
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
		count++
	}

	if next != nil {
		return next.HandleThread(tc, nil)
	}

	return tc, nil
}
