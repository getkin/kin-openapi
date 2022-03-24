package openapi3filter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter(t *testing.T, spec string) routers.Router {
	t.Helper()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	return router
}

func TestValidateRequest(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: 'Validator'
  version: 0.0.1
paths:
  /category:
    post:
      parameters:
        - name: category
          in: query
          schema:
            type: string
          required: true
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - subCategory
              properties:
                subCategory:
                  type: string
      responses:
        '201':
          description: Created
      security:
      - apiKey: []
components:
  securitySchemes:
    apiKey:
      type: apiKey
      name: Api-Key
      in: header
`

	router := setupTestRouter(t, spec)

	verifyAPIKeyPresence := func(c context.Context, input *AuthenticationInput) error {
		if input.SecurityScheme.Type == "apiKey" {
			var found bool
			switch input.SecurityScheme.In {
			case "query":
				_, found = input.RequestValidationInput.GetQueryParams()[input.SecurityScheme.Name]
			case "header":
				_, found = input.RequestValidationInput.Request.Header[http.CanonicalHeaderKey(input.SecurityScheme.Name)]
			case "cookie":
				_, err := input.RequestValidationInput.Request.Cookie(input.SecurityScheme.Name)
				found = !errors.Is(err, http.ErrNoCookie)
			}
			if !found {
				return fmt.Errorf("%v not found in %v", input.SecurityScheme.Name, input.SecurityScheme.In)
			}
		}
		return nil
	}

	type testRequestBody struct {
		SubCategory string `json:"subCategory"`
	}
	type args struct {
		requestBody *testRequestBody
		url         string
		apiKey      string
	}
	tests := []struct {
		name        string
		args        args
		expectedErr error
	}{
		{
			name: "Valid request",
			args: args{
				requestBody: &testRequestBody{SubCategory: "Chocolate"},
				url:         "/category?category=cookies",
				apiKey:      "SomeKey",
			},
			expectedErr: nil,
		},
		{
			name: "Invalid operation params",
			args: args{
				requestBody: &testRequestBody{SubCategory: "Chocolate"},
				url:         "/category?invalidCategory=badCookie",
				apiKey:      "SomeKey",
			},
			expectedErr: &RequestError{},
		},
		{
			name: "Invalid request body",
			args: args{
				requestBody: nil,
				url:         "/category?category=cookies",
				apiKey:      "SomeKey",
			},
			expectedErr: &RequestError{},
		},
		{
			name: "Invalid security",
			args: args{
				requestBody: &testRequestBody{SubCategory: "Chocolate"},
				url:         "/category?category=cookies",
				apiKey:      "",
			},
			expectedErr: &SecurityRequirementsError{},
		},
		{
			name: "Invalid request body and security",
			args: args{
				requestBody: nil,
				url:         "/category?category=cookies",
				apiKey:      "",
			},
			expectedErr: &SecurityRequirementsError{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var requestBody io.Reader
			if tc.args.requestBody != nil {
				testingBody, err := json.Marshal(tc.args.requestBody)
				require.NoError(t, err)
				requestBody = bytes.NewReader(testingBody)
			}
			req, err := http.NewRequest(http.MethodPost, tc.args.url, requestBody)
			require.NoError(t, err)
			req.Header.Add("Content-Type", "application/json")
			if tc.args.apiKey != "" {
				req.Header.Add("Api-Key", tc.args.apiKey)
			}

			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			validationInput := &RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
				Options: &Options{
					AuthenticationFunc: verifyAPIKeyPresence,
				},
			}
			err = ValidateRequest(context.Background(), validationInput)
			assert.IsType(t, tc.expectedErr, err, "ValidateRequest(): error = %v, expectedError %v", err, tc.expectedErr)
		})
	}
}
