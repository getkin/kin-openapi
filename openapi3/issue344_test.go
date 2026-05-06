package openapi3_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestIssue344(t *testing.T) {
	sl := openapi3.NewLoader()
	sl.IsExternalRefsAllowed = true

	doc, err := sl.LoadFromFile("testdata/spec.yaml")
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)

	require.Equal(t, &openapi3.Types{"string"}, doc.Components.Schemas["Test"].Value.Properties["test"].Value.Properties["name"].Value.Type)
}
