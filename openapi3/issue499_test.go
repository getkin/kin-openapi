package openapi3_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestIssue499(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	_, err := loader.LoadFromFile("testdata/issue499/main.yml")
	require.NoError(t, err)
}
