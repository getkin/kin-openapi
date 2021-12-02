package legacy_test

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
	"github.com/stretchr/testify/require"
)

func TestIssue444(t *testing.T) {
	loader := openapi3.NewLoader()
	oas, err := loader.LoadFromData([]byte(`
openapi: '3.0.0'
info:
  title: API
  version: 1.0.0
paths:
  '/path':
    post:
      requestBody:
        required: true
        content:
          application/x-yaml:
            schema:
              type: object
      responses:
        '200':
          description: x
          content:
            application/json:
              schema:
                type: string
`))
	require.NoError(t, err)
	router, err := legacyrouter.NewRouter(oas)
	require.NoError(t, err)

	r := httptest.NewRequest("POST", "/path", bytes.NewReader([]byte(`
foo: bar
`)))
	r.Header.Set("Content-Type", "application/x-yaml")

	openapi3.SchemaErrorDetailsDisabled = true
	route, pathParams, err := router.FindRoute(r)
	require.NoError(t, err)
	reqValidationInput := &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
	}
	err = openapi3filter.ValidateRequest(context.Background(), reqValidationInput)
	require.NoError(t, err)
}
