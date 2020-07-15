package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue235(t *testing.T) {
	loader := NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadSwaggerFromFile("testdata/issue235.spec0.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
}
