package openapi3_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestIssue625(t *testing.T) {
	for _, spec := range []string{
		`
openapi: 3.0.0
info:
  version: 1.0.0
  title: Sample API
paths:
 /items:
  get:
    description: Returns a list of stuff
    parameters:
    - description: test object
      explode: false
      in: query
      name: test
      required: false
      schema:
        type: array
        items: {}   ###
    responses:
      '200':
        description: Successful response
`[1:],
		`
openapi: 3.0.0
info:
  version: 1.0.0
  title: Sample API
paths:
 /items:
  get:
    description: Returns a list of stuff
    parameters:
    - description: test object
      explode: false
      in: query
      name: test
      required: false
      schema:
        type: array
        items:
          type: object   ###
          properties:
            name:
              type: string
    responses:
      '200':
        description: Successful response
`[1:],
	} {
		loader := openapi3.NewLoader()
		ctx := loader.Context

		doc, err := loader.LoadFromData([]byte(spec))
		require.NoError(t, err)

		err = doc.Validate(ctx)
		require.NoError(t, err)

		router, err := gorillamux.NewRouter(doc)
		require.NoError(t, err)

		for _, testcase := range []string{
			`/items?test=3`,
			`/items?test={"name": "test1"}`,
		} {
			t.Run(testcase, func(t *testing.T) {
				httpReq, err := http.NewRequest(http.MethodGet, testcase, nil)
				require.NoError(t, err)

				route, pathParams, err := router.FindRoute(httpReq)
				require.NoError(t, err)

				requestValidationInput := &openapi3filter.RequestValidationInput{
					Request:    httpReq,
					PathParams: pathParams,
					Route:      route,
				}
				err = openapi3filter.ValidateRequest(ctx, requestValidationInput)
				require.NoError(t, err)
			})
		}
	}
}
