package openapi3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestLoaderResolvesHeaderContent(t *testing.T) {
	t.Parallel()

	spec := []byte(`
openapi: 3.0.0
info: {title: t, version: "1"}
paths:
  /x:
    get:
      responses:
        '200':
          description: d
          headers:
            h:
              content:
                application/json:
                  schema: {$ref: '#/components/schemas/S'}
components:
  schemas:
    S: {type: string}
`[1:])

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)
	schema := doc.Paths.Value("/x").Get.Responses.Value("200").Value.Headers["h"].Value.Content["application/json"].Schema
	require.NotNil(t, schema.Value)
	require.NoError(t, doc.Validate(context.Background()))
}

func TestLoaderRejectsNullSchemaInHeaderContent(t *testing.T) {
	t.Parallel()

	spec := []byte(`
openapi: 3.0.0
info: {title: t, version: "1"}
paths: {}
components:
  headers:
    h:
      content:
        application/json:
          schema:
            properties:
              p:
`[1:])

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.EqualError(t, err, `invalid schema: value MUST be an object`)
	require.Nil(t, doc)
}
