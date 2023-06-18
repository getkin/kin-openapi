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

func TestIssue641(t *testing.T) {

	anyOfSpec := `
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
        anyOf:
        - pattern: "^[0-9]{1,4}$"
        - pattern: "^[0-9]{1,4}$"
        type: string
    responses:
      '200':
        description: Successful response
`[1:]

	allOfSpec := strings.ReplaceAll(anyOfSpec, "anyOf", "allOf")

	tests := []struct {
		name   string
		spec   string
		req    string
		errStr string
	}{

		{
			name: "success anyof pattern",
			spec: anyOfSpec,
			req:  "/items?test=51",
		},
		{
			name:   "failed anyof pattern",
			spec:   anyOfSpec,
			req:    "/items?test=999999",
			errStr: `parameter "test" in query has an error: doesn't match any schema from "anyOf"`,
		},

		{
			name: "success allof pattern",
			spec: allOfSpec,
			req:  `/items?test=51`,
		},
		{
			name:   "failed allof pattern",
			spec:   allOfSpec,
			req:    `/items?test=999999`,
			errStr: `parameter "test" in query has an error: string doesn't match the regular expression "^[0-9]{1,4}$"`,
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
