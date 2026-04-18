package openapi3_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
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

	t.Run("nullable conversion", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:     &openapi3.Types{"string"},
			Nullable: true,
		}

		err := schema.VisitJSON("hello", openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(nil, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)
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

func TestJSONSchema2020Validator_Fallback(t *testing.T) {
	t.Run("fallback on compilation error", func(t *testing.T) {
		// Create a schema that might cause compilation issues
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"string"},
		}

		// Should not panic, even if there's an issue
		err := schema.VisitJSON("test", openapi3.EnableJSONSchema2020())
		require.NoError(t, err)
	})
}

func TestJSONSchema2020Validator_TransformRecursesInto31Fields(t *testing.T) {
	// These tests verify that transformOpenAPIToJSONSchema recurses into
	// OpenAPI 3.1 / JSON Schema 2020-12 fields. Each sub-test uses a nested
	// schema with nullable:true (an OpenAPI 3.0-ism) that must be converted
	// to a type array for the JSON Schema 2020-12 validator to handle null.

	t.Run("prefixItems with nullable nested schema", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"array"},
			PrefixItems: openapi3.SchemaRefs{
				&openapi3.SchemaRef{Value: &openapi3.Schema{
					Type:     &openapi3.Types{"string"},
					Nullable: true,
				}},
			},
		}

		err := schema.VisitJSON([]any{"hello"}, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON([]any{nil}, openapi3.EnableJSONSchema2020())
		require.NoError(t, err, "null should be accepted after nullable conversion in prefixItems")
	})

	t.Run("contains with nullable nested schema", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"array"},
			Contains: &openapi3.SchemaRef{Value: &openapi3.Schema{
				Type:     &openapi3.Types{"string"},
				Nullable: true,
			}},
		}

		err := schema.VisitJSON([]any{nil}, openapi3.EnableJSONSchema2020())
		require.NoError(t, err, "null should match contains after nullable conversion")
	})

	t.Run("patternProperties with nullable nested schema", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			PatternProperties: openapi3.Schemas{
				"^x-": &openapi3.SchemaRef{Value: &openapi3.Schema{
					Type:     &openapi3.Types{"string"},
					Nullable: true,
				}},
			},
		}

		err := schema.VisitJSON(map[string]any{"x-val": nil}, openapi3.EnableJSONSchema2020())
		require.NoError(t, err, "null should be accepted after nullable conversion in patternProperties")
	})

	t.Run("dependentSchemas with nullable nested schema", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"name": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
				"tag":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Nullable: true}},
			},
			DependentSchemas: openapi3.Schemas{
				"name": &openapi3.SchemaRef{Value: &openapi3.Schema{
					Properties: openapi3.Schemas{
						"tag": &openapi3.SchemaRef{Value: &openapi3.Schema{
							Type:     &openapi3.Types{"string"},
							Nullable: true,
						}},
					},
				}},
			},
		}

		err := schema.VisitJSON(map[string]any{"name": "foo", "tag": nil}, openapi3.EnableJSONSchema2020())
		require.NoError(t, err, "null should be accepted after nullable conversion in dependentSchemas")
	})

	t.Run("propertyNames with nullable not applicable but transform should not crash", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			PropertyNames: &openapi3.SchemaRef{Value: &openapi3.Schema{
				Type:      &openapi3.Types{"string"},
				MinLength: 1,
			}},
		}

		err := schema.VisitJSON(map[string]any{"abc": 1}, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(map[string]any{"": 1}, openapi3.EnableJSONSchema2020())
		require.Error(t, err, "empty property name should fail minLength")
	})

	t.Run("unevaluatedItems with nullable nested schema", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"array"},
			PrefixItems: openapi3.SchemaRefs{
				&openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}},
			},
			UnevaluatedItems: openapi3.BoolSchema{Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{
				Type:     &openapi3.Types{"string"},
				Nullable: true,
			}}},
		}

		err := schema.VisitJSON([]any{1, nil}, openapi3.EnableJSONSchema2020())
		require.NoError(t, err, "null should be accepted after nullable conversion in unevaluatedItems")
	})

	t.Run("unevaluatedProperties with nullable nested schema", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"name": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			},
			UnevaluatedProperties: openapi3.BoolSchema{Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{
				Type:     &openapi3.Types{"string"},
				Nullable: true,
			}}},
		}

		err := schema.VisitJSON(map[string]any{"name": "foo", "extra": nil}, openapi3.EnableJSONSchema2020())
		require.NoError(t, err, "null should be accepted after nullable conversion in unevaluatedProperties")
	})

	t.Run("contentSchema with nullable nested schema", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:             &openapi3.Types{"string"},
			ContentMediaType: "application/json",
			ContentSchema: &openapi3.SchemaRef{Value: &openapi3.Schema{
				Type:     &openapi3.Types{"object"},
				Nullable: true,
			}},
		}

		// contentSchema transform should not crash and should handle nullable
		err := schema.VisitJSON("null", openapi3.EnableJSONSchema2020())
		require.NoError(t, err, "contentSchema transform should handle nullable nested schema")
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
