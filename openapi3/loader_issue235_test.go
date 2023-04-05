package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue235OK(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile("testdata/issue235.spec0.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
}

func TestIssue235CircularDep(t *testing.T) {
	t.Skip("TODO: return an error on circular dependencies between external files of a spec")
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile("testdata/issue235.spec0-typo.yml")
	require.Nil(t, doc)
	require.Error(t, err)
}
