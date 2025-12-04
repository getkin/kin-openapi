package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalError(t *testing.T) {
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

		sl := NewLoader()

		_, err := sl.LoadFromData(spec)
		require.ErrorContains(t, err, `json: cannot unmarshal object into field Schema.allOf of type openapi3.SchemaRefs`)
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

	sl := NewLoader()

	doc, err := sl.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)
}
