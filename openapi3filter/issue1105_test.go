package openapi3filter_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestIssue1105(t *testing.T) {
	testSchema := `
openapi: 3.0.0
info:
  title: ''
  version: 0.0.1
paths:
  /some/path/{token}:
    parameters:
      - $ref: '#/components/parameters/Token'
    post:
      responses:
components:
  parameters:
    Token:
      name: token
      in: path
      required: true
      schema:
        type: string
        pattern: ^[a-zA-Z0-9\-_=]+$
`[1:]

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(testSchema))
	require.NoError(t, err)
	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	try := func(token string) error {
		t.Helper()
		req, err := http.NewRequest(http.MethodPost, "/some/path/"+token, nil)
		require.NoError(t, err)
		t.Log("path", req.URL.Path, req.URL.RawPath)
		route, pathParams, err := router.FindRoute(req)
		require.NoError(t, err)
		t.Log("route", route.Path)
		t.Log("pathParameters", pathParams)

		requestValidationInput := &openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
		}

		return openapi3filter.ValidateRequest(t.Context(), requestValidationInput)
	}

	err = try("asdf%3D")
	require.Error(t, err)

	err = try("asdf=")
	require.NoError(t, err)
}
