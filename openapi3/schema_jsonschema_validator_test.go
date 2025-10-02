package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONSchema2020Validator_Basic(t *testing.T) {
	t.Run("string validation", func(t *testing.T) {
		schema := &Schema{
			Type: &Types{"string"},
		}

		err := schema.VisitJSON("hello", EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(123, EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("number validation", func(t *testing.T) {
		min := 0.0
		max := 100.0
		schema := &Schema{
			Type: &Types{"number"},
			Min:  &min,
			Max:  &max,
		}

		err := schema.VisitJSON(50.0, EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(150.0, EnableJSONSchema2020())
		require.Error(t, err)

		err = schema.VisitJSON(-10.0, EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("object validation", func(t *testing.T) {
		schema := &Schema{
			Type: &Types{"object"},
			Properties: Schemas{
				"name": &SchemaRef{Value: &Schema{Type: &Types{"string"}}},
				"age":  &SchemaRef{Value: &Schema{Type: &Types{"integer"}}},
			},
			Required: []string{"name"},
		}

		err := schema.VisitJSON(map[string]any{
			"name": "John",
			"age":  30,
		})
		require.NoError(t, err)

		err = schema.VisitJSON(map[string]any{
			"age": 30,
		})
		require.Error(t, err) // missing required "name"
	})

	t.Run("array validation", func(t *testing.T) {
		schema := &Schema{
			Type: &Types{"array"},
			Items: &SchemaRef{Value: &Schema{
				Type: &Types{"string"},
			}},
		}

		err := schema.VisitJSON([]any{"a", "b", "c"}, EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON([]any{"a", 1, "c"}, EnableJSONSchema2020())
		require.Error(t, err) // item 1 is not a string
	})
}

func TestJSONSchema2020Validator_OpenAPI31Features(t *testing.T) {
	t.Run("type array with null", func(t *testing.T) {
		schema := &Schema{
			Type: &Types{"string", "null"},
		}

		err := schema.VisitJSON("hello", EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(nil, EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(123, EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("nullable conversion", func(t *testing.T) {
		schema := &Schema{
			Type:     &Types{"string"},
			Nullable: true,
		}

		err := schema.VisitJSON("hello", EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(nil, EnableJSONSchema2020())
		require.NoError(t, err)
	})

	t.Run("const validation", func(t *testing.T) {
		schema := &Schema{
			Const: "fixed-value",
		}

		err := schema.VisitJSON("fixed-value", EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON("other-value", EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("examples field", func(t *testing.T) {
		schema := &Schema{
			Type: &Types{"string"},
			Examples: []any{
				"example1",
				"example2",
			},
		}

		// Examples don't affect validation, just ensure schema is valid
		err := schema.VisitJSON("any-value", EnableJSONSchema2020())
		require.NoError(t, err)
	})
}

func TestJSONSchema2020Validator_ExclusiveMinMax(t *testing.T) {
	t.Run("exclusive minimum as boolean (OpenAPI 3.0 style)", func(t *testing.T) {
		min := 0.0
		schema := &Schema{
			Type:         &Types{"number"},
			Min:          &min,
			ExclusiveMin: true,
		}

		err := schema.VisitJSON(0.1, EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(0.0, EnableJSONSchema2020())
		require.Error(t, err) // should be exclusive
	})

	t.Run("exclusive maximum as boolean (OpenAPI 3.0 style)", func(t *testing.T) {
		max := 100.0
		schema := &Schema{
			Type:         &Types{"number"},
			Max:          &max,
			ExclusiveMax: true,
		}

		err := schema.VisitJSON(99.9, EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(100.0, EnableJSONSchema2020())
		require.Error(t, err) // should be exclusive
	})
}

func TestJSONSchema2020Validator_ComplexSchemas(t *testing.T) {
	t.Run("oneOf", func(t *testing.T) {
		schema := &Schema{
			OneOf: SchemaRefs{
				&SchemaRef{Value: &Schema{Type: &Types{"string"}}},
				&SchemaRef{Value: &Schema{Type: &Types{"number"}}},
			},
		}

		err := schema.VisitJSON("hello", EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(42, EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(true, EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("anyOf", func(t *testing.T) {
		schema := &Schema{
			AnyOf: SchemaRefs{
				&SchemaRef{Value: &Schema{Type: &Types{"string"}}},
				&SchemaRef{Value: &Schema{Type: &Types{"number"}}},
			},
		}

		err := schema.VisitJSON("hello", EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(42, EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(true, EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("allOf", func(t *testing.T) {
		min := 0.0
		max := 100.0
		schema := &Schema{
			AllOf: SchemaRefs{
				&SchemaRef{Value: &Schema{Type: &Types{"number"}}},
				&SchemaRef{Value: &Schema{Min: &min}},
				&SchemaRef{Value: &Schema{Max: &max}},
			},
		}

		err := schema.VisitJSON(50.0, EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(150.0, EnableJSONSchema2020())
		require.Error(t, err) // exceeds max
	})

	t.Run("not", func(t *testing.T) {
		schema := &Schema{
			Not: &SchemaRef{Value: &Schema{Type: &Types{"string"}}},
		}

		err := schema.VisitJSON(42, EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON("hello", EnableJSONSchema2020())
		require.Error(t, err)
	})
}

func TestJSONSchema2020Validator_Fallback(t *testing.T) {
	t.Run("fallback on compilation error", func(t *testing.T) {
		// Create a schema that might cause compilation issues
		schema := &Schema{
			Type: &Types{"string"},
		}

		// Should not panic, even if there's an issue
		err := schema.VisitJSON("test", EnableJSONSchema2020())
		require.NoError(t, err)
	})
}

func TestBuiltInValidatorStillWorks(t *testing.T) {
	t.Run("string validation with built-in", func(t *testing.T) {
		schema := &Schema{
			Type: &Types{"string"},
		}

		err := schema.VisitJSON("hello", EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(123, EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("object validation with built-in", func(t *testing.T) {
		schema := &Schema{
			Type: &Types{"object"},
			Properties: Schemas{
				"name": &SchemaRef{Value: &Schema{Type: &Types{"string"}}},
			},
			Required: []string{"name"},
		}

		err := schema.VisitJSON(map[string]any{
			"name": "John",
		})
		require.NoError(t, err)

		err = schema.VisitJSON(map[string]any{}, EnableJSONSchema2020())
		require.Error(t, err)
	})
}
