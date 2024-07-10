package openapi3

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIssue961(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	_, err := loader.LoadFromFile("./testdata/issue961/main.yml")
	require.NoError(t, err)
}
