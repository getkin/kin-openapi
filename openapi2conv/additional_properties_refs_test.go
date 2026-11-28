package openapi2conv_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

const additionalPropertiesRefsV2 = `{
  "Pet": {"type": "string"},
  "Direct": {
    "type": "object",
    "additionalProperties": {"$ref": "#/definitions/Pet"}
  },
  "Property": {
    "type": "object",
    "additionalProperties": {
      "type": "object",
      "properties": {"pet": {"$ref": "#/definitions/Pet"}}
    }
  },
  "ItemsProperty": {
    "type": "object",
    "additionalProperties": {
      "type": "array",
      "items": {"type": "object", "properties": {"pet": {"$ref": "#/definitions/Pet"}}}
    }
  },
  "AllOfProperty": {
    "type": "object",
    "additionalProperties": {
      "allOf": [{"type": "object", "properties": {"pet": {"$ref": "#/definitions/Pet"}}}]
    }
  },
  "Recursive": {
    "type": "object",
    "additionalProperties": {"$ref": "#/definitions/Recursive"}
  }
}`

const additionalPropertiesRefsV3 = `{
  "Pet": {"type": "string"},
  "Direct": {
    "type": "object",
    "additionalProperties": {"$ref": "#/components/schemas/Pet"}
  },
  "Property": {
    "type": "object",
    "additionalProperties": {
      "type": "object",
      "properties": {"pet": {"$ref": "#/components/schemas/Pet"}}
    }
  },
  "ItemsProperty": {
    "type": "object",
    "additionalProperties": {
      "type": "array",
      "items": {"type": "object", "properties": {"pet": {"$ref": "#/components/schemas/Pet"}}}
    }
  },
  "AllOfProperty": {
    "type": "object",
    "additionalProperties": {
      "allOf": [{"type": "object", "properties": {"pet": {"$ref": "#/components/schemas/Pet"}}}]
    }
  },
  "Recursive": {
    "type": "object",
    "additionalProperties": {"$ref": "#/components/schemas/Recursive"}
  }
}`

func TestToV3RefsUnderAdditionalProperties(t *testing.T) {
	t.Parallel()

	var doc2 openapi2.T
	err := json.Unmarshal([]byte(`{
  "swagger": "2.0",
  "info": {"title": "title", "version": "1"},
  "paths": {"/pets": {"get": {"responses": {"200": {"description": "OK"}}}}},
  "definitions": `+additionalPropertiesRefsV2+`
}`), &doc2)
	require.NoError(t, err)

	doc3, err := openapi2conv.ToV3(&doc2)
	require.NoError(t, err)
	require.NoError(t, doc3.Validate(t.Context()))

	schemas, err := json.Marshal(doc3.Components.Schemas)
	require.NoError(t, err)
	require.JSONEq(t, additionalPropertiesRefsV3, string(schemas))

	// The converted document keeps its own references.
	definitions, err := json.Marshal(doc2.Definitions)
	require.NoError(t, err)
	require.JSONEq(t, additionalPropertiesRefsV2, string(definitions))
}

func TestFromV3RefsUnderAdditionalProperties(t *testing.T) {
	t.Parallel()

	loader := openapi3.NewLoader()
	doc3, err := loader.LoadFromData([]byte(`{
  "openapi": "3.0.3",
  "info": {"title": "title", "version": "1"},
  "paths": {},
  "components": {"schemas": ` + additionalPropertiesRefsV3 + `}
}`))
	require.NoError(t, err)
	require.NoError(t, doc3.Validate(t.Context()))

	doc2, err := openapi2conv.FromV3(doc3)
	require.NoError(t, err)

	definitions, err := json.Marshal(doc2.Definitions)
	require.NoError(t, err)
	require.JSONEq(t, additionalPropertiesRefsV2, string(definitions))

	// The converted document keeps its own references.
	schemas, err := json.Marshal(doc3.Components.Schemas)
	require.NoError(t, err)
	require.JSONEq(t, additionalPropertiesRefsV3, string(schemas))
}
