package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue495(t *testing.T) {
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
        $ref: '#'
paths:
  /categories:
    get:
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                properties:
                  allOf:
                    $ref: '#/components/schemas/schemaArray'
`[1:])

		sl := NewLoader()

		doc, err := sl.LoadFromData(spec)
		require.NoError(t, err)

		err = doc.Validate(sl.Context)
		require.EqualError(t, err, `invalid components: schema "schemaArray": found unresolved ref: "#"`)
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
                properties:
                  allOf:
                    $ref: '#/components/schemas/schemaArray'
`[1:])

	sl := NewLoader()

	doc, err := sl.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)

	require.Equal(t, &Schema{Type: "object"}, doc.Components.Schemas["schemaArray"].Value.Items.Value)
}

func TestIssue495WithDraft04(t *testing.T) {
	spec := []byte(`
openapi: 3.0.1
servers:
- url: http://localhost:5000
info:
  version: v1
  title: Products api
  contact:
    name: me
    email: me@github.com
  description: This is a sample
paths:
  /categories:
    get:
      summary: Provides the available categories for the store
      operationId: list-categories
      responses:
        '200':
          description: this is a desc
          content:
            application/json:
              schema:
                $ref: http://json-schema.org/draft-04/schema
`[1:])

	sl := NewLoader()
	sl.IsExternalRefsAllowed = true

	doc, err := sl.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.ErrorContains(t, err, `found unresolved ref: "#"`)
}

func TestIssue495WithDraft04Bis(t *testing.T) {
	spec := []byte(`
openapi: 3.0.1
servers:
- url: http://localhost:5000
info:
  version: v1
  title: Products api
  contact:
    name: me
    email: me@github.com
  description: This is a sample
paths:
  /categories:
    get:
      summary: Provides the available categories for the store
      operationId: list-categories
      responses:
        '200':
          description: this is a desc
          content:
            application/json:
              schema:
                $ref: testdata/draft04.yml
`[1:])

	sl := NewLoader()
	sl.IsExternalRefsAllowed = true

	doc, err := sl.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.ErrorContains(t, err, `found unresolved ref: "#"`)
}
