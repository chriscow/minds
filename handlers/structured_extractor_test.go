package handlers

import (
	"context"
	"testing"

	"github.com/chriscow/minds"
	"github.com/matryer/is"
)

func TestStructuredExtractor(t *testing.T) {
	is := is.New(t)

	// Define a schema for structured data extraction
	type PersonInfo struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	schema, err := minds.NewResponseSchema("person_info", "Information about a person", PersonInfo{})
	is.NoErr(err)

	// Mock response with structured data matching the schema
	mockResponse := `{"name": "Jane Smith", "age": 25, "email": "jane@example.com"}`

	// Create a mock content generator
	generator := &MockContentGenerator{response: mockResponse}

	// Create a thread context with messages
	tc := minds.NewThreadContext(context.Background())
	tc.AppendMessages(
		minds.Message{Role: minds.RoleUser, Content: "Hello, my name is Jane Smith"},
		minds.Message{Role: minds.RoleAssistant, Content: "Hi Jane, how can I help?"},
		minds.Message{Role: minds.RoleUser, Content: "I'm 25 years old and my email is jane@example.com"},
	)

	// Create a StructuredExtractor handler
	extractor := NewStructuredExtractor(
		"person-extractor",
		generator,
		"Extract name, age, and email information from the conversation.",
		*schema,
	)

	// Process the thread
	result, err := extractor.HandleThread(tc, nil)

	// Verify no error occurred
	is.NoErr(err)

	// Verify the metadata contains the extracted structured data under the schema name
	metadata := result.Metadata()
	personInfo, exists := metadata["person_info"]
	is.True(exists)

	// Check the structure of the extracted data
	info, ok := personInfo.(map[string]interface{})
	is.True(ok)
	is.Equal(info["name"], "Jane Smith")
	is.Equal(info["age"], float64(25)) // JSON numbers are parsed as float64
	is.Equal(info["email"], "jane@example.com")
}

func TestStructuredExtractor_WithNext(t *testing.T) {
	is := is.New(t)

	// Define a schema for structured data extraction
	type OrderInfo struct {
		ProductName string  `json:"product_name"`
		Quantity    int     `json:"quantity"`
		Price       float64 `json:"price"`
	}

	schema, err := minds.NewResponseSchema("order_info", "Information about an order", OrderInfo{})
	is.NoErr(err)

	// Mock response with structured data matching the schema
	mockResponse := `{"product_name": "Laptop", "quantity": 1, "price": 999.99}`

	// Create a mock content generator
	generator := &MockContentGenerator{response: mockResponse}

	// Create a thread context with messages
	tc := minds.NewThreadContext(context.Background())
	tc.AppendMessages(
		minds.Message{Role: minds.RoleUser, Content: "I'd like to order a laptop"},
		minds.Message{Role: minds.RoleAssistant, Content: "How many would you like?"},
		minds.Message{Role: minds.RoleUser, Content: "Just one, and my budget is $1000"},
	)

	// Create a next handler that adds more metadata
	nextHandler := minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
		tc.SetKeyValue("order_processed", true)
		return tc, nil
	})

	// Create a StructuredExtractor handler
	extractor := NewStructuredExtractor(
		"order-extractor",
		generator,
		"Extract product name, quantity, and price information from the conversation.",
		*schema,
	)

	// Process the thread
	result, err := extractor.HandleThread(tc, nextHandler)

	// Verify no error occurred
	is.NoErr(err)

	// Verify the metadata contains both the extracted structured data and the next handler's addition
	metadata := result.Metadata()
	orderInfo, exists := metadata["order_info"]
	is.True(exists)

	// Check the structure of the extracted data
	info, ok := orderInfo.(map[string]interface{})
	is.True(ok)
	is.Equal(info["product_name"], "Laptop")
	is.Equal(info["quantity"], float64(1))
	is.Equal(info["price"], 999.99)

	// Check the next handler was executed
	is.Equal(metadata["order_processed"], true)
}

func TestStructuredExtractor_WithMiddleware(t *testing.T) {
	is := is.New(t)

	// Define a schema for structured data extraction
	type Address struct {
		Street  string `json:"street"`
		City    string `json:"city"`
		ZipCode string `json:"zip_code"`
	}

	schema, err := minds.NewResponseSchema("address_info", "Address information", Address{})
	is.NoErr(err)

	// Mock response with structured data matching the schema
	mockResponse := `{"street": "123 Main St", "city": "Springfield", "zip_code": "12345"}`

	// Create a mock content generator
	generator := &MockContentGenerator{response: mockResponse}

	// Create a thread context with messages
	tc := minds.NewThreadContext(context.Background())
	tc.AppendMessages(
		minds.Message{Role: minds.RoleUser, Content: "My address is 123 Main St, Springfield, 12345"},
	)

	// Create middleware that adds metadata
	middleware := minds.MiddlewareFunc(func(next minds.ThreadHandler) minds.ThreadHandler {
		return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			tc.SetKeyValue("address_validated", true)
			return next.HandleThread(tc, nil)
		})
	})

	// Create a StructuredExtractor handler with middleware
	extractor := NewStructuredExtractor(
		"address-extractor",
		generator,
		"Extract street, city, and zip code information from the conversation.",
		*schema,
	)
	extractor.Use(middleware)

	// Process the thread
	result, err := extractor.HandleThread(tc, nil)

	// Verify no error occurred
	is.NoErr(err)

	// Verify the metadata contains both the extracted structured data and the middleware's addition
	metadata := result.Metadata()
	addressInfo, exists := metadata["address_info"]
	is.True(exists)

	// Check the structure of the extracted data
	info, ok := addressInfo.(map[string]interface{})
	is.True(ok)
	is.Equal(info["street"], "123 Main St")
	is.Equal(info["city"], "Springfield")
	is.Equal(info["zip_code"], "12345")

	// Check the middleware was applied
	is.Equal(metadata["address_validated"], true)
}
