package openapi2conv_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

// A schema may be referenced from outside #/components/schemas, which
// OpenAPI 2 cannot express: the converted document must not dangle.
func TestFromV3SchemaRefOutsideComponentsSchemas(t *testing.T) {
	loader := openapi3.NewLoader()
	doc3, err := loader.LoadFromData([]byte(`
openapi: 3.0.3
info: {title: t, version: "1"}
paths:
  /foo:
    post:
      requestBody:
        content:
          application/json:
            schema: {$ref: '#/components/requestBodies/test/content/application~1json/schema'}
      responses: {"200": {description: ok}}
components:
  requestBodies:
    test:
      content:
        application/json:
          schema: {type: string}
`))
	require.NoError(t, err)

	doc2, err := openapi2conv.FromV3(doc3)
	require.NoError(t, err)

	back, err := openapi2conv.ToV3(doc2)
	require.NoError(t, err)
	schema := back.Paths.Value("/foo").Post.RequestBody.Value.Content["application/json"].Schema
	require.True(t, schema.Value.Type.Is("string"))
}
