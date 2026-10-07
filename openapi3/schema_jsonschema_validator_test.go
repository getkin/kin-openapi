package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestJSONSchema2020Validator_Basic(t *testing.T) {
	t.Run("string validation", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"string"},
		}

		err := schema.VisitJSON("hello", openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(123, openapi3.EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("number validation", func(t *testing.T) {
		min := 0.0
		max := 100.0
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"number"},
			Min:  &min,
			Max:  &max,
		}

		err := schema.VisitJSON(50.0, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(150.0, openapi3.EnableJSONSchema2020())
		require.Error(t, err)

		err = schema.VisitJSON(-10.0, openapi3.EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("object validation", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"name": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
				"age":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}},
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
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"array"},
			Items: &openapi3.SchemaRef{Value: &openapi3.Schema{
				Type: &openapi3.Types{"string"},
			}},
		}

		err := schema.VisitJSON([]any{"a", "b", "c"}, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON([]any{"a", 1, "c"}, openapi3.EnableJSONSchema2020())
		require.Error(t, err) // item 1 is not a string
	})
}

func TestJSONSchema2020Validator_OpenAPI31Features(t *testing.T) {
	t.Run("type array with null", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"string", "null"},
		}

		err := schema.VisitJSON("hello", openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(nil, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(123, openapi3.EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("nullable is an unknown keyword", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:     &openapi3.Types{"string"},
			Nullable: true,
		}

		err := schema.VisitJSON("hello", openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		// Since OpenAPI 3.1, null is allowed by a type array including "null", not by nullable.
		err = schema.VisitJSON(nil, openapi3.EnableJSONSchema2020())
		require.Error(t, err)

		// The built-in (OpenAPI 3.0) validator still honours nullable.
		require.NoError(t, schema.VisitJSON(nil))
	})

	t.Run("const validation", func(t *testing.T) {
		schema := &openapi3.Schema{
			Const: "fixed-value",
		}

		err := schema.VisitJSON("fixed-value", openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON("other-value", openapi3.EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("examples field", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"string"},
			Examples: []any{
				"example1",
				"example2",
			},
		}

		// Examples don't affect validation, just ensure schema is valid
		err := schema.VisitJSON("any-value", openapi3.EnableJSONSchema2020())
		require.NoError(t, err)
	})
}

func TestJSONSchema2020Validator_ExclusiveMinMax(t *testing.T) {
	t.Run("exclusive minimum as boolean (OpenAPI 3.0 style)", func(t *testing.T) {
		min := 0.0
		boolTrue := true
		schema := &openapi3.Schema{
			Type:         &openapi3.Types{"number"},
			Min:          &min,
			ExclusiveMin: openapi3.ExclusiveBound{Bool: &boolTrue},
		}

		err := schema.VisitJSON(0.1, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(0.0, openapi3.EnableJSONSchema2020())
		require.Error(t, err) // should be exclusive
	})

	t.Run("exclusive maximum as boolean (OpenAPI 3.0 style)", func(t *testing.T) {
		max := 100.0
		boolTrue := true
		schema := &openapi3.Schema{
			Type:         &openapi3.Types{"number"},
			Max:          &max,
			ExclusiveMax: openapi3.ExclusiveBound{Bool: &boolTrue},
		}

		err := schema.VisitJSON(99.9, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(100.0, openapi3.EnableJSONSchema2020())
		require.Error(t, err) // should be exclusive
	})
}

func TestJSONSchema2020Validator_ComplexSchemas(t *testing.T) {
	t.Run("oneOf", func(t *testing.T) {
		schema := &openapi3.Schema{
			OneOf: openapi3.SchemaRefs{
				&openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
				&openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"number"}}},
			},
		}

		err := schema.VisitJSON("hello", openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(42, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(true, openapi3.EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("anyOf", func(t *testing.T) {
		schema := &openapi3.Schema{
			AnyOf: openapi3.SchemaRefs{
				&openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
				&openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"number"}}},
			},
		}

		err := schema.VisitJSON("hello", openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(42, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(true, openapi3.EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("allOf", func(t *testing.T) {
		min := 0.0
		max := 100.0
		schema := &openapi3.Schema{
			AllOf: openapi3.SchemaRefs{
				&openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"number"}}},
				&openapi3.SchemaRef{Value: &openapi3.Schema{Min: &min}},
				&openapi3.SchemaRef{Value: &openapi3.Schema{Max: &max}},
			},
		}

		err := schema.VisitJSON(50.0, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(150.0, openapi3.EnableJSONSchema2020())
		require.Error(t, err) // exceeds max
	})

	t.Run("not", func(t *testing.T) {
		schema := &openapi3.Schema{
			Not: &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
		}

		err := schema.VisitJSON(42, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON("hello", openapi3.EnableJSONSchema2020())
		require.Error(t, err)
	})
}

func TestJSONSchema2020Validator_CompilationError(t *testing.T) {
	schema := &openapi3.Schema{
		AllOf:                 openapi3.SchemaRefs{{Ref: "#/components/schemas/Base"}},
		UnevaluatedProperties: openapi3.BoolSchema{Has: new(bool)},
	}
	// The standalone JSON Schema compiler cannot resolve a document-relative
	// reference when its target is unavailable.
	err := schema.VisitJSON(map[string]any{}, openapi3.EnableJSONSchema2020())
	require.ErrorContains(t, err, "failed to compile schema")
	require.ErrorContains(t, err, "#/components/schemas/Base")
}

func TestJSONSchema2020Validator_PatternPropertiesWithComponentRef(t *testing.T) {
	const spec = `
openapi: 3.1.0
info: {title: t, version: "1"}
paths: {}
components:
  schemas:
    Node:
      type: object
      required: [name]
      properties:
        name: {type: string}
        tagged:
          type: object
          patternProperties:
            "^x": {$ref: "#/components/schemas/Node"}
`
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)
	node := doc.Components.Schemas["Node"].Value

	opt := openapi3.EnableJSONSchema2020()
	err = node.VisitJSON(map[string]any{"name": "a", "tagged": map[string]any{"x1": map[string]any{}}}, opt)
	require.Error(t, err)
	err = node.VisitJSON(map[string]any{"name": "a", "tagged": map[string]any{"x1": map[string]any{"name": "b"}, "y": map[string]any{}}}, opt)
	require.NoError(t, err)
}

func TestJSONSchema2020Validator_TransformRecursesInto31Fields(t *testing.T) {
	// These tests verify that transformOpenAPIToJSONSchema recurses into
	// OpenAPI 3.1 / JSON Schema 2020-12 fields. Each sub-test nests a schema
	// with a boolean exclusiveMinimum (an OpenAPI 3.0-ism) that must be
	// converted to a number for the JSON Schema 2020-12 validator to reject 0.
	positive := func() *openapi3.SchemaRef {
		return &openapi3.SchemaRef{Value: &openapi3.Schema{
			Type:         &openapi3.Types{"number"},
			Min:          openapi3.Float64Ptr(0),
			ExclusiveMin: openapi3.ExclusiveBound{Bool: openapi3.BoolPtr(true)},
		}}
	}
	opt := openapi3.EnableJSONSchema2020()

	t.Run("prefixItems", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:        &openapi3.Types{"array"},
			PrefixItems: openapi3.SchemaRefs{positive()},
		}
		require.NoError(t, schema.VisitJSON([]any{1}, opt))
		require.Error(t, schema.VisitJSON([]any{0}, opt))
	})

	t.Run("contains", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:     &openapi3.Types{"array"},
			Contains: positive(),
		}
		require.NoError(t, schema.VisitJSON([]any{0, 1}, opt))
		require.Error(t, schema.VisitJSON([]any{0}, opt))
	})

	t.Run("patternProperties", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:              &openapi3.Types{"object"},
			PatternProperties: openapi3.Schemas{"^x-": positive()},
		}
		require.NoError(t, schema.VisitJSON(map[string]any{"x-val": 1}, opt))
		require.Error(t, schema.VisitJSON(map[string]any{"x-val": 0}, opt))
	})

	t.Run("dependentSchemas", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			DependentSchemas: openapi3.Schemas{
				"name": &openapi3.SchemaRef{Value: &openapi3.Schema{
					Properties: openapi3.Schemas{"count": positive()},
				}},
			},
		}
		require.NoError(t, schema.VisitJSON(map[string]any{"name": "foo", "count": 1}, opt))
		require.Error(t, schema.VisitJSON(map[string]any{"name": "foo", "count": 0}, opt))
	})

	t.Run("propertyNames", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			PropertyNames: &openapi3.SchemaRef{Value: &openapi3.Schema{
				Type:      &openapi3.Types{"string"},
				MinLength: 1,
			}},
		}
		require.NoError(t, schema.VisitJSON(map[string]any{"abc": 1}, opt))
		require.Error(t, schema.VisitJSON(map[string]any{"": 1}, opt), "empty property name should fail minLength")
	})

	t.Run("unevaluatedItems", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"array"},
			PrefixItems: openapi3.SchemaRefs{
				&openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			},
			UnevaluatedItems: openapi3.BoolSchema{Schema: positive()},
		}
		require.NoError(t, schema.VisitJSON([]any{"a", 1}, opt))
		require.Error(t, schema.VisitJSON([]any{"a", 0}, opt))
	})

	t.Run("unevaluatedProperties", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"name": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			},
			UnevaluatedProperties: openapi3.BoolSchema{Schema: positive()},
		}
		require.NoError(t, schema.VisitJSON(map[string]any{"name": "foo", "extra": 1}, opt))
		require.Error(t, schema.VisitJSON(map[string]any{"name": "foo", "extra": 0}, opt))
	})

	t.Run("contentSchema", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:             &openapi3.Types{"string"},
			ContentMediaType: "application/json",
			ContentSchema:    positive(),
		}
		// contentSchema is an annotation in 2020-12: the transform must not crash on it.
		require.NoError(t, schema.VisitJSON("0", opt))
	})
}

func TestBuiltInValidatorStillWorks(t *testing.T) {
	t.Run("string validation with built-in", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"string"},
		}

		err := schema.VisitJSON("hello", openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(123, openapi3.EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("object validation with built-in", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"name": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			},
			Required: []string{"name"},
		}

		err := schema.VisitJSON(map[string]any{
			"name": "John",
		})
		require.NoError(t, err)

		err = schema.VisitJSON(map[string]any{}, openapi3.EnableJSONSchema2020())
		require.Error(t, err)
	})
}
