package openapi3

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoaderSupportsHTTPFileSystem(t *testing.T) {
	loader := NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadSwaggerFromFileSystem(http.Dir("testdata/httpFileSystem"), "openapi.yml")
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.NoError(t, doc.Validate(loader.Context))
	require.Equal(t, "foo", doc.Paths["/foo"].Get.Responses.Get(200).Value.Content.Get("application/json").Schema.Value.Properties["foo"].Value.Example)
}
