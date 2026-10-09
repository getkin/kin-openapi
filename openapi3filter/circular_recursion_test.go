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

// These tests cover a fatal stack overflow triggered by a schema whose
// allOf/anyOf/oneOf chain refers back to itself. Both the spec loader and
// Schema.validate() accept such schemas, but the runtime visitor
// (Schema.visitJSON -> visitXOFOperations) and the request-decoding
// recursion (openapi3filter.decodeValue) lacked cycle detection and
// recursed until the goroutine stack was exhausted, aborting the process
// with an unrecoverable runtime error. Both paths must now terminate
// with a clean error instead.

func TestCircularAllOf_VisitJSONNoCrash(t *testing.T) {
	t.Parallel()

	recursive := &openapi3.Schema{
		Type: &openapi3.Types{"object"},
		Properties: openapi3.Schemas{
			"name": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
		},
	}
	recursive.AllOf = openapi3.SchemaRefs{{Value: recursive}}

	require.NoError(t, recursive.Validate(t.Context()))
	_ = recursive.VisitJSON(map[string]any{"name": "test"})
}

const circularAllOfSpec = `
openapi: 3.0.0
info: {title: t, version: '1.0'}
paths:
  /q:
    get:
      parameters:
        - name: data
          in: query
          schema: {$ref: '#/components/schemas/Recursive'}
      responses: {"200": {description: OK}}
components:
  schemas:
    Recursive:
      type: object
      allOf:
        - $ref: '#/components/schemas/Recursive'
      properties:
        name: {type: string}
`

func TestCircularAllOf_DecodeValueNoCrash(t *testing.T) {
	t.Parallel()

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(circularAllOfSpec))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(loader.Context))
	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "/q?data=x", strings.NewReader(""))
	require.NoError(t, err)
	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	err = openapi3filter.ValidateRequest(t.Context(), &openapi3filter.RequestValidationInput{
		Request: req, PathParams: pathParams, Route: route,
		Options: &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc},
	})
	require.Error(t, err, "must return a clean error, not crash")
}
