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

func TestIssue789(t *testing.T) {
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
        required: true
        schema:
          type: string
          anyOf:
            - pattern: '\babc\b'
            - pattern: '\bfoo\b'
            - pattern: '\bbar\b'
      responses:
        '200':
          description: Successful response
`[1:]

	oneOfArraySpec := strings.ReplaceAll(anyOfArraySpec, "anyOf", "oneOf")

	allOfArraySpec := strings.ReplaceAll(anyOfArraySpec, "anyOf", "allOf")

	tests := []struct {
		name   string
		spec   string
		req    string
		errStr string
	}{
		{
			name: "success anyof string pattern match",
			spec: anyOfArraySpec,
			req:  "/items?test=abc",
		},
		{
			name:   "failed anyof string pattern match",
			spec:   anyOfArraySpec,
			req:    "/items?test=def",
			errStr: `parameter "test" in query has an error: doesn't match any schema from "anyOf"`,
		},
		{
			name: "success allof object array",
			spec: allOfArraySpec,
			req:  `/items?test=abc foo bar`,
		},
		{
			name:   "failed allof object array",
			spec:   allOfArraySpec,
			req:    `/items?test=foo`,
			errStr: `parameter "test" in query has an error: string doesn't match the regular expression`,
		},
		{
			name: "success oneof string pattern match",
			spec: oneOfArraySpec,
			req:  `/items?test=foo`,
		},
		{
			name:   "failed oneof string pattern match",
			spec:   oneOfArraySpec,
			req:    `/items?test=def`,
			errStr: `parameter "test" in query has an error: doesn't match schema due to: string doesn't match the regular expression`,
		},
		{
			name:   "failed oneof string pattern match",
			spec:   oneOfArraySpec,
			req:    `/items?test=foo bar`,
			errStr: `parameter "test" in query has an error: input matches more than one oneOf schemas`,
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
