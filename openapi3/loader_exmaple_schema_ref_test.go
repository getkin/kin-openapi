package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadSpecWithExampleRefToSchemas(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	api, err := loader.LoadFromFile("testdata/example-schemas-ref.json")
	exampleValue := api.Paths["/pet"].Put.Responses.Get(400).Value.Content.Get("*/*").Examples["200"].Value.Value
	schemaRef := api.Components.Schemas["Pet"].Value
	require.Equal(t, exampleValue, schemaRef)
	require.NoError(t, err)
}
