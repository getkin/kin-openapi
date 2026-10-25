package openapi3_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func loadSchemaRefSiblings(t *testing.T, version, schemas string) *openapi3.T {
	t.Helper()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(fmt.Appendf(nil, `{
  "openapi": %q,
  "info": {"title": "t", "version": "1"},
  "paths": {},
  "components": {"schemas": %s}
}`, version, schemas))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(t.Context()))
	return doc
}

// In OpenAPI 3.1 a Schema Object is a JSON Schema 2020-12 schema, in which the
// keywords next to $ref apply alongside it: a value must satisfy both.
func TestSchemaRefSiblingsApplyAlongsideTheRef(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		sibling string
		valid   []any
		invalid []any
	}{{
		name:    "a looser bound does not relax the referenced one",
		sibling: `"maxLength": 10`,
		valid:   []any{"abc"},
		invalid: []any{"abcd", "abcdefghijk"},
	}, {
		name:    "a tighter bound applies",
		sibling: `"minLength": 2`,
		valid:   []any{"ab"},
		invalid: []any{"a", "abcd"},
	}, {
		name:    "a different type matches nothing",
		sibling: `"type": "integer"`,
		invalid: []any{"a", 1.0},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			doc := loadSchemaRefSiblings(t, "3.1.0", `{
  "Base": {"type": "string", "maxLength": 3},
  "S": {"properties": {"p": {"$ref": "#/components/schemas/Base", `+tc.sibling+`}}}
}`)
			value := doc.Components.Schemas["S"].Value.Properties["p"].Value
			for _, v := range tc.valid {
				require.NoError(t, value.VisitJSON(v), "%v", v)
			}
			for _, v := range tc.invalid {
				require.Error(t, value.VisitJSON(v), "%v", v)
			}
			// The referenced schema keeps its own shape, and is not changed.
			require.True(t, value.Type.Is("string"))
			require.Equal(t, uint64(3), *value.MaxLength)
			require.Len(t, doc.Components.Schemas["Base"].Value.AllOf, 0)
		})
	}

	t.Run("properties and required add to the referenced ones", func(t *testing.T) {
		t.Parallel()
		doc := loadSchemaRefSiblings(t, "3.1.0", `{
  "Base": {"type": "object", "properties": {"a": {"type": "string"}}, "required": ["a"]},
  "S": {"properties": {"p": {
    "$ref": "#/components/schemas/Base",
    "properties": {"b": {"$ref": "#/components/schemas/Int"}},
    "required": ["b"]
  }}},
  "Int": {"type": "integer"}
}`)
		value := doc.Components.Schemas["S"].Value.Properties["p"].Value
		require.NoError(t, value.VisitJSON(map[string]any{"a": "x", "b": 1.0}))
		require.Error(t, value.VisitJSON(map[string]any{"a": "x"}))
		require.Error(t, value.VisitJSON(map[string]any{"b": 1.0}))
		require.Error(t, value.VisitJSON(map[string]any{"a": "x", "b": "y"}))
		require.Contains(t, value.Properties, "a")
	})

	t.Run("annotations override the referenced ones", func(t *testing.T) {
		t.Parallel()
		doc := loadSchemaRefSiblings(t, "3.1.0", `{
  "Base": {"type": "object", "description": "base", "properties": {"a": {"type": "string"}}},
  "S": {"properties": {"p": {"$ref": "#/components/schemas/Base", "description": "local", "readOnly": true}}}
}`)
		value := doc.Components.Schemas["S"].Value.Properties["p"].Value
		require.Equal(t, "local", value.Description)
		require.True(t, value.ReadOnly)
		require.Contains(t, value.Properties, "a")
		require.Empty(t, value.AllOf)
		require.Equal(t, "base", doc.Components.Schemas["Base"].Value.Description)
	})

	t.Run("through a cycle", func(t *testing.T) {
		t.Parallel()
		doc := loadSchemaRefSiblings(t, "3.1.0", `{
  "Author": {"type": "object", "properties": {"name": {"type": "string"}, "book": {"$ref": "#/components/schemas/Book", "required": ["title"]}}},
  "Book": {"type": "object", "properties": {"title": {"type": "string"}, "author": {"$ref": "#/components/schemas/Author", "required": ["name"]}}}
}`)
		book := doc.Components.Schemas["Author"].Value.Properties["book"].Value
		require.Error(t, book.VisitJSON(map[string]any{}))
		require.NoError(t, book.VisitJSON(map[string]any{"title": "t"}))
		author := doc.Components.Schemas["Book"].Value.Properties["author"].Value
		require.Error(t, author.VisitJSON(map[string]any{}))
		require.NoError(t, author.VisitJSON(map[string]any{"name": "n"}))
	})
}

// In OpenAPI 3.0 a $ref is a Reference Object, whose added properties are
// ignored.
func TestSchemaRefSiblingsAreIgnoredIn30(t *testing.T) {
	t.Parallel()
	doc := loadSchemaRefSiblings(t, "3.0.3", `{
  "Base": {"type": "string", "maxLength": 3, "description": "base"},
  "S": {"properties": {"p": {"$ref": "#/components/schemas/Base", "maxLength": 1, "description": "local"}}}
}`)
	value := doc.Components.Schemas["S"].Value.Properties["p"].Value
	require.Same(t, doc.Components.Schemas["Base"].Value, value)
	require.NoError(t, value.VisitJSON("abc"))
}
