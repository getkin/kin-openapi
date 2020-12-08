package openapi3

import (
	"reflect"
	"testing"

	"github.com/go-openapi/jsonpointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssue222(t *testing.T) {
	spec := `
openapi: 3.0.0
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
servers:
  - url: 'http://petstore.swagger.io/v1'
paths:
  /pets:
    get:
      summary: List all pets
      operationId: listPets
      tags:
        - pets
      parameters:
        - name: limit
          in: query
          description: How many items to return at one time (max 100)
          required: false
          schema:
            type: integer
            format: int32
      responses:
        '200': # <--------------- PANIC HERE

    post:
      summary: Create a pet
      operationId: createPets
      tags:
        - pets
      responses:
        '201':
          description: Null response
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
  '/pets/{petId}':
    get:
      summary: Info for a specific pet
      operationId: showPetById
      tags:
        - pets
      parameters:
        - name: petId
          in: path
          required: true
          description: The id of the pet to retrieve
          schema:
            type: string
      responses:
        '200':
          description: Expected response to a valid request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Pet'
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    Pet:
      type: object
      required:
        - id
        - name
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
        tag:
          type: string
    Pets:
      type: array
      items:
        $ref: '#/components/schemas/Pet'
    Error:
      type: object
      required:
        - code
        - message
      properties:
        code:
          type: integer
          format: int32
        message:
          type: string
`

	_, err := NewSwaggerLoader().LoadSwaggerFromData([]byte(spec))
	require.EqualError(t, err, `invalid response: value MUST be a JSON object`)
}

func TestIssue247(t *testing.T) {
	spec := `openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  license:
    name: Apache 2.0
    url: http://www.apache.org/licenses/LICENSE-2.0.html
  version: 1.0.5
servers:
- url: /api/v3
tags:
- name: pet
  description: Everything about your Pets
  externalDocs:
    description: Find out more
    url: http://swagger.io
- name: store
  description: Operations about user
- name: user
  description: Access to Petstore orders
  externalDocs:
    description: Find out more about our store
    url: http://swagger.io
paths:
  /pet:
    put:
      tags:
      - pet
      summary: Update an existing pet
      description: Update an existing pet by Id
      operationId: updatePet
      requestBody:
        description: Update an existent pet in the store
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Pet'
          application/xml:
            schema:
              $ref: '#/components/schemas/Pet'
          application/x-www-form-urlencoded:
            schema:
              $ref: '#/components/schemas/Pet'
        required: true
      responses:
        "200":
          description: Successful operation
          content:
            application/xml:
              schema:
                $ref: '#/components/schemas/Pet'
            application/json:
              schema:
                $ref: '#/components/schemas/Pet'
        "400":
          description: Invalid ID supplied
        "404":
          description: Pet not found
        "405":
          description: Validation exception
      security:
      - petstore_auth:
        - write:pets
        - read:pets
components:
  schemas:
    Pet:
      type: object
      required:
        - id
        - name
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
        tag:
          type: string
    Pets:
      type: array
      items:
        $ref: '#/components/schemas/Pet'
    Error:
      type: object
      required:
        - code
        - message
      properties:
        code:
          type: integer
          format: int32
        message:
          type: string
    OneOfTest:
      type: object
      oneOf:
        - type: string
        - type: integer
          format: int32
  `
	root, err := NewSwaggerLoader().LoadSwaggerFromData([]byte(spec))
	require.NoError(t, err)

	ptr, err := jsonpointer.New("/paths/~1pet/put/responses/200/content")
	require.NoError(t, err)
	v, kind, err := ptr.Get(root)
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(Content{}).Kind(), kind)
	assert.IsType(t, Content{}, v)

	ptr, err = jsonpointer.New("/paths/~1pet/put/responses/200/content/application~1json/schema")
	require.NoError(t, err)
	v, kind, err = ptr.Get(root)
	assert.NoError(t, err)
	assert.Equal(t, reflect.Ptr, kind)
	assert.IsType(t, &Ref{}, v)
	assert.Equal(t, "#/components/schemas/Pet", v.(*Ref).Ref)

	ptr, err = jsonpointer.New("/components/schemas/Pets/items")
	require.NoError(t, err)
	v, kind, err = ptr.Get(root)
	assert.NoError(t, err)
	assert.Equal(t, reflect.Ptr, kind)
	require.IsType(t, &Ref{}, v)
	assert.Equal(t, "#/components/schemas/Pet", v.(*Ref).Ref)

	ptr, err = jsonpointer.New("/components/schemas/Error/properties/code")
	require.NoError(t, err)
	v, kind, err = ptr.Get(root)
	assert.NoError(t, err)
	assert.Equal(t, reflect.Ptr, kind)
	require.IsType(t, &Schema{}, v)
	assert.Equal(t, "integer", v.(*Schema).Type)

	ptr, err = jsonpointer.New("/components/schemas/OneOfTest/oneOf/0")
	require.NoError(t, err)
	v, kind, err = ptr.Get(root)
	assert.NoError(t, err)
	assert.Equal(t, reflect.Ptr, kind)
	require.IsType(t, &Schema{}, v)
	assert.Equal(t, "string", v.(*Schema).Type)

	ptr, err = jsonpointer.New("/components/schemas/OneOfTest/oneOf/1")
	require.NoError(t, err)
	v, kind, err = ptr.Get(root)
	assert.NoError(t, err)
	assert.Equal(t, reflect.Ptr, kind)
	require.IsType(t, &Schema{}, v)
	assert.Equal(t, "integer", v.(*Schema).Type)

	ptr, err = jsonpointer.New("/components/schemas/OneOfTest/oneOf/5")
	require.NoError(t, err)
	_, _, err = ptr.Get(root)
	assert.Error(t, err)

	ptr, err = jsonpointer.New("/components/schemas/OneOfTest/oneOf/-1")
	require.NoError(t, err)
	_, _, err = ptr.Get(root)
	assert.Error(t, err)
}
