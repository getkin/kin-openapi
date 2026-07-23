package openapi3filter_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// Repeated scalar query params: an empty first value must not hide a later
// invalid value from schema validation. See #1230.
func TestIssue1230(t *testing.T) {
	spec := `
openapi: 3.0.3
info:
  title: issue 1230
  version: 0.0.1
paths:
  /people:
    get:
      parameters:
        - name: dateOfBirth
          in: query
          required: false
          allowEmptyValue: true
          schema:
            type: string
            format: date
      responses:
        '200':
          description: ok
`[1:]

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(loader.Context))

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	for _, tc := range []struct {
		name       string
		url        string
		shouldFail bool
	}{
		{
			name: "single empty still allowed",
			url:  "/people?dateOfBirth=",
		},
		{
			name:       "empty first then invalid is rejected",
			url:        "/people?dateOfBirth=&dateOfBirth=not-a-date",
			shouldFail: true,
		},
		{
			name:       "invalid first then empty is rejected",
			url:        "/people?dateOfBirth=not-a-date&dateOfBirth=",
			shouldFail: true,
		},
		{
			name: "empty first then valid is accepted",
			url:  "/people?dateOfBirth=&dateOfBirth=1970-01-01",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tc.url, nil)
			require.NoError(t, err)

			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			err = openapi3filter.ValidateRequest(loader.Context, &openapi3filter.RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			})
			if tc.shouldFail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
