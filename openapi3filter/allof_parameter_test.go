package openapi3filter_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// A parameter whose allOf has a subschema that decodes nothing by itself, such
// as {minimum: 3}, used to lose the value decoded by the other subschemas and
// to be refused as an empty value.
func TestAllOfParameterWithTypelessSubschema(t *testing.T) {
	t.Parallel()

	const spec = `
openapi: 3.0.3
info: {title: t, version: '1'}
paths:
  /items/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            allOf: [{type: integer}, {minimum: 3}]
        - name: q
          in: query
          schema:
            allOf: [{type: integer}, {minimum: 3}]
        - name: tags
          in: query
          schema:
            type: array
            items:
              allOf: [{type: integer}, {minimum: 3}]
        - name: X-Level
          in: header
          schema:
            allOf: [{description: a level}, {type: integer, minimum: 3}]
      responses:
        '200': {description: ok}
`
	doc, err := openapi3.NewLoader().LoadFromData([]byte(spec))
	require.NoError(t, err)
	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	for _, tc := range []struct {
		name   string
		url    string
		header string
		errStr string
	}{
		{name: "valid", url: "/items/5?q=4&tags=3&tags=9", header: "7"},
		{name: "path below minimum", url: "/items/1", errStr: `parameter "id" in path has an error: doesn't match schema due to: number must be at least 3`},
		{name: "query below minimum", url: "/items/5?q=1", errStr: `parameter "q" in query has an error: doesn't match schema due to: number must be at least 3`},
		{name: "item below minimum", url: "/items/5?tags=3&tags=1", errStr: `parameter "tags" in query has an error: Error at "/1": doesn't match schema due to: number must be at least 3`},
		{name: "header below minimum", url: "/items/5", header: "1", errStr: `parameter "X-Level" in header has an error: doesn't match schema due to: number must be at least 3`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tc.url, nil)
			require.NoError(t, err)
			if tc.header != "" {
				req.Header.Set("X-Level", tc.header)
			}
			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			err = openapi3filter.ValidateRequest(context.Background(), &openapi3filter.RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			})
			if tc.errStr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.errStr)
		})
	}
}
