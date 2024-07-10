package openapi3

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIssue499(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	_, err := loader.LoadFromFile("testdata/issue499/main.yml")
	require.NoError(t, err)
}
