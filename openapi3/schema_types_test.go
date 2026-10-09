package openapi3_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestTypes_HelperMethods(t *testing.T) {
	t.Parallel()

	t.Run("IncludesNull", func(t *testing.T) {
		// Single type without null
		types := &openapi3.Types{"string"}
		require.False(t, types.IncludesNull())

		// Type array with null
		types = &openapi3.Types{"string", "null"}
		require.True(t, types.IncludesNull())

		// Multiple types without null
		types = &openapi3.Types{"string", "number"}
		require.False(t, types.IncludesNull())

		// Nil types
		var nilTypes *openapi3.Types
		require.False(t, nilTypes.IncludesNull())
	})

	t.Run("IsMultiple", func(t *testing.T) {
		// Single type
		types := &openapi3.Types{"string"}
		require.False(t, types.IsMultiple())

		// Multiple types
		types = &openapi3.Types{"string", "null"}
		require.True(t, types.IsMultiple())

		types = &openapi3.Types{"string", "number", "null"}
		require.True(t, types.IsMultiple())

		// Empty types
		types = &openapi3.Types{}
		require.False(t, types.IsMultiple())

		// Nil types
		var nilTypes *openapi3.Types
		require.False(t, nilTypes.IsMultiple())
	})

	t.Run("IsSingle", func(t *testing.T) {
		// Single type
		types := &openapi3.Types{"string"}
		require.True(t, types.IsSingle())

		// Multiple types
		types = &openapi3.Types{"string", "null"}
		require.False(t, types.IsSingle())

		// Empty types
		types = &openapi3.Types{}
		require.False(t, types.IsSingle())

		// Nil types
		var nilTypes *openapi3.Types
		require.False(t, nilTypes.IsSingle())
	})

	t.Run("IsEmpty", func(t *testing.T) {
		// Single type
		types := &openapi3.Types{"string"}
		require.False(t, types.IsEmpty())

		// Multiple types
		types = &openapi3.Types{"string", "null"}
		require.False(t, types.IsEmpty())

		// Empty types
		types = &openapi3.Types{}
		require.True(t, types.IsEmpty())

		// Nil types
		var nilTypes *openapi3.Types
		require.True(t, nilTypes.IsEmpty())
	})
}

func TestTypes_ArraySerialization(t *testing.T) {
	t.Parallel()

	t.Run("single type serializes as string", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"string"},
		}

		data, err := json.Marshal(schema)
		require.NoError(t, err)

		// Should serialize as "type": "string" (not array)
		require.Contains(t, string(data), `"type":"string"`)
		require.NotContains(t, string(data), `"type":["string"]`)
	})

	t.Run("multiple types serialize as array", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"string", "null"},
		}

		data, err := json.Marshal(schema)
		require.NoError(t, err)

		// Should serialize as "type": ["string", "null"]
		require.Contains(t, string(data), `"type":["string","null"]`)
	})

	t.Run("deserialize string to single type", func(t *testing.T) {
		jsonData := []byte(`{"type":"string"}`)

		var schema openapi3.Schema
		err := json.Unmarshal(jsonData, &schema)
		require.NoError(t, err)

		require.NotNil(t, schema.Type)
		require.True(t, schema.Type.IsSingle())
		require.True(t, schema.Type.Is("string"))
	})

	t.Run("deserialize array to multiple types", func(t *testing.T) {
		jsonData := []byte(`{"type":["string","null"]}`)

		var schema openapi3.Schema
		err := json.Unmarshal(jsonData, &schema)
		require.NoError(t, err)

		require.NotNil(t, schema.Type)
		require.True(t, schema.Type.IsMultiple())
		require.True(t, schema.Type.Includes("string"))
		require.True(t, schema.Type.IncludesNull())
	})
}

func TestTypes_OpenAPI31Features(t *testing.T) {
	t.Parallel()

	t.Run("type array with null", func(t *testing.T) {
		types := &openapi3.Types{"string", "null"}

		require.True(t, types.Includes("string"))
		require.True(t, types.IncludesNull())
		require.True(t, types.IsMultiple())
		require.False(t, types.IsSingle())
		require.False(t, types.IsEmpty())

		// Test Permits
		require.True(t, types.Permits("string"))
		require.True(t, types.Permits("null"))
		require.False(t, types.Permits("number"))
	})

	t.Run("type array without null", func(t *testing.T) {
		types := &openapi3.Types{"string", "number"}

		require.True(t, types.Includes("string"))
		require.True(t, types.Includes("number"))
		require.False(t, types.IncludesNull())
		require.True(t, types.IsMultiple())
	})

	t.Run("OpenAPI 3.0 style single type", func(t *testing.T) {
		types := &openapi3.Types{"string"}

		require.True(t, types.Is("string"))
		require.True(t, types.Includes("string"))
		require.False(t, types.IncludesNull())
		require.False(t, types.IsMultiple())
		require.True(t, types.IsSingle())
	})
}

func TestTypes_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("nil types permits everything", func(t *testing.T) {
		var types *openapi3.Types

		require.True(t, types.Permits("string"))
		require.True(t, types.Permits("number"))
		require.True(t, types.Permits("null"))
		require.True(t, types.IsEmpty())
	})

	t.Run("empty slice of types", func(t *testing.T) {
		types := &openapi3.Types{}

		require.False(t, types.Includes("string"))
		require.False(t, types.Permits("string"))
		require.True(t, types.IsEmpty())
		require.False(t, types.IsSingle())
		require.False(t, types.IsMultiple())
	})

	t.Run("Slice method", func(t *testing.T) {
		types := &openapi3.Types{"string", "null"}
		slice := types.Slice()

		require.Equal(t, []string{"string", "null"}, slice)

		// Nil types
		var nilTypes *openapi3.Types
		require.Nil(t, nilTypes.Slice())
	})
}

func TestTypes_BackwardCompatibility(t *testing.T) {
	t.Parallel()

	t.Run("existing Is method still works", func(t *testing.T) {
		// Single type
		types := &openapi3.Types{"string"}
		require.True(t, types.Is("string"))
		require.False(t, types.Is("number"))

		// Multiple types - Is should return false
		types = &openapi3.Types{"string", "null"}
		require.False(t, types.Is("string"))
		require.False(t, types.Is("null"))
	})

	t.Run("existing Includes method still works", func(t *testing.T) {
		types := &openapi3.Types{"string"}
		require.True(t, types.Includes("string"))
		require.False(t, types.Includes("number"))

		types = &openapi3.Types{"string", "null"}
		require.True(t, types.Includes("string"))
		require.True(t, types.Includes("null"))
		require.False(t, types.Includes("number"))
	})

	t.Run("existing Permits method still works", func(t *testing.T) {
		// Nil types permits everything
		var types *openapi3.Types
		require.True(t, types.Permits("anything"))

		// Specific types
		types = &openapi3.Types{"string"}
		require.True(t, types.Permits("string"))
		require.False(t, types.Permits("number"))
	})
}

func TestTypes_CloneWithWithout(t *testing.T) {
	t.Parallel()

	t.Run("Clone", func(t *testing.T) {
		var nilTypes *openapi3.Types
		require.Nil(t, nilTypes.Clone())

		types := &openapi3.Types{"string"}
		clone := types.Clone()
		require.Equal(t, openapi3.Types{"string"}, clone)
		clone[0] = "number"
		require.Equal(t, &openapi3.Types{"string"}, types)
	})

	t.Run("With", func(t *testing.T) {
		nilTypes := &openapi3.Schema{}
		nilTypes.WithTypes("object")
		require.Equal(t, &openapi3.Types{"object"}, nilTypes.Type)

		types := &openapi3.Types{"string"}
		types.With("null", "string", "null")
		require.Equal(t, &openapi3.Types{"string", "null"}, types)

		types = &openapi3.Types{"null", "string", "null"}
		types.With("string")
		require.Equal(t, &openapi3.Types{"null", "string"}, types)

		types = &openapi3.Types{}
		types.With("integer")
		require.Equal(t, &openapi3.Types{"integer"}, types)

		// Spare capacity shared with another slice is not written to.
		backing := make([]string /*,*/, 1, 2)
		backing[0] = "string"
		other := openapi3.Types(backing[:2])
		types = &openapi3.Types{}
		*types = backing[:1]
		types.With("null")
		require.Equal(t, &openapi3.Types{"string", "null"}, types)
		require.Equal(t, openapi3.Types{"string", ""}, other)
	})

	t.Run("Without", func(t *testing.T) {
		var types *openapi3.Types
		types.Without("object")
		require.Nil(t, types)

		types = &openapi3.Types{"string", "null", "number", "number"}
		types.Without("null", "number", "boolean")
		require.Equal(t, &openapi3.Types{"string"}, types)

		types = &openapi3.Types{"string", "null", "number"}
		types.Without("null", "number", "boolean", "boolean")
		require.Equal(t, &openapi3.Types{"string"}, types)

		types = &openapi3.Types{"null", "string", "null", "string"}
		types.Without("boolean")
		require.Equal(t, &openapi3.Types{"null", "string"}, types)

		// The backing array of another slice is not modified.
		other := openapi3.Types{"null", "string"}
		types = &openapi3.Types{}
		*types = other
		types.Without("null")
		require.Equal(t, &openapi3.Types{"string"}, types)
		require.Equal(t, openapi3.Types{"null", "string"}, other)
	})
}

func TestSchema_WithTypes(t *testing.T) {
	t.Parallel()

	schema := &openapi3.Schema{}
	require.Same(t, schema, schema.WithTypes("string", "null", "string"))
	require.Equal(t, &openapi3.Types{"string", "null"}, schema.Type)

	schema.WithTypes("null", "integer")
	require.Equal(t, &openapi3.Types{"string", "null", "integer"}, schema.Type)
}
