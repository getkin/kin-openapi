package openapi3filter

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// TestValidateRequestCustomMethods checks that request validation is oblivious
// to the HTTP method: the OpenAPI 3.2 `query` operation and custom methods held
// in `additionalOperations` get their request body validated like any other.
func TestValidateRequestCustomMethods(t *testing.T) {
	const spec = `
openapi: 3.2.0
info:
  title: 'Validator'
  version: 0.0.1
paths:
  /search:
    query:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - term
              properties:
                term:
                  type: string
      responses:
        '200':
          description: OK
    additionalOperations:
      COPY:
        requestBody:
          required: true
          content:
            application/json:
              schema:
                type: object
                required:
                  - destination
                properties:
                  destination:
                    type: string
        responses:
          '200':
            description: OK
`

	router := setupTestRouter(t, spec)

	for _, tc := range []struct {
		name    string
		method  string
		body    string
		wantErr bool
	}{
		{"QUERY valid body", openapi3.MethodQuery, `{"term":"cats"}`, false},
		{"QUERY missing required property", openapi3.MethodQuery, `{}`, true},
		{"QUERY wrong property type", openapi3.MethodQuery, `{"term":42}`, true},
		{"COPY valid body", "COPY", `{"destination":"/elsewhere"}`, false},
		{"COPY missing required property", "COPY", `{}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, "/search", strings.NewReader(tc.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)
			require.Equal(t, tc.method, route.Method)

			err = ValidateRequest(t.Context(), &RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			})
			if tc.wantErr {
				require.Error(t, err)
				require.IsType(t, &RequestError{}, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("missing required body", func(t *testing.T) {
		req, err := http.NewRequest(openapi3.MethodQuery, "/search", nil)
		require.NoError(t, err)

		route, pathParams, err := router.FindRoute(req)
		require.NoError(t, err)

		err = ValidateRequest(t.Context(), &RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
		})
		require.Error(t, err)
	})
}
