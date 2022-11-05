package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue652(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// Test checks that no slice bounds out of range error occurs while loading
	// from file that contains reference to file in the parent directory.
	require.NotPanics(t, func() {
		_, err := loader.LoadFromFile("testdata/issue652/nested/schema.yml")
		require.NoError(t, err)
	})
}
