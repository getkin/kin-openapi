package openapi3_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestLoaderRefToNullNode(t *testing.T) {
	t.Parallel()

	t.Run("component", func(t *testing.T) {
		t.Parallel()

		spec := []byte(`
openapi: 3.0.0
info: {title: t, version: '1'}
paths: {}
components:
  schemas:
    A:
  responses:
    R:
      description: r
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/A'
`)
		loader := openapi3.NewLoader()
		_, err := loader.LoadFromData(spec)
		require.ErrorContains(t, err, `failed to resolve "A" in fragment in URI: "#/components/schemas/A"`)
	})

	t.Run("field of an external document", func(t *testing.T) {
		t.Parallel()

		spec := []byte(`
openapi: 3.0.0
info: {title: t, version: '1'}
paths:
  /a:
    $ref: 'other.yml#/paths/~1a'
`)
		loader := openapi3.NewLoader()
		loader.IsExternalRefsAllowed = true
		loader.ReadFromURIFunc = func(*openapi3.Loader, *url.URL) ([]byte, error) {
			return []byte(`{"openapi": "3.0.0", "info": {"title": "t", "version": "1"}}`), nil
		}
		_, err := loader.LoadFromData(spec)
		require.ErrorContains(t, err, `failed to resolve "paths" in fragment in URI: "#/paths/~1a"`)
	})
}
