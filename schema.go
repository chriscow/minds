package minds

// Provides very simple functionality for representing a JSON schema as a
// (nested) struct. This struct can be used with the chat completion "function call" feature.
// For more complicated schemas, it is recommended to use a dedicated JSON schema library
// and/or pass in the schema in []byte format.

// This code was substantially borrowed with appreciationfrom
// github.com/sashabaranov/go-openai/jsonschema
// and modified to fit the needs of the project.

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type DataType string

const (
	Object  DataType = "object"
	Number  DataType = "number"
	Integer DataType = "integer"
	String  DataType = "string"
	Array   DataType = "array"
	Null    DataType = "null"
	Boolean DataType = "boolean"
)

// Definition is a struct for describing a JSON Schema.
// It is fairly limited, and you may have better luck using a third-party library.
type Definition struct {
	// Type specifies the data type of the schema.
	Type DataType `json:"type,omitempty"`
	// Description is the description of the schema.
	Description string `json:"description,omitempty"`
	// Enum is used to restrict a value to a fixed set of values. It must be an array with at least
	// one element, where each element is unique. You will probably only use this with strings.
	Enum []string `json:"enum,omitempty"`
	// Properties describes the properties of an object, if the schema type is Object.
	Properties map[string]Definition `json:"properties,omitempty"`
	// Required specifies which properties are required, if the schema type is Object.
	Required []string `json:"required,omitempty"`
	// Items specifies which data type an array contains, if the schema type is Array.
	Items *Definition `json:"items,omitempty"`
	// AdditionalProperties is used to control the handling of properties in an object
	// that are not explicitly defined in the properties section of the schema. example:
	// additionalProperties: true
	// additionalProperties: false
	// additionalProperties: Definition{Type: String}
	AdditionalProperties any `json:"additionalProperties,omitempty"`
	// Whether the schema is nullable or not.
	Nullable bool `json:"nullable,omitempty"`
}

func (d *Definition) MarshalJSON() ([]byte, error) {
	if d.Properties == nil {
		d.Properties = make(map[string]Definition)
	}
	type Alias Definition
	return json.Marshal(struct {
		Alias
	}{
		Alias: (Alias)(*d),
	})
}

func (d *Definition) Unmarshal(content string, v any) error {
	return VerifySchemaAndUnmarshal(*d, []byte(content), v)
}

func GenerateSchema(v any) (*Definition, error) {
	return reflectSchema(reflect.TypeOf(v))
}

func reflectSchema(t reflect.Type) (*Definition, error) {
	var d Definition
	switch t.Kind() {
	case reflect.String:
		d.Type = String
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		d.Type = Integer
	case reflect.Float32, reflect.Float64:
		d.Type = Number
	case reflect.Bool:
		d.Type = Boolean
	case reflect.Slice, reflect.Array:
		d.Type = Array
		items, err := reflectSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		d.Items = items
	case reflect.Struct:
		d.Type = Object
		d.AdditionalProperties = false
		object, err := reflectSchemaObject(t)
		if err != nil {
			return nil, err
		}
		d = *object
	case reflect.Ptr:
		definition, err := reflectSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		d = *definition
	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128,
		reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.UnsafePointer:
		return nil, fmt.Errorf("unsupported type: %s", t.Kind().String())
	default:
	}
	return &d, nil
}

func reflectSchemaObject(t reflect.Type) (*Definition, error) {
	var d = Definition{
		Type:                 Object,
		Description:          t.Name(),
		AdditionalProperties: false,
	}
	properties := make(map[string]Definition)
	var requiredFields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		ignoreTag := field.Tag.Get("jsonschema")
		if ignoreTag == "-" {
			continue
		}

		jsonTag := field.Tag.Get("json")
		var required = true
		if jsonTag == "" {
			jsonTag = field.Name
		} else if strings.HasSuffix(jsonTag, ",omitempty") {
			jsonTag = strings.TrimSuffix(jsonTag, ",omitempty")
			required = false
		}

		item, err := reflectSchema(field.Type)
		if err != nil {
			return nil, err
		}
		description := field.Tag.Get("description")
		if description != "" {
			item.Description = description
		}
		properties[jsonTag] = *item

		enumTag := field.Tag.Get("enum")
		if enumTag != "" {
			item.Enum = strings.Split(enumTag, ",")
		}

		if n := field.Tag.Get("nullable"); n != "" {
			nullable, _ := strconv.ParseBool(n)
			item.Nullable = nullable
		}

		if s := field.Tag.Get("required"); s != "" {
			required, _ = strconv.ParseBool(s)
		}
		if required {
			requiredFields = append(requiredFields, jsonTag)
		}
	}
	d.Required = requiredFields
	d.Properties = properties
	return &d, nil
}

func VerifySchemaAndUnmarshal(schema Definition, content []byte, v any) error {
	var data any
	err := json.Unmarshal(content, &data)
	if err != nil {
		return err
	}
	if !Validate(schema, data) {
		return errors.New("data validation failed against the provided schema")
	}
	return json.Unmarshal(content, &v)
}

func Validate(schema Definition, data any) bool {
	switch schema.Type {
	case Object:
		return validateObject(schema, data)
	case Array:
		return validateArray(schema, data)
	case String:
		_, ok := data.(string)
		return ok
	case Number: // float64 and int
		_, ok := data.(float64)
		if !ok {
			_, ok = data.(int)
		}
		return ok
	case Boolean:
		_, ok := data.(bool)
		return ok
	case Integer:
		// Golang unmarshals all numbers as float64, so we need to check if the float64 is an integer
		if num, ok := data.(float64); ok {
			return num == float64(int64(num))
		}
		_, ok := data.(int)
		return ok
	case Null:
		return data == nil
	default:
		return false
	}
}

func validateObject(schema Definition, data any) bool {
	dataMap, ok := data.(map[string]any)
	if !ok {
		return false
	}
	for _, field := range schema.Required {
		if _, exists := dataMap[field]; !exists {
			return false
		}
	}
	for key, valueSchema := range schema.Properties {
		value, exists := dataMap[key]
		if exists && !Validate(valueSchema, value) {
			return false
		} else if !exists && contains(schema.Required, key) {
			return false
		}
	}
	return true
}

func validateArray(schema Definition, data any) bool {
	dataArray, ok := data.([]any)
	if !ok {
		return false
	}
	for _, item := range dataArray {
		if !Validate(*schema.Items, item) {
			return false
		}
	}
	return true
}

func contains[S ~[]E, E comparable](s S, v E) bool {
	for i := range s {
		if v == s[i] {
			return true
		}
	}
	return false
}
