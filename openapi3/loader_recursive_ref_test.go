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
	require.NotNil(t, doc)
	require.NoError(t, doc.Validate(loader.Context))
	require.Equal(t, "bar", doc.Paths["/foo"].Get.Responses.Get(200).Value.Content.Get("application/json").Schema.Value.Properties["foo2"].Value.Properties["foo"].Value.Properties["bar"].Value.Example)
}
