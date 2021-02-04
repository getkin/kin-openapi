package openapi3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadsRelativeDocWithEmbeddedSchema(t *testing.T) {
	loader := NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true

	swagger, err := loader.LoadSwaggerFromFile("testdata/relativeDocsWithEmbeddedSchema/index.yml")
	require.NoError(t, err)

	vErr := swagger.Validate(context.TODO())
	require.NoError(t, vErr)
}
