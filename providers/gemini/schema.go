package gemini

import (
	"fmt"

	"github.com/chriscow/minds"

	"github.com/google/generative-ai-go/genai"
)

func GenerateSchema(v any) (*genai.Schema, error) {
	d, err := minds.GenerateSchema(v)
	if err != nil {
		return nil, fmt.Errorf("error generating schema: %w", err)
	}

	if d == nil {
		return nil, fmt.Errorf("definition cannot be nil")
	}

	return convertSchema(*d)
}

func convertSchema(d minds.Definition) (*genai.Schema, error) {
	schema := &genai.Schema{}

	// Map the type
	switch d.Type {
	case minds.String:
		schema.Type = genai.TypeString
	case minds.Number:
		schema.Type = genai.TypeNumber
	case minds.Integer:
		schema.Type = genai.TypeInteger
	case minds.Boolean:
		schema.Type = genai.TypeBoolean
	case minds.Array:
		schema.Type = genai.TypeArray
		if d.Items != nil {
			items, err := convertSchema(*d.Items)
			if err != nil {
				return nil, fmt.Errorf("error converting array items: %w", err)
			}
			schema.Items = items
		}
	case minds.Object:
		schema.Type = genai.TypeObject
		if len(d.Properties) > 0 {
			schema.Properties = make(map[string]*genai.Schema)
			for key, prop := range d.Properties {
				propSchema, err := convertSchema(prop)
				if err != nil {
					return nil, fmt.Errorf("error converting property %s: %w", key, err)
				}
				schema.Properties[key] = propSchema
			}
		}
		if len(d.Required) > 0 {
			schema.Required = d.Required
		}
	case minds.Null:
		schema.Type = genai.TypeUnspecified
		schema.Nullable = true
	default:
		return nil, fmt.Errorf("unsupported type: %s", d.Type)
	}

	// Copy description if present
	if d.Description != "" {
		schema.Description = d.Description
	}

	// Copy enum values if present
	if len(d.Enum) > 0 {
		schema.Enum = d.Enum
	}

	return schema, nil
}

func reflectMindsDefinition(schema *genai.Schema) (*minds.Definition, error) {
	definition := &minds.Definition{}

	// Map the type
	switch schema.Type {
	case genai.TypeString:
		definition.Type = minds.String
	case genai.TypeNumber:
		definition.Type = minds.Number
	case genai.TypeInteger:
		definition.Type = minds.Integer
	case genai.TypeBoolean:
		definition.Type = minds.Boolean
	case genai.TypeArray:
		definition.Type = minds.Array
		if schema.Items != nil {
			items, err := reflectMindsDefinition(schema.Items)
			if err != nil {
				return nil, fmt.Errorf("error converting array items: %w", err)
			}
			definition.Items = items
		}
	case genai.TypeObject:
		definition.Type = minds.Object
		if len(schema.Properties) > 0 {
			definition.Properties = make(map[string]minds.Definition)
			for key, propSchema := range schema.Properties {
				propDefinition, err := reflectMindsDefinition(propSchema)
				if err != nil {
					return nil, fmt.Errorf("error converting property %s: %w", key, err)
				}
				definition.Properties[key] = *propDefinition
			}
		}
	}

	return definition, nil
}
