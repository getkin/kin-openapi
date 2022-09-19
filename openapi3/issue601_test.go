package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue601(t *testing.T) {
	sl := NewLoader()
	doc, err := sl.LoadFromFile("testdata/lxkns.yaml")
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)
}
