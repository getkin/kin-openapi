package openapi3filter

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestIssue707(t *testing.T) {
	loader := openapi3.NewLoader()
	ctx := loader.Context
	spec := `
openapi: 3.0.0
info:
  version: 1.0.0
  title: Sample API
paths:
  /items:
    get:
      description: Returns a list of stuff
      parameters:
      - description: parameter with a default value
        explode: true
        in: query
        name: param-with-default
        schema:
          default: 124
          type: integer
        required: false
      responses:
        '200':
          description: Successful response
`[1:]

	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	err = doc.Validate(ctx)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	tests := []struct {
		name          string
		options       *Options
		expectedQuery string
	}{
		{
			name: "no defaults are added to requests parameters",
			options: &Options{
				SkipSettingDefaults: true,
			},
			expectedQuery: "",
		},

		{
			name:          "defaults are added to requests",
			expectedQuery: "param-with-default=124",
		},
	}

	for _, testcase := range tests {
		t.Run(testcase.name, func(t *testing.T) {
			httpReq, err := http.NewRequest(http.MethodGet, "/items", strings.NewReader(""))
			require.NoError(t, err)

			route, pathParams, err := router.FindRoute(httpReq)
			require.NoError(t, err)

			requestValidationInput := &RequestValidationInput{
				Request:    httpReq,
				PathParams: pathParams,
				Route:      route,
				Options:    testcase.options,
			}
			err = ValidateRequest(ctx, requestValidationInput)
			require.NoError(t, err)

			require.NoError(t, err)
			require.Equal(t, testcase.expectedQuery,
				httpReq.URL.RawQuery, "default value must not be included")
		})
	}
}
