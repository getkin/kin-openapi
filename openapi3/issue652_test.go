package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue652(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// Test checks that no slice bounds out of range error occurs while loading
	// from file that contains reference to file in the parent directory.
	require.NotPanics(t, func() {
		const schemaName = "ReferenceToParentDirectory"

		spec, err := loader.LoadFromFile("testdata/issue652/nested/schema.yml")
		require.NoError(t, err)
		require.Contains(t, spec.Components.Schemas, schemaName)

		schema := spec.Components.Schemas[schemaName]
		assert.Equal(t, schema.Ref, "../definitions.yml#/components/schemas/TestSchema")
		assert.Equal(t, schema.Value.Type, &openapi3.Types{"string"})
	})
}
