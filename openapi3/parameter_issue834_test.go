package openapi3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathItemParametersAreValidated(t *testing.T) {
	spec := []byte(`
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
servers:
  - url: http://petstore.swagger.io/v1
paths:
  /pets:
    parameters:
      - in: invalid
        name: test
        schema:
          type: string
    get:
      summary: List all pets
      operationId: listPets
      tags:
        - pets
      responses:
        '200':
          description: A paged array of pets
`[1:])

	loader := NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)
	err = doc.Validate(context.Background())
	require.EqualError(t, err, `invalid paths: invalid path /pets: parameter can't have 'in' value "invalid"`)
}

func TestParameterMultipleContentEntries(t *testing.T) {
	spec := []byte(`
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
servers:
  - url: http://petstore.swagger.io/v1
paths:
  /pets:
    parameters:
      - in: query
        name: test
        content:
          application/json:
            schema:
              type: string
          application/xml:
            schema:
              type: string
    get:
      summary: List all pets
      operationId: listPets
      tags:
        - pets
      responses:
        '200':
          description: A paged array of pets
`[1:])

	loader := NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)
	err = doc.Validate(context.Background())
	require.EqualError(t, err, `invalid paths: invalid path /pets: parameter "test" content is invalid: parameter content must only contain one entry`)
}

func TestParameterEmptyContent(t *testing.T) {
	spec := []byte(`
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
servers:
  - url: http://petstore.swagger.io/v1
paths:
  /pets:
    parameters:
      - in: query
        name: test
        content: {}
    get:
      summary: List all pets
      operationId: listPets
      tags:
        - pets
      responses:
        '200':
          description: A paged array of pets
`[1:])

	loader := NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)
	err = doc.Validate(context.Background())
	require.EqualError(t, err, `invalid paths: invalid path /pets: parameter "test" schema is invalid: parameter must contain exactly one of content and schema`)
}
