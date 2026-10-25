package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestLoaderResolvesComponentLinks(t *testing.T) {
	t.Parallel()

	spec := []byte(`
openapi: 3.0.0
info: {title: t, version: '1'}
paths: {}
components:
  links:
    a:
      $ref: '#/components/links/b'
    b:
      operationId: x
`)
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	links := doc.Components.Links
	require.Equal(t, "#/components/links/b", links["a"].Ref)
	require.Same(t, links["b"].Value, links["a"].Value)
}
