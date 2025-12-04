package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue638(t *testing.T) {
	for i := 0; i < 50; i++ {
		loader := NewLoader()
		loader.IsExternalRefsAllowed = true
		// This path affects the occurrence of the issue #638.
		// ../openapi3/testdata/issue638/test1.yaml : reproduce
		// ./testdata/issue638/test1.yaml           : not reproduce
		// testdata/issue638/test1.yaml             : reproduce
		doc, err := loader.LoadFromFile("testdata/issue638/test1.yaml")
		require.NoError(t, err)
		require.Equal(t, &Types{"int"}, doc.Components.Schemas["test1d"].Value.Type)
	}
}
