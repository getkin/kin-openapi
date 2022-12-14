package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue615(t *testing.T) {
	{
		loader := openapi3.NewLoader()
		loader.IsExternalRefsAllowed = true
		_, err := loader.LoadFromFile("testdata/recursiveRef/issue615.yml")
		require.NoError(t, err)
	}

	var old int
	old, openapi3.CircularReferenceCounter = openapi3.CircularReferenceCounter, 4
	defer func() { openapi3.CircularReferenceCounter = old }()

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	_, err := loader.LoadFromFile("testdata/recursiveRef/issue615.yml")
	require.NoError(t, err)
}
