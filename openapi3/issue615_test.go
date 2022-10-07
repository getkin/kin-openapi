package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue615(t *testing.T) {
	for {
		loader := openapi3.NewLoader()
		loader.IsExternalRefsAllowed = true
		_, err := loader.LoadFromFile("testdata/recursiveRef/issue615.yml")
		if err == nil {
			continue
		}
		// Test currently reproduces the issue 615: failure to load a valid spec
		// Upon issue resolution, this check should be changed to require.NoError
		require.Error(t, err, openapi3.CircularReferenceError)
		break
	}

	var old int
	old, openapi3.CircularReferenceCounter = openapi3.CircularReferenceCounter, 4
	defer func() { openapi3.CircularReferenceCounter = old }()

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	_, err := loader.LoadFromFile("testdata/recursiveRef/issue615.yml")
	require.NoError(t, err)
}
