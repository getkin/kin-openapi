package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue796(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromFile("testdata/issue796.yml")
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)
}
