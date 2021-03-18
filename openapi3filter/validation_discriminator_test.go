package openapi3filter

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

var yaJsonSpecWithDiscriminator = []byte(`
openapi: 3.0.0
info:
  version: 0.2.0
  title: yaAPI

paths:

  /blob:
    put:
      operationId: SetObj
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/blob'
      responses:
        '200':
          description: Ok

components:
  schemas:
    blob:
      oneOf:
        - $ref: '#/components/schemas/objA'
        - $ref: '#/components/schemas/objB'
      discriminator:
        propertyName: discr
        mapping:
          objA: '#/components/schemas/objA'
          objB: '#/components/schemas/objB'
    genericObj:
      type: object
      required:
        - discr
      properties:
        discr:
          type: string
          enum:
            - objA
            - objB
      discriminator:
        propertyName: discr
        mapping:
          objA: '#/components/schemas/objA'
          objB: '#/components/schemas/objB'
    objA:
      allOf:
      - $ref: '#/components/schemas/genericObj'
      - type: object
        properties:
          base64:
            type: string

    objB:
      allOf:
      - $ref: '#/components/schemas/genericObj'
      - type: object
        properties:
          value:
            type: integer
`)

func forgeRequest(body string) *http.Request {
	iobody := bytes.NewReader([]byte(body))
	req, _ := http.NewRequest("PUT", "/blob", iobody)
	req.Header.Add(headerCT, "application/json")
	return req
}

func TestValidationWithDiscriminatorSelection(t *testing.T) {
	openapi, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(yaJsonSpecWithDiscriminator)
	require.NoError(t, err)
	router := NewRouter().WithSwagger(openapi)
	req := forgeRequest(`{"discr": "objA", "base64": "S25vY2sgS25vY2ssIE5lbyAuLi4="}`)
	route, pathParams, _ := router.FindRoute(req.Method, req.URL)
	requestValidationInput := &RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}
	ctx := context.Background()
	err = ValidateRequest(ctx, requestValidationInput)
	require.NoError(t, err)
}
