package openapi3

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue220(t *testing.T) {
	for _, specPath := range []string{
		"testdata/my-openapi.json",
		filepath.FromSlash("testdata/my-openapi.json"),
	} {
		t.Logf("specPath: %q", specPath)
		pwd, err := os.Getwd()
		require.NoError(t, err)
		t.Logf("pwd: %q", pwd)

		loader := NewSwaggerLoader()
		loader.IsExternalRefsAllowed = true
		doc, err := loader.LoadSwaggerFromFile(specPath)
		require.NoError(t, err)

		err = doc.Validate(loader.Context)
		require.NoError(t, err)

		require.Equal(t, "integer", doc.Paths["/foo"].Get.Responses["200"].Value.Content["application/json"].Schema.Value.Properties["bar"].Value.Type)
	}
}
