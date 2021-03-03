package openapi3

import (
	"io/ioutil"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoaderReadFromURIFunc(t *testing.T) {
	loader := NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *SwaggerLoader, url *url.URL) ([]byte, error) {
		return ioutil.ReadFile(filepath.Join("testdata", url.Path))
	}
	doc, err := loader.LoadSwaggerFromFile("recursiveRef/openapi.yml")
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.NoError(t, doc.Validate(loader.Context))
	require.Equal(t, "bar", doc.Paths["/foo"].Get.Responses.Get(200).Value.Content.Get("application/json").Schema.Value.Properties["foo"].Value.Properties["bar"].Value.Items.Value.Example)
}
