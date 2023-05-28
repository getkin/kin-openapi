package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO: update file name, test yml and test name to iclude issue number
func TestIssue(t *testing.T) {
	// Need to set CircularReferenceCounter to > 10
	CircularReferenceCounter = 20
	loader := NewLoader()
	doc, err := loader.LoadFromFile("testdata/issue.yml")
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)
}
