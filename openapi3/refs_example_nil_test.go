package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestExampleRefNil(t *testing.T) {
	spec := []byte(`
openapi: 3.0.0
info:
  version: 1.0.0
  title: Swagger Petstore
paths:
  /health:
    get:
      responses:
        '200':
          description: OK
          content:
            text/plain:
              schema: { type: string }
              examples:
                alive:
`[1:])

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.EqualError(t, err, `invalid example: value MUST be an object`)
	require.Nil(t, doc)
}
