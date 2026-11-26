package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// TestIssue958: a property declared by an "allOf" subschema must not be
// rejected by "additionalProperties: false" on the composed schema.
func TestIssue958(t *testing.T) {
	t.Parallel()

	spec := []byte(`
openapi: 3.0.0
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
paths:
  "/pets":
    post:
      requestBody:
        content:
          application/json:
            schema:
              "$ref": "#/components/schemas/PetWithCategory"
        required: true
      responses:
        '201':
          description: Null response
components:
  schemas:
    PetCategory:
      type: object
      properties:
        category:
          type: string
    PetWithCategory:
      allOf:
      - "$ref": "#/components/schemas/PetCategory"
      - type: object
        properties:
          pet:
            "$ref": "#/components/schemas/Pet"
      additionalProperties: false
    Pet:
      type: object
      required:
      - id
      - name
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
        tag:
          type: string
`)

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	schema := doc.Components.Schemas["PetWithCategory"].Value

	valid := []map[string]any{
		{"category": "test"},
		{"pet": map[string]any{"id": 1, "name": "rex"}},
		{"category": "test", "pet": map[string]any{"id": 1, "name": "rex"}},
	}
	for _, value := range valid {
		require.NoError(t, schema.VisitJSON(value, openapi3.VisitAsRequest()), "value %v", value)
	}

	// A property declared by no subschema is still unsupported.
	err = schema.VisitJSON(map[string]any{"nope": 1}, openapi3.VisitAsRequest())
	require.Error(t, err)
	require.ErrorContains(t, err, `property "nope" is unsupported`)

	// A property declared by a subschema is still validated against it.
	err = schema.VisitJSON(map[string]any{"category": 42}, openapi3.VisitAsRequest())
	require.Error(t, err)
}
