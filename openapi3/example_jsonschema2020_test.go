package openapi3_test

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// Example demonstrates how to enable and use the JSON Schema 2020-12 validator
// with OpenAPI 3.1 features.
func Example_jsonSchema2020Validator() {
	// Enable JSON Schema 2020-12 validator

	// Create a schema using OpenAPI 3.1 features
	schema := &openapi3.Schema{
		Type: &openapi3.Types{"object"},
		Properties: openapi3.Schemas{
			"name": &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},
					Examples: []any{
						"John Doe",
						"Jane Smith",
					},
				},
			},
			"age": &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					// Type array with null - OpenAPI 3.1 feature
					Type: &openapi3.Types{"integer", "null"},
				},
			},
			"status": &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					// Const keyword - OpenAPI 3.1 feature
					Const: "active",
				},
			},
		},
		Required: []string{"name", "status"},
	}

	// Valid data
	validData := map[string]any{
		"name":   "John Doe",
		"age":    30,
		"status": "active",
	}

	if err := schema.VisitJSON(validData, openapi3.EnableJSONSchema2020()); err != nil {
		fmt.Println("validation failed:", err)
	} else {
		fmt.Println("valid data passed")
	}

	// Valid with null age
	validWithNull := map[string]any{
		"name":   "Jane Smith",
		"age":    nil, // null is allowed in type array
		"status": "active",
	}

	if err := schema.VisitJSON(validWithNull, openapi3.EnableJSONSchema2020()); err != nil {
		fmt.Println("validation failed:", err)
	} else {
		fmt.Println("valid data with null passed")
	}

	// Invalid: wrong const value
	invalidData := map[string]any{
		"name":   "Bob Wilson",
		"age":    25,
		"status": "inactive", // should be "active"
	}

	if err := schema.VisitJSON(invalidData, openapi3.EnableJSONSchema2020()); err != nil {
		fmt.Println("invalid data rejected")
	}

	// Output:
	// valid data passed
	// valid data with null passed
	// invalid data rejected
}

// Example demonstrates type arrays with null support
func Example_typeArrayWithNull() {

	schema := &openapi3.Schema{
		Type: &openapi3.Types{"string", "null"},
	}

	// Both string and null are valid
	if err := schema.VisitJSON("hello", openapi3.EnableJSONSchema2020()); err == nil {
		fmt.Println("string accepted")
	}

	if err := schema.VisitJSON(nil, openapi3.EnableJSONSchema2020()); err == nil {
		fmt.Println("null accepted")
	}

	if err := schema.VisitJSON(123, openapi3.EnableJSONSchema2020()); err != nil {
		fmt.Println("number rejected")
	}

	// Output:
	// string accepted
	// null accepted
	// number rejected
}

// Example demonstrates the const keyword
func Example_constKeyword() {

	schema := &openapi3.Schema{
		Const: "production",
	}

	if err := schema.VisitJSON("production", openapi3.EnableJSONSchema2020()); err == nil {
		fmt.Println("const value accepted")
	}

	if err := schema.VisitJSON("development", openapi3.EnableJSONSchema2020()); err != nil {
		fmt.Println("other value rejected")
	}

	// Output:
	// const value accepted
	// other value rejected
}

// Example demonstrates the examples field
func Example_examplesField() {

	schema := &openapi3.Schema{
		Type: &openapi3.Types{"string"},
		// Examples array - OpenAPI 3.1 feature
		Examples: []any{
			"red",
			"green",
			"blue",
		},
	}

	// Examples don't affect validation, any string is valid
	if err := schema.VisitJSON("yellow", openapi3.EnableJSONSchema2020()); err == nil {
		fmt.Println("any string accepted")
	}

	// Output:
	// any string accepted
}

// Example demonstrates backward compatibility with nullable
func Example_nullableBackwardCompatibility() {

	// OpenAPI 3.0 style nullable
	schema := &openapi3.Schema{
		Type:     &openapi3.Types{"string"},
		Nullable: true,
	}

	// Automatically converted to type array ["string", "null"]
	if err := schema.VisitJSON("hello", openapi3.EnableJSONSchema2020()); err == nil {
		fmt.Println("string accepted")
	}

	if err := schema.VisitJSON(nil, openapi3.EnableJSONSchema2020()); err == nil {
		fmt.Println("null accepted")
	}

	// Output:
	// string accepted
	// null accepted
}

// Example demonstrates complex nested schemas
func Example_complexNestedSchema() {

	min := 0.0
	max := 100.0

	schema := &openapi3.Schema{
		Type: &openapi3.Types{"object"},
		Properties: openapi3.Schemas{
			"user": &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"object"},
					Properties: openapi3.Schemas{
						"name": &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
						"email": &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type:   &openapi3.Types{"string"},
								Format: "email",
							},
						},
					},
					Required: []string{"name", "email"},
				},
			},
			"score": &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"number"},
					Min:  &min,
					Max:  &max,
				},
			},
		},
		Required: []string{"user", "score"},
	}

	validData := map[string]any{
		"user": map[string]any{
			"name":  "John Doe",
			"email": "john@example.com",
		},
		"score": 85.5,
	}

	if err := schema.VisitJSON(validData, openapi3.EnableJSONSchema2020()); err == nil {
		fmt.Println("complex nested object validated")
	}

	// Output:
	// complex nested object validated
}

// Example demonstrates using both validators for comparison
func Example_comparingValidators() {
	schema := &openapi3.Schema{
		Type:      &openapi3.Types{"string"},
		MinLength: 5,
	}

	testValue := "test"

	// Test with built-in validator (no option)
	err1 := schema.VisitJSON(testValue)
	if err1 != nil {
		fmt.Println("built-in validator: rejected")
	}

	// Test with JSON Schema 2020-12 validator
	err2 := schema.VisitJSON(testValue, openapi3.EnableJSONSchema2020())
	if err2 != nil {
		fmt.Println("visit JSON Schema 2020-12 validator: rejected")
	}

	// Output:
	// built-in validator: rejected
	// visit JSON Schema 2020-12 validator: rejected
}
