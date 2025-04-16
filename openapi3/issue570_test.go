package openapi3

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssue570(t *testing.T) {
	loader := NewLoader()
	_, err := loader.LoadFromFile("testdata/issue570.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), CircularReferenceError)
}
