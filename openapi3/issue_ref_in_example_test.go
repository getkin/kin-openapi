package openapi3_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestRefinExample(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("testdata/issue_ref_in_example.yml")
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	param := doc.Components.Parameters["SortBy"].Value
	require.NotNil(t, param, "Parameter should not be nil")

	exampleRef := param.Examples["NiceExample"]
	require.NotNil(t, exampleRef, "Example ref should not be nil")

	// Before the fix, this would be nil
	require.NotNil(t, exampleRef.Value, "Example value should not be nil after loading")
	require.Equal(t, "Just a nice example.", exampleRef.Value.Summary)
	require.Equal(t, "fooBar", exampleRef.Value.Value)
}
