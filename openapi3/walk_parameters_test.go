package openapi3_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// walkParametersSpec exercises parameters in components (shared via $ref from
// two operations: dedup), path items, operations, a callback, and a webhook.
var walkParametersSpec = []byte(`
openapi: 3.1.0
info: {title: t, version: "1"}
paths:
  /pets:
    parameters:
      - name: tenant
        in: header
        schema: {type: string}
    get:
      parameters:
        - $ref: '#/components/parameters/Limit'
        - name: filter
          in: query
          schema: {type: string}
      responses:
        "200": {description: ok}
    post:
      parameters:
        - $ref: '#/components/parameters/Limit'
      callbacks:
        onEvent:
          '{$request.body#/url}':
            post:
              parameters:
                - name: attempt
                  in: query
                  schema: {type: integer}
              responses:
                "200": {description: ok}
      responses:
        "201": {description: created}
webhooks:
  newPet:
    post:
      parameters:
        - name: signature
          in: header
          schema: {type: string}
      responses:
        "200": {description: ok}
components:
  parameters:
    Limit:
      name: limit
      in: query
      schema: {type: integer}
`)

func TestWalkParameters(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(walkParametersSpec)
	require.NoError(t, err)

	visited := map[string]string{} // pointer -> parameter name
	require.NoError(t, doc.WalkParameters(func(jsonPointer string, param *openapi3.ParameterRef) error {
		require.NotNil(t, param)
		require.NotNil(t, param.Value)
		_, dup := visited[jsonPointer]
		require.Falsef(t, dup, "pointer %q visited twice", jsonPointer)
		visited[jsonPointer] = param.Value.Name
		return nil
	}))

	require.Equal(t, map[string]string{
		// the shared Limit parameter is visited once, at its definition
		"/components/parameters/Limit":   "limit",
		"/paths/~1pets/parameters/0":     "tenant",
		"/paths/~1pets/get/parameters/1": "filter",
		"/paths/~1pets/post/callbacks/onEvent/{$request.body#~1url}/post/parameters/0": "attempt",
		"/webhooks/newPet/post/parameters/0":                                           "signature",
	}, visited)
}

func TestWalkParameters_ErrorAborts(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(walkParametersSpec)
	require.NoError(t, err)

	boom := errors.New("boom")
	calls := 0
	require.ErrorIs(t, doc.WalkParameters(func(string, *openapi3.ParameterRef) error {
		calls++
		return boom
	}), boom)
	require.Equal(t, 1, calls, "the walk stops at the first error")
}

func TestWalkParameters_NilDoc(t *testing.T) {
	var doc *openapi3.T
	require.NoError(t, doc.WalkParameters(func(string, *openapi3.ParameterRef) error {
		t.Fatal("no visits on a nil document")
		return nil
	}))
}
