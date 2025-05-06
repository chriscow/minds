package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/chriscow/minds"
)

// FreeformExtractor is a handler that extracts name-value pairs from conversation messages
// using an LLM and stores them in the ThreadContext metadata.
// It uses the provided prompt and ContentGenerator to analyze the conversation.
type FreeformExtractor struct {
	name       string
	generator  minds.ContentGenerator
	prompt     string
	middleware []minds.Middleware
}

// NewFreeformExtractor creates a new FreeformExtractor handler.
// The name parameter is used for debugging and logging.
// The generator is used to analyze messages with the given prompt.
// The prompt should instruct the LLM to extract name-value pairs from the conversation.
func NewFreeformExtractor(name string, generator minds.ContentGenerator, prompt string) *FreeformExtractor {
	return &FreeformExtractor{
		name:       name,
		generator:  generator,
		prompt:     prompt,
		middleware: []minds.Middleware{},
	}
}

// Use applies middleware to the FreeformExtractor handler.
func (f *FreeformExtractor) Use(middleware ...minds.Middleware) {
	f.middleware = append(f.middleware, middleware...)
}

// With returns a new FreeformExtractor with additional middleware, preserving existing state.
func (f *FreeformExtractor) With(middleware ...minds.Middleware) *FreeformExtractor {
	newExtractor := &FreeformExtractor{
		name:       f.name,
		generator:  f.generator,
		prompt:     f.prompt,
		middleware: append([]minds.Middleware{}, f.middleware...),
	}
	newExtractor.Use(middleware...)
	return newExtractor
}

// HandleThread processes the thread context by extracting name-value pairs from messages.
func (f *FreeformExtractor) HandleThread(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	// If there are middleware registered, apply them first and let them handle the processing
	if len(f.middleware) > 0 {
		// Create a handler that executes the actual logic
		handler := minds.ThreadHandlerFunc(func(ctx minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			return f.extractData(ctx, next)
		})

		// Apply middleware in reverse order
		var wrappedHandler minds.ThreadHandler = handler
		for i := len(f.middleware) - 1; i >= 0; i-- {
			wrappedHandler = f.middleware[i].Wrap(wrappedHandler)
		}

		// Execute the wrapped handler
		return wrappedHandler.HandleThread(tc, next)
	}

	// Otherwise, directly process the extraction
	return f.extractData(tc, next)
}

// extractData performs the actual data extraction logic
func (f *FreeformExtractor) extractData(tc minds.ThreadContext, next minds.ThreadHandler) (minds.ThreadContext, error) {
	// Create a message that combines the instruction prompt with the conversation context
	messages := minds.Messages{
		{Role: minds.RoleSystem, Content: f.prompt},
	}

	// Add messages from the conversation as context
	for _, msg := range tc.Messages() {
		messages = append(messages, minds.Message{
			Role:    minds.RoleUser,
			Content: fmt.Sprintf("%s: %s", msg.Role, msg.Content),
		})
	}

	// Define a response schema that will contain an array of key-value pairs
	// Since we can't use interface{} in the schema, we'll store all values as strings
	// and parse them as needed after extraction
	type KeyValuePair struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	type ExtractionResult struct {
		Pairs []KeyValuePair `json:"pairs"`
	}

	schema, err := minds.NewResponseSchema("extraction_result", "Key-value pairs extracted from the conversation", ExtractionResult{})
	if err != nil {
		return tc, fmt.Errorf("%s: error creating schema: %w", f.name, err)
	}

	// Create and send the request to the LLM
	req := minds.NewRequest(messages, minds.WithResponseSchema(*schema))
	resp, err := f.generator.GenerateContent(tc.Context(), req)
	if err != nil {
		return tc, fmt.Errorf("%s: error generating content: %w", f.name, err)
	}

	// Parse the response
	var result ExtractionResult
	if err := json.Unmarshal([]byte(resp.String()), &result); err != nil {
		return tc, fmt.Errorf("%s: error parsing extraction result: %w", f.name, err)
	}

	// Add all extracted values to the thread context metadata
	// Try to parse numeric and boolean values from string
	newTc := tc.Clone()
	for _, pair := range result.Pairs {
		// Try to convert the string value to appropriate type
		value := parseValue(pair.Value)
		newTc.SetKeyValue(pair.Key, value)
	}

	// Process next handler if provided
	if next != nil {
		return next.HandleThread(newTc, nil)
	}

	return newTc, nil
}

// parseValue tries to convert a string to a more appropriate type (number, boolean, etc.)
func parseValue(s string) interface{} {
	// Try to parse as integer
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}

	// Try to parse as float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	// Try to parse as boolean
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}

	// Default to string
	return s
}

// String returns a string representation of the FreeformExtractor handler.
func (f *FreeformExtractor) String() string {
	return fmt.Sprintf("FreeformExtractor(%s)", f.name)
}
