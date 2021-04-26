package openapi3

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue212(t *testing.T) {
	spec := `
openapi: 3.0.1
info:
  title: 'test'
  version: 1.0.0
servers:
  - url: /api

paths:
  /available-products:
    get:
      operationId: getAvailableProductCollection
      responses:
        "200":
          description: test
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/AvailableProduct"

components:
  schemas:
    AvailableProduct:
      type: object
      properties:
        id:
          type: string
        type:
          type: string
        name:
          type: string
        media:
          type: object
          properties:
            documents:
              type: array
              items:
                allOf:
                  - $ref: "#/components/schemas/AvailableProduct/properties/previewImage/allOf/0"
                  - type: object
                    properties:
                      uri:
                        type: string
                        pattern: ^\/documents\/[0-9a-f]{64}$
        previewImage:
          allOf:
            - type: object
              required:
                - id
                - uri
              properties:
                id:
                  type: string
                uri:
                  type: string
            - type: object
              properties:
                uri:
                  type: string
                  pattern: ^\/images\/[0-9a-f]{64}$
`

	loader := NewSwaggerLoader()
	doc, err := loader.LoadSwaggerFromData([]byte(spec))
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	expected, err := json.Marshal(&Schema{
		Type:     "object",
		Required: []string{"id", "uri"},
		Properties: map[string]*SchemaRef{
			"id":  {Value: &Schema{Type: "string"}},
			"uri": {Value: &Schema{Type: "string"}},
		},
	},
	)
	require.NoError(t, err)
	got, err := json.Marshal(doc.Components.Schemas["AvailableProduct"].Value.Properties["media"].Value.Properties["documents"].Value.Items.Value.AllOf[0].Value)
	require.NoError(t, err)

	require.Equal(t, expected, got)
}
