package openapi3filter

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestIssue624(t *testing.T) {
	ctx := context.Background()
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	spec := `
openapi: 3.0.0
info:
  version: 1.0.0
  title: Sample API
paths:
  /items:
    get:
      description: Returns a list of stuff
      parameters:
      - description: "test non object"
        explode: true
        style: form
        in: query
        name: test
        required: false
        content:
          application/json:
            schema:
              anyOf:
              - type: string
              - type: integer
      responses:
        '200':
          description: Successful response
`[1:]

	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	err = doc.Validate(ctx)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)
	httpReq, err := http.NewRequest(http.MethodGet, `/items?test=test1`, nil)
	require.NoError(t, err)

	route, pathParams, err := router.FindRoute(httpReq)
	require.NoError(t, err)

	requestValidationInput := &RequestValidationInput{
		Request:    httpReq,
		PathParams: pathParams,
		Route:      route,
	}
	err = ValidateRequest(ctx, requestValidationInput)
	require.NoError(t, err)
}
