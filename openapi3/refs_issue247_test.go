package openapi3_test

import (
	"reflect"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-openapi/jsonpointer"
	"github.com/stretchr/testify/require"
)

func TestIssue247(t *testing.T) {
	spec := []byte(`
openapi: 3.0.2
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
`[1:])

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)
	require.NotNil(t, doc)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	var ptr jsonpointer.Pointer
	var v any
	var kind reflect.Kind

	ptr, err = jsonpointer.New("/paths")
	require.NoError(t, err)
	v, kind, err = ptr.Get(doc)
	require.NoError(t, err)
	require.NotNil(t, v)
	require.IsType(t, &openapi3.Paths{}, v)
	require.Equal(t, reflect.TypeOf(&openapi3.Paths{}).Kind(), kind)

	ptr, err = jsonpointer.New("/paths/~1pet")
	require.NoError(t, err)
	v, kind, err = ptr.Get(doc)
	require.NoError(t, err)
	require.NotNil(t, v)
	require.IsType(t, &openapi3.PathItem{}, v)
	require.Equal(t, reflect.TypeOf(&openapi3.PathItem{}).Kind(), kind)

	ptr, err = jsonpointer.New("/paths/~1pet/put")
	require.NoError(t, err)
	v, kind, err = ptr.Get(doc)
	require.NoError(t, err)
	require.NotNil(t, v)
	require.IsType(t, &openapi3.Operation{}, v)
	require.Equal(t, reflect.TypeOf(&openapi3.Operation{}).Kind(), kind)

	ptr, err = jsonpointer.New("/paths/~1pet/put/responses")
	require.NoError(t, err)
	v, kind, err = ptr.Get(doc)
	require.NoError(t, err)
	require.NotNil(t, v)
	require.IsType(t, &openapi3.Responses{}, v)
	require.Equal(t, reflect.TypeOf(&openapi3.Responses{}).Kind(), kind)

	ptr, err = jsonpointer.New("/paths/~1pet/put/responses/200")
	require.NoError(t, err)
	v, kind, err = ptr.Get(doc)
	require.NoError(t, err)
	require.NotNil(t, v)
	require.IsType(t, &openapi3.Response{}, v)
	require.Equal(t, reflect.TypeOf(&openapi3.Response{}).Kind(), kind)

	ptr, err = jsonpointer.New("/paths/~1pet/put/responses/200/content")
	require.NoError(t, err)
	v, kind, err = ptr.Get(doc)
	require.NoError(t, err)
	require.NotNil(t, v)
	require.IsType(t, openapi3.Content{}, v)
	require.Equal(t, reflect.TypeOf(openapi3.Content{}).Kind(), kind)

	ptr, err = jsonpointer.New("/paths/~1pet/put/responses/200/content/application~1json/schema")
	require.NoError(t, err)
	v, kind, err = ptr.Get(doc)
	require.NoError(t, err)
	require.NotNil(t, v)
	require.IsType(t, &openapi3.Ref{}, v)
	require.Equal(t, reflect.Ptr, kind)
	require.Equal(t, "#/components/schemas/Pet", v.(*openapi3.Ref).Ref)

	ptr, err = jsonpointer.New("/components/schemas/Pets/items")
	require.NoError(t, err)
	v, kind, err = ptr.Get(doc)
	require.NoError(t, err)
	require.NotNil(t, v)
	require.IsType(t, &openapi3.Ref{}, v)
	require.Equal(t, reflect.Ptr, kind)
	require.Equal(t, "#/components/schemas/Pet", v.(*openapi3.Ref).Ref)

	ptr, err = jsonpointer.New("/components/schemas/Error/properties/code")
	require.NoError(t, err)
	v, kind, err = ptr.Get(doc)
	require.NoError(t, err)
	require.NotNil(t, v)
	require.IsType(t, &openapi3.Schema{}, v)
	require.Equal(t, reflect.Ptr, kind)
	require.Equal(t, &openapi3.Types{"integer"}, v.(*openapi3.Schema).Type)

	ptr, err = jsonpointer.New("/components/schemas/OneOfTest/oneOf/0")
	require.NoError(t, err)
	v, kind, err = ptr.Get(doc)
	require.NoError(t, err)
	require.NotNil(t, v)
	require.IsType(t, &openapi3.Schema{}, v)
	require.Equal(t, reflect.Ptr, kind)
	require.Equal(t, &openapi3.Types{"string"}, v.(*openapi3.Schema).Type)

	ptr, err = jsonpointer.New("/components/schemas/OneOfTest/oneOf/1")
	require.NoError(t, err)
	v, kind, err = ptr.Get(doc)
	require.NoError(t, err)
	require.NotNil(t, v)
	require.IsType(t, &openapi3.Schema{}, v)
	require.Equal(t, reflect.Ptr, kind)
	require.Equal(t, &openapi3.Types{"integer"}, v.(*openapi3.Schema).Type)

	ptr, err = jsonpointer.New("/components/schemas/OneOfTest/oneOf/5")
	require.NoError(t, err)
	_, _, err = ptr.Get(doc)
	require.Error(t, err)

	ptr, err = jsonpointer.New("/components/schemas/OneOfTest/oneOf/-1")
	require.NoError(t, err)
	_, _, err = ptr.Get(doc)
	require.Error(t, err)
}
