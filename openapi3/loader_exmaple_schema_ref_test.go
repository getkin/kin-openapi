package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadSpecWithExampleRefToSchemas(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromFile("testdata/example-schemas-ref.json")
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	exampleValue := doc.Paths["/pet"].Put.Responses.Get(400).Value.Content.Get("*/*").Examples["200"].Value.Value
	schemaRef := doc.Components.Schemas["Pet"].Value
	require.Equal(t, exampleValue, schemaRef)
	require.NoError(t, err)
}
