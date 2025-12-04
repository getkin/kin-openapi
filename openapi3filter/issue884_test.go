package openapi3filter

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestIssue884(t *testing.T) {
	loader := openapi3.NewLoader()
	ctx := loader.Context
	spec := `
  openapi: 3.0.0
  info:
    version: 1.0.0
    title: Sample API
  components:
    schemas:
      TaskSortEnum:
        enum:
          - createdAt
          - -createdAt
          - updatedAt
          - -updatedAt
  paths:
    /tasks:
      get:
        operationId: ListTask
        parameters:
          - in: query
            name: withDefault
            schema:
              allOf:
                - $ref: '#/components/schemas/TaskSortEnum'
                - default: -createdAt
          - in: query
            name: withoutDefault
            schema:
              allOf:
                - $ref: '#/components/schemas/TaskSortEnum'
          - in: query
            name: withManyDefaults
            schema:
              allOf:
                - default: -updatedAt
                - $ref: '#/components/schemas/TaskSortEnum'
                - default: -createdAt
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
		expectedQuery url.Values
	}{
		{
			name: "no defaults are added to requests",
			options: &Options{
				SkipSettingDefaults: true,
			},
			expectedQuery: url.Values{},
		},

		{
			name: "defaults are added to requests",
			expectedQuery: url.Values{
				"withDefault":      []string{"-createdAt"},
				"withManyDefaults": []string{"-updatedAt"}, // first default is win
			},
		},
	}

	for _, testcase := range tests {
		t.Run(testcase.name, func(t *testing.T) {
			httpReq, err := http.NewRequest(http.MethodGet, "/tasks", nil)
			require.NoError(t, err)
			httpReq.Header.Set("Content-Type", "application/json")
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

			q := httpReq.URL.Query()
			assert.Equal(t, testcase.expectedQuery, q)
		})
	}
}
