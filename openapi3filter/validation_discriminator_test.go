package openapi3filter

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
	"github.com/stretchr/testify/require"
)

func TestValidationWithDiscriminatorSelection(t *testing.T) {
	const spec = `
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
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	router, err := legacyrouter.NewRouter(doc)
	require.NoError(t, err)

	body := bytes.NewReader([]byte(`{"discr": "objA", "base64": "S25vY2sgS25vY2ssIE5lbyAuLi4="}`))
	req, err := http.NewRequest("PUT", "/blob", body)
	require.NoError(t, err)
	req.Header.Add(headerCT, "application/json")

	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	requestValidationInput := &RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}
	err = ValidateRequest(loader.Context, requestValidationInput)
	require.NoError(t, err)
}
