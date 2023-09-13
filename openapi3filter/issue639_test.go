package openapi3filter

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestIssue639(t *testing.T) {
	loader := openapi3.NewLoader()
	ctx := loader.Context
	spec := `
    openapi: 3.0.0
    info:
     version: 1.0.0
     title: Sample API
    paths:
     /items:
      put:
        requestBody:
         content:
          application/json:
            schema:
              properties:
                testWithdefault:
                  default: false
                  type: boolean
                testNoDefault:
                  type: boolean
              type: object
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
		name               string
		options            *Options
		expectedDefaultVal interface{}
	}{
		{
			name: "no defaults are added to requests",
			options: &Options{
				SkipSettingDefaults: true,
			},
			expectedDefaultVal: nil,
		},

		{
			name:               "defaults are added to requests",
			expectedDefaultVal: false,
		},
	}

	for _, testcase := range tests {
		t.Run(testcase.name, func(t *testing.T) {
			body := "{\"testNoDefault\": true}"
			httpReq, err := http.NewRequest(http.MethodPut, "/items", strings.NewReader(body))
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
			bodyAfterValidation, err := io.ReadAll(httpReq.Body)
			require.NoError(t, err)

			raw := map[string]interface{}{}
			err = json.Unmarshal(bodyAfterValidation, &raw)
			require.NoError(t, err)
			require.Equal(t, testcase.expectedDefaultVal,
				raw["testWithdefault"], "default value must not be included")
		})
	}
}
