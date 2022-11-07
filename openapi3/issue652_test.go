package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue652(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	require.NotPanics(t, func() {
		loader.LoadFromFile("testdata/issue652/nested/schema.yml")
	})
}
