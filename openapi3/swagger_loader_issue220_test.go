package openapi3

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue220(t *testing.T) {
	specPath := filepath.FromSlash("testdata/my-openapi.json")

	loader := NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadSwaggerFromFile(specPath)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)
}
