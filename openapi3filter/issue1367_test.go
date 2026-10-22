package openapi3filter

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue1367TypelessConstAndEnumPrimitives(t *testing.T) {
	t.Parallel()

	const spec = `
openapi: 3.1.0
info:
  title: repro
  version: 1.0.0
paths:
  /items:
    get:
      parameters:
        - name: kind
          in: query
          required: true
          schema:
            const: widget
        - name: withenum
          in: query
          schema:
            enum: [widget]
      responses:
        '200':
          description: ok
`

	testcases := []struct {
		url  string
		fail bool
	}{{
		url: "?kind=widget",
	}, {
		url: "?kind=widget&withenum=widget",
	}, {
		url:  "?kind=widget&withenum=gadget",
		fail: true,
	}, {
		url:  "?kind=gadget",
		fail: true,
	}, {
		url:  "?kind=widget&withenum=",
		fail: true,
	}}
	for _, tc := range testcases {
		t.Run(tc.url, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/items"+tc.url, nil)
			require.NoError(t, err)

			router := setupTestRouter(t, spec)
			require.NoError(t, err)
			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			err = ValidateRequest(t.Context(), &RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			})
			if tc.fail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIssue1367TypelessConstAndEnumNumberAndBoolean(t *testing.T) {
	t.Parallel()

	const spec = `
openapi: 3.1.0
info:
  title: repro
  version: 1.0.0
paths:
  /picked:
    get:
      parameters:
        - name: rank
          in: query
          required: true
          schema:
            enum: [1, 2, 3]
        - name: flag
          in: query
          schema:
            const: true
      responses:
        '200':
          description: ok
`

	testcases := []struct {
		url  string
		fail bool
	}{{
		url: "?rank=2&flag=true",
	}, {
		url:  "?rank=2&flag=false",
		fail: true,
	}, {
		url:  "?rank=9&flag=true",
		fail: true,
	}}
	for _, tc := range testcases {
		t.Run(tc.url, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/picked"+tc.url, nil)
			require.NoError(t, err)

			router := setupTestRouter(t, spec)
			require.NoError(t, err)
			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			err = ValidateRequest(t.Context(), &RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			})
			if tc.fail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
