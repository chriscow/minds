package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/chriscow/minds"
)

// StructuredExtractor is a handler that extracts structured data from conversation messages
// using an LLM and stores it in the ThreadContext metadata.
// It uses the provided prompt, ResponseSchema, and ContentGenerator to analyze the conversation.
type StructuredExtractor struct {
	name       string
	generator  minds.ContentGenerator
	prompt     string
	schema     minds.ResponseSchema
	middleware []minds.Middleware
}

// NewStructuredExtractor creates a new StructuredExtractor handler.
// The name parameter is used for debugging and logging.
// The generator is used to analyze messages with the given prompt and schema.
// The prompt should instruct the LLM to extract structured data from the conversation.
// The schema defines the structure of the data to extract.
func NewStructuredExtractor(name string, generator minds.ContentGenerator, prompt string, schema minds.ResponseSchema) *StructuredExtractor {
	return &StructuredExtractor{
		name:       name,
		generator:  generator,
		prompt:     prompt,
		schema:     schema,
		middleware: []minds.Middleware{},
	}
}

// Use applies middleware to the StructuredExtractor handler.
func (s *StructuredExtractor) Use(middleware ...minds.Middleware) {
	s.middleware = append(s.middleware, middleware...)
}

// With returns a new StructuredExtractor with additional middleware, preserving existing state.
func (s *StructuredExtractor) With(middleware ...minds.Middleware) *StructuredExtractor {
	newExtractor := &StructuredExtractor{
		name:       s.name,
		generator:  s.generator,
		prompt:     s.prompt,
		schema:     s.schema,
		middleware: append([]minds.Middleware{}, s.middleware...),
	}
	newExtractor.Use(middleware...)
	return newExtractor
}

// HandleThread processes the thread context by extracting structured data from messages.
func (s *StructuredExtractor) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	// If there are middleware registered, apply them first and let them handle the processing
	if len(s.middleware) > 0 {
		// Create a handler that executes the actual logic
		handler := minds.ThreadHandlerFunc(func(ctx minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			return s.extractData(ctx, next)
		})

		// Apply middleware in reverse order
		var wrappedHandler minds.ThreadHandler = handler
		for i := len(s.middleware) - 1; i >= 0; i-- {
			wrappedHandler = s.middleware[i].Wrap(wrappedHandler)
		}

		// Execute the wrapped handler
		return wrappedHandler.HandleThread(tc, next)
	}

	// Otherwise, directly process the extraction
	return s.extractData(tc, next)
}

// extractData performs the actual data extraction logic
func (s *StructuredExtractor) extractData(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	// Create a message that combines the instruction prompt with the conversation context
	messages := minds.Messages{
		{Role: minds.RoleSystem, Content: s.prompt},
	}

	// Add messages from the conversation as context
	for _, msg := range tc.Messages() {
		messages = append(messages, minds.Message{
			Role:    minds.RoleUser,
			Content: fmt.Sprintf("%s: %s", msg.Role, msg.Content),
		})
	}

	// Create and send the request to the LLM
	req := minds.NewRequest(messages, minds.WithResponseSchema(s.schema))
	resp, err := s.generator.GenerateContent(tc.Context(), req)
	if err != nil {
		return tc, fmt.Errorf("%s: error generating content: %w", s.name, err)
	}

	// Use the schema name as the key in metadata
	newTc := tc.Clone()

	// Parse the response as a generic JSON structure
	var data any
	if err := json.Unmarshal([]byte(resp.String()), &data); err != nil {
		return tc, fmt.Errorf("%s: error parsing structured data: %w", s.name, err)
	}

	// Store the structured data in metadata using the schema name as the key
	newTc.SetKeyValue(s.schema.Name, data)

	// Process next handler if provided
	if next != nil {
		return next.HandleThread(newTc, nil)
	}

	return newTc, nil
}

// String returns a string representation of the StructuredExtractor handler.
func (s *StructuredExtractor) String() string {
	return fmt.Sprintf("StructuredExtractor(%s)", s.name)
}
