package openapi3

import (
  "testing"

  "github.com/stretchr/testify/require"
)

func TestIssue376(t *testing.T) {
  spec := []byte(`
openapi: 3.0.0
components:
  schemas:
    schema1:
      type: object
      additionalProperties:
        type: string
    schema2:
      type: object
      properties:
        prop:
          $ref: '#/components/schemas/schema1/additionalProperties'
`)

  loader := NewLoader()
  doc, err := loader.LoadFromData(spec)
  require.NoError(t, err)
  // require.Equal(t, "An API", doc.Info.Title)
  require.Equal(t, 2, len(doc.Components.Schemas))
  // require.Equal(t, 1, len(doc.Paths))
  // def := doc.Paths["/items"].Put.Responses.Default().Value
  // desc := "unexpected error"
  // require.Equal(t, &desc, def.Description)
  err = doc.Validate(loader.Context)
  require.NoError(t, err)
}
