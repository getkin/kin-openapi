package openapi3

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue220(t *testing.T) {
	specPath := filepath.FromSlash("testdata/my-openapi.json")
	t.Logf("specPath: %q", specPath)

	loader := NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadSwaggerFromFile(specPath)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	require.Equal(t, "integer", doc.Paths["/foo"].Get.Responses["200"].Value.Content["application/json"].Schema.Value.Properties["bar"].Value.Type)
}
