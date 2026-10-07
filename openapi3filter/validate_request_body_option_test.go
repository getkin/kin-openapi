package openapi3filter

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRejectWhenRequestBodyNotSpecified(t *testing.T) {
	const spec = `
openapi: 3.0.3
info:
  title: Request body validation
  version: 1.0.0
paths:
  /documented:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name]
              properties:
                name:
                  type: string
      responses:
        '200':
          description: OK
  /undocumented:
    post:
      responses:
        '200':
          description: OK
`
	router := setupTestRouter(t, spec)
	tests := []struct {
		name    string
		path    string
		body    string
		options Options
		wantErr string
	}{
		{
			name:    "documented body is accepted",
			path:    "/documented",
			body:    `{"name":"widget"}`,
			options: Options{RejectWhenRequestBodyNotSpecified: true},
		},
		{
			name:    "documented body still undergoes schema validation",
			path:    "/documented",
			body:    `{"name":42}`,
			options: Options{RejectWhenRequestBodyNotSpecified: true},
			wantErr: "doesn't match schema",
		},
		{
			name:    "required documented body cannot be omitted",
			path:    "/documented",
			options: Options{RejectWhenRequestBodyNotSpecified: true},
			wantErr: "value is required but missing",
		},
		{
			name:    "undocumented body is rejected",
			path:    "/undocumented",
			body:    `{"name":"widget"}`,
			options: Options{RejectWhenRequestBodyNotSpecified: true},
			wantErr: "request body not allowed for this request",
		},
		{
			name:    "no body is accepted for an operation without a body definition",
			path:    "/undocumented",
			options: Options{RejectWhenRequestBodyNotSpecified: true},
		},
		{
			name: "undocumented body is accepted when rejection is disabled",
			path: "/undocumented",
			body: `{"name":"widget"}`,
		},
		{
			name:    "excluded request body is not rejected",
			path:    "/undocumented",
			body:    `{"name":"widget"}`,
			options: Options{RejectWhenRequestBodyNotSpecified: true, ExcludeRequestBody: true},
		},
	}
	for _, multiError := range []bool{false, true} {
		for _, tt := range tests {
			name := tt.name
			if multiError {
				name += " with MultiError"
			}
			t.Run(name, func(t *testing.T) {
				request, err := http.NewRequest(http.MethodPost, "http://example.com"+tt.path, strings.NewReader(tt.body))
				require.NoError(t, err)
				request.Header.Set("Content-Type", "application/json")
				route, pathParams, err := router.FindRoute(request)
				require.NoError(t, err)
				options := tt.options
				options.MultiError = multiError
				err = ValidateRequest(t.Context(), &RequestValidationInput{
					Request:    request,
					Route:      route,
					PathParams: pathParams,
					Options:    &options,
				})
				if tt.wantErr == "" {
					require.NoError(t, err)
				} else {
					require.ErrorContains(t, err, tt.wantErr)
				}
			})
		}
	}
}
