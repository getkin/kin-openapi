package openapi3filter

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestIssue689(t *testing.T) {
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
                    testWithReadOnly:
                      readOnly: true
                      type: boolean
                    testNoReadOnly:
                      type: boolean
                  type: object
            responses:
              '200':
               description: OK
          get:
            responses:
              '200':
                description: OK
                content:
                  application/json:
                    schema:
                      properties:
                        testWithWriteOnly:
                          writeOnly: true
                          type: boolean
                        testNoWriteOnly:
                          type: boolean
`[1:]

	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	err = doc.Validate(ctx)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	tests := []struct {
		name     string
		options  *Options
		body     string
		method   string
		checkErr require.ErrorAssertionFunc
	}{
		// read-only
		{
			name:     "non read-only property is added to request when validation enabled",
			body:     `{"testNoReadOnly": true}`,
			method:   http.MethodPut,
			checkErr: require.NoError,
		},
		{
			name:   "non read-only property is added to request when validation disabled",
			body:   `{"testNoReadOnly": true}`,
			method: http.MethodPut,
			options: &Options{
				ExcludeReadOnlyValidations: true,
			},
			checkErr: require.NoError,
		},
		{
			name:     "read-only property is added to requests when validation enabled",
			body:     `{"testWithReadOnly": true}`,
			method:   http.MethodPut,
			checkErr: require.Error,
		},
		{
			name:   "read-only property is added to requests when validation disabled",
			body:   `{"testWithReadOnly": true}`,
			method: http.MethodPut,
			options: &Options{
				ExcludeReadOnlyValidations: true,
			},
			checkErr: require.NoError,
		},
		// write-only
		{
			name:     "non write-only property is added to request when validation enabled",
			body:     `{"testNoWriteOnly": true}`,
			method:   http.MethodGet,
			checkErr: require.NoError,
		},
		{
			name:   "non write-only property is added to request when validation disabled",
			body:   `{"testNoWriteOnly": true}`,
			method: http.MethodGet,
			options: &Options{
				ExcludeWriteOnlyValidations: true,
			},
			checkErr: require.NoError,
		},
		{
			name:     "write-only property is added to requests when validation enabled",
			body:     `{"testWithWriteOnly": true}`,
			method:   http.MethodGet,
			checkErr: require.Error,
		},
		{
			name:   "write-only property is added to requests when validation disabled",
			body:   `{"testWithWriteOnly": true}`,
			method: http.MethodGet,
			options: &Options{
				ExcludeWriteOnlyValidations: true,
			},
			checkErr: require.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpReq, err := http.NewRequest(test.method, "/items", strings.NewReader(test.body))
			require.NoError(t, err)
			httpReq.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)

			route, pathParams, err := router.FindRoute(httpReq)
			require.NoError(t, err)

			requestValidationInput := &RequestValidationInput{
				Request:    httpReq,
				PathParams: pathParams,
				Route:      route,
				Options:    test.options,
			}

			if test.method == http.MethodGet {
				responseValidationInput := &ResponseValidationInput{
					RequestValidationInput: requestValidationInput,
					Status:                 200,
					Header:                 httpReq.Header,
					Body:                   io.NopCloser(strings.NewReader(test.body)),
					Options:                test.options,
				}
				err = ValidateResponse(ctx, responseValidationInput)

			} else {
				err = ValidateRequest(ctx, requestValidationInput)
			}
			test.checkErr(t, err)
		})
	}
}
