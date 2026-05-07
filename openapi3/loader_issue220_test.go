package openapi3_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue220(t *testing.T) {
	for _, specPath := range []string{
		"testdata/my-openapi.json",
		filepath.FromSlash("testdata/my-openapi.json"),
	} {
		t.Logf("specPath: %q", specPath)

		loader := openapi3.NewLoader()
		loader.IsExternalRefsAllowed = true
		doc, err := loader.LoadFromFile(specPath)
		require.NoError(t, err)

		err = doc.Validate(loader.Context)
		require.NoError(t, err)

		require.Equal(t, &openapi3.Types{"integer"}, doc.
			Paths.Value("/foo").
			Get.Responses.Value("200").Value.
			Content["application/json"].
			Schema.Value.Properties["bar"].Value.
			Type)
	}
}
