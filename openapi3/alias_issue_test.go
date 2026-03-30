package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAliasIssue(t *testing.T) {
	// IncludeOrigin = true
	loader := NewLoader()
	_, err := loader.LoadFromFile("testdata/alias.yml")
	require.NoError(t, err)
	// require.Nil(t, doc)
}
