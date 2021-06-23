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
paths: {}
info:
  title: An API
  version: 1.2.3.4
`)

	loader := NewLoader()

	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	require.Equal(t, "An API", doc.Info.Title)
	require.Equal(t, 2, len(doc.Components.Schemas))
	require.Equal(t, 0, len(doc.Paths))

	require.Equal(t, "string", doc.Components.Schemas["schema2"].Value.Properties["prop"].Value.Type)
}
