package openapi3_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestUnmarshalError(t *testing.T) {
	t.Parallel()

	{
		spec := []byte(`
openapi: 3.0.1
info:
  version: v1
  title: Products api
components:
  schemas:
    someSchema:
      type: object
    schemaArray:
      type: array
      minItems: 1
      items:
        $ref: '#/components/schemas/someSchema'
paths:
  /categories:
    get:
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                allOf:
                  $ref: '#/components/schemas/schemaArray'  # <- Should have been a list
`[1:])

		sl := openapi3.NewLoader()

		_, err := sl.LoadFromData(spec)
		if openapi3.UsingJSONv2 {
			require.ErrorContains(t, err, `json: cannot unmarshal object into Go struct field .paths.~1categories.get.responses.200.content.application~1json.schema.allOf of type openapi3.SchemaRefs`)
		} else {
			require.ErrorContains(t, err, `json: cannot unmarshal object into field Schema.allOf of type openapi3.SchemaRefs`)
		}
	}

	spec := []byte(`
openapi: 3.0.1
info:
  version: v1
  title: Products api
components:
  schemas:
    someSchema:
      type: object
    schemaArray:
      type: array
      minItems: 1
      items:
        $ref: '#/components/schemas/someSchema'
paths:
  /categories:
    get:
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                allOf:
                - $ref: '#/components/schemas/schemaArray'  # <-
`[1:])

	sl := openapi3.NewLoader()

	doc, err := sl.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)
}

func TestUnmarshalErrorWrapsTheDecodersTried(t *testing.T) {
	t.Parallel()

	const yamlSpec = `
openapi: 3.0.3
info: {title: t, version: v}
paths: {}
components:
  schemas:
    A:
      type: 123
`
	const jsonSpec = `
{"openapi": "3.0.3", "info": {"title": "t", "version": "v"}, "paths": {},
 "components": {"schemas": {"A": {"type": 123}}}}
`

	t.Run("yaml is not tried as json", func(t *testing.T) {
		t.Parallel()

		loader := openapi3.NewLoader()
		_, err := loader.LoadFromData([]byte(yamlSpec))
		require.ErrorContains(t, err, "failed to unmarshal data: yaml error: ")
		msg := err.Error()
		require.NotContains(t, msg, "json error")
		require.Error(t, errors.Unwrap(err))
	})

	t.Run("json reports both decoders", func(t *testing.T) {
		t.Parallel()

		loader := openapi3.NewLoader()
		_, err := loader.LoadFromData([]byte(jsonSpec))
		require.ErrorContains(t, err, "failed to unmarshal data: json error: ")
		require.ErrorContains(t, err, ", yaml error: ")
		var wrapped interface{ Unwrap() []error }
		require.ErrorAs(t, err, &wrapped)
		require.Len(t, wrapped.Unwrap(), 2)
	})
}
