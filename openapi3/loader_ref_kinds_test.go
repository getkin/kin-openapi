package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// The same $ref written where two kinds of objects are expected, one within
// the other, must not hand the value resolved for one to the other.
func TestSameRefForDifferentKindsOfObjects(t *testing.T) {
	t.Parallel()
	loader := openapi3.NewLoader()
	require.NotPanics(t, func() {
		_, _ = loader.LoadFromData([]byte(`{
  "openapi": "3.0.3",
  "info": {"title": "t", "version": "1"},
  "paths": {
    "/x": {"$ref": "#/paths/~1y"},
    "/y": {"get": {"responses": {"200": {"$ref": "#/paths/~1y"}}}}
  }
}`))
	})
}
