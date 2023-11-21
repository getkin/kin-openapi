package openapi3filter_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestIssue294(t *testing.T) {

	deepObjectSpec := `
openapi: 3.0.0
info:
  version: 1.0.0
  title: Sample API
paths:
  /items:
    get:
      description: List documents in a collection
      parameters:
      - in: query
        name: filter
        required: false
        style: deepObject
        explode: true
        schema:
          type: object
          properties:
            id:
              oneOf:
                - oneOf:
                  - type: string
                  - type: object
                    title: StringFieldEqualsComparison
                    additionalProperties: false
                    properties:
                      eq:
                        type: string
                    required: [eq]
                - oneOf:
                  - type: object
                    title: StringFieldOEQFilter
                    additionalProperties: false
                    properties:
                      oeq:
                        type: string
                    required: [oeq]
      responses:
        '200':
          description: Documents in the collection
`
	tests := []struct {
		name   string
		req    string
		errStr string
	}{
		{
			name: "success",
			req:  "/items?filter[id][eq]=1",
		},
		{
			name: "success",
			req:  "/items?filter[id][oeq]=1,2",
		},
		{
			name: "success",
			req:  "/items?filter[id]=1",
		},
		{
			name: "success",
			req:  "/items?filter[id]=1",
		},
	}
	for _, testcase := range tests {
		t.Run(testcase.name, func(t *testing.T) {
			loader := openapi3.NewLoader()
			ctx := loader.Context

			doc, err := loader.LoadFromData([]byte(deepObjectSpec))
			require.NoError(t, err)

			err = doc.Validate(ctx)
			require.NoError(t, err)

			router, err := gorillamux.NewRouter(doc)
			require.NoError(t, err)
			httpReq, err := http.NewRequest(http.MethodGet, testcase.req, nil)
			require.NoError(t, err)

			route, pathParams, err := router.FindRoute(httpReq)
			require.NoError(t, err)

			requestValidationInput := &openapi3filter.RequestValidationInput{
				Request:    httpReq,
				PathParams: pathParams,
				Route:      route,
			}
			err = openapi3filter.ValidateRequest(ctx, requestValidationInput)
			if testcase.errStr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, testcase.errStr)
			}
		})
	}
}
