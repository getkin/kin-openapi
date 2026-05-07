package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue638(t *testing.T) {
	for range 50 {
		loader := openapi3.NewLoader()
		loader.IsExternalRefsAllowed = true
		// This path affects the occurrence of the issue #638.
		// ../openapi3/testdata/issue638/test1.yaml : reproduce
		// ./testdata/issue638/test1.yaml           : not reproduce
		// testdata/issue638/test1.yaml             : reproduce
		doc, err := loader.LoadFromFile("testdata/issue638/test1.yaml")
		require.NoError(t, err)
		require.Equal(t, &openapi3.Types{"int"}, doc.Components.Schemas["test1d"].Value.Type)
	}
}
