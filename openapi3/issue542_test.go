package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue542(t *testing.T) {
	sl := NewLoader()

	doc, err := sl.LoadFromFile("testdata/issue542.yml")
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)

	require.Empty(t, doc.Components.Schemas["Cat"].Value.AnyOf[1].Value.Properties["offspring"])
}
