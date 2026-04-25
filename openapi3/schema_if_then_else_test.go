package openapi3_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestSchemaIfThenElse_BuiltInValidator(t *testing.T) {
	t.Run("schema with if/then/else is not empty", func(t *testing.T) {
		schema := &openapi3.Schema{
			If:   &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			Then: &openapi3.SchemaRef{Value: &openapi3.Schema{MinLength: 3}},
			Else: &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"number"}}},
		}
		require.False(t, schema.IsEmpty())
	})

	t.Run("schema with dependentRequired is not empty", func(t *testing.T) {
		schema := &openapi3.Schema{
			DependentRequired: map[string][]string{
				"creditCard": {"billingAddress"},
			},
		}
		require.False(t, schema.IsEmpty())
	})
}

func TestSchemaIfThenElse_JSONSchema2020(t *testing.T) {
	t.Run("if/then/else conditional validation", func(t *testing.T) {
		// If type is string, then minLength=3; else must be number
		schema := &openapi3.Schema{
			If:   &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			Then: &openapi3.SchemaRef{Value: &openapi3.Schema{MinLength: 3}},
			Else: &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"number"}}},
		}

		// String with length >= 3 → passes if+then
		err := schema.VisitJSON("hello", openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		// Number → fails if, passes else
		err = schema.VisitJSON(float64(42), openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		// Short string → passes if, fails then
		err = schema.VisitJSON("ab", openapi3.EnableJSONSchema2020())
		require.Error(t, err)

		// Boolean → fails if, fails else
		err = schema.VisitJSON(true, openapi3.EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("if/then without else", func(t *testing.T) {
		// If type is string, then minLength=5
		schema := &openapi3.Schema{
			If:   &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			Then: &openapi3.SchemaRef{Value: &openapi3.Schema{MinLength: 5}},
		}

		// String with length >= 5 → passes
		err := schema.VisitJSON("hello", openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		// Short string → fails then
		err = schema.VisitJSON("hi", openapi3.EnableJSONSchema2020())
		require.Error(t, err)

		// Number → fails if, no else so passes
		err = schema.VisitJSON(float64(42), openapi3.EnableJSONSchema2020())
		require.NoError(t, err)
	})

	t.Run("dependentRequired validation", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"name":           &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
				"creditCard":     &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
				"billingAddress": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			},
			DependentRequired: map[string][]string{
				"creditCard": {"billingAddress"},
			},
		}

		// Has creditCard and billingAddress → passes
		err := schema.VisitJSON(map[string]any{
			"name":           "John",
			"creditCard":     "1234",
			"billingAddress": "123 Main St",
		}, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		// No creditCard → passes (dependency not triggered)
		err = schema.VisitJSON(map[string]any{
			"name": "John",
		}, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		// Has creditCard but no billingAddress → fails
		err = schema.VisitJSON(map[string]any{
			"name":       "John",
			"creditCard": "1234",
		}, openapi3.EnableJSONSchema2020())
		require.Error(t, err)
	})
}

func TestSchemaIfThenElse_MarshalRoundTrip(t *testing.T) {
	t.Run("if/then/else round-trip", func(t *testing.T) {
		schema := &openapi3.Schema{
			If:   &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			Then: &openapi3.SchemaRef{Value: &openapi3.Schema{MinLength: 3}},
			Else: &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"number"}}},
		}

		data, err := schema.MarshalJSON()
		require.NoError(t, err)

		var roundTripped openapi3.Schema
		err = roundTripped.UnmarshalJSON(data)
		require.NoError(t, err)

		require.NotNil(t, roundTripped.If)
		require.NotNil(t, roundTripped.Then)
		require.NotNil(t, roundTripped.Else)
		require.True(t, roundTripped.If.Value.Type.Is("string"))
		require.Equal(t, uint64(3), roundTripped.Then.Value.MinLength)
		require.True(t, roundTripped.Else.Value.Type.Is("number"))
	})

	t.Run("dependentRequired round-trip", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			DependentRequired: map[string][]string{
				"creditCard": {"billingAddress", "cvv"},
			},
		}

		data, err := schema.MarshalJSON()
		require.NoError(t, err)

		var roundTripped openapi3.Schema
		err = roundTripped.UnmarshalJSON(data)
		require.NoError(t, err)

		require.Equal(t, map[string][]string{
			"creditCard": {"billingAddress", "cvv"},
		}, roundTripped.DependentRequired)
	})

	t.Run("no extensions leak", func(t *testing.T) {
		schema := &openapi3.Schema{
			If:                &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			DependentRequired: map[string][]string{"a": {"b"}},
		}

		data, err := schema.MarshalJSON()
		require.NoError(t, err)

		var roundTripped openapi3.Schema
		err = roundTripped.UnmarshalJSON(data)
		require.NoError(t, err)

		// if/then/else/dependentRequired should not leak into extensions
		require.Nil(t, roundTripped.Extensions)
	})
}

func TestSchemaIfThenElse_Validate(t *testing.T) {
	t.Run("unresolved if ref fails validation", func(t *testing.T) {
		schema := &openapi3.Schema{
			If: &openapi3.SchemaRef{Ref: "#/components/schemas/Missing"},
		}
		err := schema.Validate(t.Context(), openapi3.IsOpenAPI31OrLater())
		require.Error(t, err)
		require.ErrorContains(t, err, "unresolved ref")
	})

	t.Run("unresolved then ref fails validation", func(t *testing.T) {
		schema := &openapi3.Schema{
			Then: &openapi3.SchemaRef{Ref: "#/components/schemas/Missing"},
		}
		err := schema.Validate(t.Context(), openapi3.IsOpenAPI31OrLater())
		require.Error(t, err)
		require.ErrorContains(t, err, "unresolved ref")
	})

	t.Run("unresolved else ref fails validation", func(t *testing.T) {
		schema := &openapi3.Schema{
			Else: &openapi3.SchemaRef{Ref: "#/components/schemas/Missing"},
		}
		err := schema.Validate(t.Context(), openapi3.IsOpenAPI31OrLater())
		require.Error(t, err)
		require.ErrorContains(t, err, "unresolved ref")
	})

	t.Run("valid if/then/else passes validation", func(t *testing.T) {
		schema := &openapi3.Schema{
			If:   &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			Then: &openapi3.SchemaRef{Value: &openapi3.Schema{MinLength: 1}},
			Else: &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"number"}}},
		}
		err := schema.Validate(t.Context(), openapi3.IsOpenAPI31OrLater())
		require.NoError(t, err)
	})
}
