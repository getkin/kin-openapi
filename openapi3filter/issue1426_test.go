package openapi3filter

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue1426EmptyQueryValueWithDefault(t *testing.T) {
	t.Parallel()

	const spec = `
openapi: 3.0.3
info:
  title: repro
  version: 1.0.0
paths:
  /items:
    get:
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            default: 100
        - name: offset
          in: query
          allowEmptyValue: true
          schema:
            type: integer
            default: 0
      responses:
        '200':
          description: ok
`

	testcases := []struct {
		name      string
		url       string
		wantErr   error
		wantQuery string
	}{{
		name:      "missing parameter gets the default",
		url:       "/items",
		wantQuery: "limit=100&offset=0",
	}, {
		name:    "empty parameter is rejected",
		url:     "/items?limit=",
		wantErr: ErrInvalidEmptyValue,
	}, {
		name:      "empty parameter with allowEmptyValue gets the default",
		url:       "/items?offset=",
		wantQuery: "limit=100&offset=0",
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tc.url, nil)
			require.NoError(t, err)

			router := setupTestRouter(t, spec)
			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			err = ValidateRequest(t.Context(), &RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			})
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.wantQuery, req.URL.RawQuery)
		})
	}
}
