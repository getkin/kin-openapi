package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoaderSupportsRecursiveReference(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile("testdata/recursiveRef/openapi.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	require.Equal(t, "bar", doc.Paths["/foo"].Get.Responses.Get(200).Value.Content.Get("application/json").Schema.Value.Properties["foo2"].Value.Properties["foo"].Value.Properties["bar"].Value.Example)
}

func TestIssue447(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.0.1
info:
  title: Recursive refs example
  version: "1.0"
paths: {}
components:
  schemas:
    Complex:
      type: object
      properties:
        parent:
          $ref: '#/components/schemas/Complex'
`))
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	require.Equal(t, "object", doc.Components.
		// Complex
		Schemas["Complex"].
		// parent
		Value.Properties["parent"].
		// parent
		Value.Properties["parent"].
		// parent
		Value.Properties["parent"].
		// type
		Value.Type)
}
