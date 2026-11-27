package openapi3filter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// The name field of the Parameter Object, in both OpenAPI 3.0 and 3.1: if in is
// "header" and the name field is "Accept", "Content-Type" or "Authorization",
// the parameter definition SHALL be ignored.
func TestReservedHeaderParametersAreIgnored(t *testing.T) {
	t.Parallel()

	const spec = `
openapi: 3.0.3
info:
  title: reserved header parameters
  version: 1.0.0
paths:
  /items:
    get:
      parameters:
        - name: Accept
          in: header
          required: true
          schema:
            type: string
            enum: [application/xml]
        - name: Content-Type
          in: header
          required: true
          schema:
            type: string
        - name: authorization
          in: header
          required: true
          schema:
            type: string
            pattern: '^Bearer '
        - name: X-Tenant
          in: header
          required: true
          schema:
            type: string
      responses:
        '200':
          description: ok
`

	router := setupTestRouter(t, spec)

	testcases := []struct {
		name    string
		headers map[string]string
		wantErr bool
	}{{
		name:    "none of the reserved headers is sent",
		headers: map[string]string{"X-Tenant": "acme"},
	}, {
		name: "the reserved headers are sent with values their schemas reject",
		headers: map[string]string{
			"X-Tenant":      "acme",
			"Accept":        "application/json",
			"Content-Type":  "application/json",
			"Authorization": "Basic dXNlcjpwYXNz",
		},
	}, {
		name:    "a parameter that is not reserved is still enforced",
		headers: map[string]string{},
		wantErr: true,
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/items", nil)
			require.NoError(t, err)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			err = ValidateRequest(t.Context(), &RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			})
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

// An ignored definition must not reach the handler either: the default of a
// Content-Type parameter used to be added to the request before the handler ran.
func TestReservedHeaderParameterDefaultIsNotAdded(t *testing.T) {
	t.Parallel()

	const spec = `
openapi: 3.0.3
info:
  title: reserved header parameter default
  version: 1.0.0
paths:
  /items:
    get:
      parameters:
        - name: Content-Type
          in: header
          schema:
            type: string
            default: application/vnd.unsent
      responses:
        '200':
          description: ok
`

	router := setupTestRouter(t, spec)

	var seen http.Header
	handler := NewValidator(router,
		OnErr(func(_ context.Context, w http.ResponseWriter, status int, _ ErrCode, err error) {
			w.WriteHeader(status)
			if err != nil {
				_, _ = w.Write([]byte(err.Error()))
			}
		}),
	).Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/items", nil))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Empty(t, seen.Get("Content-Type"))
}
