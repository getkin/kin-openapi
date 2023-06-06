package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue796(t *testing.T) {
	var old int
	// Need to set CircularReferenceCounter to > 10
	old, CircularReferenceCounter = CircularReferenceCounter, 20
	defer func() { CircularReferenceCounter = old }()

	loader := NewLoader()
	doc, err := loader.LoadFromFile("testdata/issue796.yml")
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)
}
