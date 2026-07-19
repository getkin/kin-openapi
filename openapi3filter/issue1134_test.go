package openapi3filter_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// Empty string query params used to become nil in parsePrimitive, which skipped
// schema checks. After #1096 they stay "", so format:date was applied even when
// allowEmptyValue is true. See #1134.
func TestIssue1134(t *testing.T) {
	spec := `
openapi: 3.0.3
info:
  title: issue 1134
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
        - name: updatedAt
          in: query
          required: false
          allowEmptyValue: true
          schema:
            type: string
            format: date-time
        - name: strictDate
          in: query
          required: false
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
			name: "empty date with allowEmptyValue",
			url:  "/people?dateOfBirth=",
		},
		{
			name: "empty date-time with allowEmptyValue",
			url:  "/people?updatedAt=",
		},
		{
			name: "valid date still accepted",
			url:  "/people?dateOfBirth=1970-01-01",
		},
		{
			name:       "invalid non-empty date still rejected",
			url:        "/people?dateOfBirth=not-a-date",
			shouldFail: true,
		},
		{
			name:       "empty date without allowEmptyValue still rejected",
			url:        "/people?strictDate=",
			shouldFail: true,
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
