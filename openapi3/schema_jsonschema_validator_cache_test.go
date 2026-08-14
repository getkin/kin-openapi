package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSchemaJSONSchema2020ValidatorCache(t *testing.T) {
	schema := &Schema{Type: &Types{"string"}}

	require.NoError(t, schema.useJSONSchema2020(&schemaValidationSettings{useJSONSchema2020: true}, "hello"))

	// The compiled validator should be cached
	v, ok := compiledJSONSchemaValidators.Load(schema)
	require.True(t, ok)
	require.NotNil(t, v)

	// A second call should reuse the same compiled validator
	require.NoError(t, schema.useJSONSchema2020(&schemaValidationSettings{useJSONSchema2020: true}, "world"))
	v2, ok := compiledJSONSchemaValidators.Load(schema)
	require.True(t, ok)
	require.Same(t, v, v2)
}
