package openapi3filter_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestIssue625(t *testing.T) {

	anyOfArraySpec := `
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
         anyOf:
          - type: integer
          - type: boolean
    responses:
      '200':
        description: Successful response
`[1:]

	oneOfArraySpec := strings.ReplaceAll(anyOfArraySpec, "anyOf", "oneOf")

	allOfArraySpec := strings.ReplaceAll(strings.ReplaceAll(anyOfArraySpec, "anyOf", "allOf"),
		"type: boolean", "type: number")

	tests := []struct {
		name   string
		spec   string
		req    string
		errStr string
	}{
		{
			name: "success anyof object array",
			spec: anyOfArraySpec,
			req:  "/items?test=3,7",
		},
		{
			name:   "failed anyof object array",
			spec:   anyOfArraySpec,
			req:    "/items?test=s1,s2",
			errStr: `parameter "test" in query has an error: path 0: value s1: an invalid boolean: invalid syntax`,
		},

		{
			name: "success allof object array",
			spec: allOfArraySpec,
			req:  `/items?test=1,3`,
		},
		{
			name:   "failed allof object array",
			spec:   allOfArraySpec,
			req:    `/items?test=1.2,3.1`,
			errStr: `parameter "test" in query has an error: path 0: value 1.2: an invalid integer: invalid syntax`,
		},
		{
			name: "success oneof object array",
			spec: oneOfArraySpec,
			req:  `/items?test=true,3`,
		},
		{
			name:   "failed oneof object array",
			spec:   oneOfArraySpec,
			req:    `/items?test="val1","val2"`,
			errStr: `parameter "test" in query has an error: item 0: decoding oneOf failed: 0 schemas matched`,
		},
	}

	for _, testcase := range tests {
		t.Run(testcase.name, func(t *testing.T) {
			loader := openapi3.NewLoader()
			ctx := loader.Context

			doc, err := loader.LoadFromData([]byte(testcase.spec))
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
		},
		)
	}
}
