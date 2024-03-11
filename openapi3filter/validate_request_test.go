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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
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
                category:
                  type: string
                  default: Sweets
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
		Category    string `json:"category,omitempty"`
	}
	type args struct {
		requestBody *testRequestBody
		url         string
		apiKey      string
	}
	tests := []struct {
		name                 string
		args                 args
		expectedModification bool
		expectedErr          error
	}{
		{
			name: "Valid request with all fields set",
			args: args{
				requestBody: &testRequestBody{SubCategory: "Chocolate", Category: "Food"},
				url:         "/category?category=cookies",
				apiKey:      "SomeKey",
			},
			expectedModification: false,
			expectedErr:          nil,
		},
		{
			name: "Valid request without certain fields",
			args: args{
				requestBody: &testRequestBody{SubCategory: "Chocolate"},
				url:         "/category?category=cookies",
				apiKey:      "SomeKey",
			},
			expectedModification: true,
			expectedErr:          nil,
		},
		{
			name: "Invalid operation params",
			args: args{
				requestBody: &testRequestBody{SubCategory: "Chocolate"},
				url:         "/category?invalidCategory=badCookie",
				apiKey:      "SomeKey",
			},
			expectedModification: false,
			expectedErr:          &RequestError{},
		},
		{
			name: "Invalid request body",
			args: args{
				requestBody: nil,
				url:         "/category?category=cookies",
				apiKey:      "SomeKey",
			},
			expectedModification: false,
			expectedErr:          &RequestError{},
		},
		{
			name: "Invalid security",
			args: args{
				requestBody: &testRequestBody{SubCategory: "Chocolate"},
				url:         "/category?category=cookies",
				apiKey:      "",
			},
			expectedModification: false,
			expectedErr:          &SecurityRequirementsError{},
		},
		{
			name: "Invalid request body and security",
			args: args{
				requestBody: nil,
				url:         "/category?category=cookies",
				apiKey:      "",
			},
			expectedModification: false,
			expectedErr:          &SecurityRequirementsError{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var requestBody io.Reader
			var originalBodySize int
			if tc.args.requestBody != nil {
				testingBody, err := json.Marshal(tc.args.requestBody)
				require.NoError(t, err)
				requestBody = bytes.NewReader(testingBody)
				originalBodySize = len(testingBody)
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
			if tc.expectedErr != nil {
				return
			}
			body, err := io.ReadAll(validationInput.Request.Body)
			contentLen := int(validationInput.Request.ContentLength)
			bodySize := len(body)
			assert.NoError(t, err, "unable to read request body: %v", err)
			assert.Equal(t, contentLen, bodySize, "expect ContentLength %d to equal body size %d", contentLen, bodySize)
			bodyModified := originalBodySize != bodySize
			assert.Equal(t, bodyModified, tc.expectedModification, "expect request body modification happened: %t, expected %t", bodyModified, tc.expectedModification)

			validationInput.Request.Body, err = validationInput.Request.GetBody()
			assert.NoError(t, err, "unable to re-generate body by GetBody(): %v", err)
			body2, err := io.ReadAll(validationInput.Request.Body)
			assert.NoError(t, err, "unable to read request body: %v", err)
			assert.Equal(t, body, body2, "body by GetBody() is not matched")
		})
	}
}

func TestValidateQueryParams(t *testing.T) {
	type testCase struct {
		name  string
		param *openapi3.Parameter
		query string
		want  map[string]interface{}
		err   error // ParseError or openapi3.SchemaError
	}

	testCases := []testCase{
		// TODO: move tests and any logic regarding schema validation in req_resp_decoder to use ValidateParameter.
		// those should just decode
		{
			name: "deepObject explode additionalProperties with object properties - missing required property",
			param: &openapi3.Parameter{
				Name: "param", In: "query", Style: "deepObject", Explode: explode,
				Schema: objectOf(
					"obj", additionalPropertiesObjectOf(func() *openapi3.SchemaRef {
						sc := openapi3.SchemaRef{}
						s := objectOf(
							"item1", integerSchema,
							"requiredProp", stringSchema,
						)
						sc = *s
						sc.Value.Required = []string{"requiredProp"}

						return &sc
					}()),
					"objIgnored", objectOf("items", stringArraySchema),
				),
			},
			query: "param[obj][prop1][item1]=1",
			err:   &openapi3.SchemaError{SchemaField: "required", Reason: "property \"requiredProp\" is missing"},
		},
		{
			// XXX should this error out?
			name: "deepObject explode additionalProperties with object properties - extraneous nested param property ignored",
			param: &openapi3.Parameter{
				Name: "param", In: "query", Style: "deepObject", Explode: explode,
				Schema: objectOf(
					"obj", additionalPropertiesObjectOf(objectOf(
						"item1", integerSchema,
						"requiredProp", stringSchema,
					)),
					"objIgnored", objectOf("items", stringArraySchema),
				),
			},
			query: "param[obj][prop1][inexistent]=1",
			want: map[string]interface{}{
				"obj": map[string]interface{}{
					"prop1": map[string]interface{}{},
				},
			},
		},
		{
			name: "deepObject explode additionalProperties with object properties",
			param: &openapi3.Parameter{
				Name: "param", In: "query", Style: "deepObject", Explode: explode,
				Schema: objectOf(
					"obj", additionalPropertiesObjectOf(objectOf(
						"item1", integerSchema,
						"requiredProp", stringSchema,
					)),
					"objIgnored", objectOf("items", stringArraySchema),
				),
			},
			query: "param[obj][prop1][item1]=1",
			want: map[string]interface{}{
				"obj": map[string]interface{}{
					"prop1": map[string]interface{}{
						"item1": 1,
					},
				},
			},
		},
		{
			name: "deepObject explode nested objects - misplaced parameter",
			param: &openapi3.Parameter{
				Name: "param", In: "query", Style: "deepObject", Explode: explode,
				Schema: objectOf(
					"obj", objectOf("nestedObjOne", objectOf("items", stringArraySchema)),
				),
			},
			query: "param[obj][nestedObjOne]=baz",
			want: map[string]interface{}{
				"obj": map[string]interface{}{},
			},
			// err: &openapi3.SchemaError{
			// 	// FIXME: failing schema should be at field items with schema: objectOf("items", stringArraySchema)
			// 	SchemaField: "type", Reason: "value must be an array", Value: "baz", Schema: stringArraySchema.Value,
			// },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info := &openapi3.Info{
				Title:   "MyAPI",
				Version: "0.1",
			}
			doc := &openapi3.T{OpenAPI: "3.0.0", Info: info, Paths: openapi3.NewPaths()}
			op := &openapi3.Operation{
				OperationID: "test",
				Parameters:  []*openapi3.ParameterRef{{Value: tc.param}},
				Responses:   openapi3.NewResponses(),
			}
			doc.AddOperation("/test", http.MethodGet, op)
			err := doc.Validate(context.Background())
			require.NoError(t, err)
			router, err := legacyrouter.NewRouter(doc)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodGet, "http://test.org/test?"+tc.query, nil)
			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			input := &RequestValidationInput{Request: req, PathParams: pathParams, Route: route}
			err = ValidateParameter(context.Background(), input, tc.param)

			if tc.err != nil {
				require.Error(t, err)
				e, ok := err.(*RequestError)
				if !ok {
					t.Errorf("error is not a RequestError")

					return
				}
				switch err := e.Unwrap().(type) {
				case *openapi3.SchemaError:
					matchSchemaError(t, err, tc.err)
				case *ParseError:
					matchParseError(t, err, tc.err)
				default:
					t.Errorf("unknown RequestError wrapped error type")
				}

				return
			}

			require.NoError(t, err)

			got, _, err := decodeStyledParameter(tc.param, input)
			require.EqualValues(t, tc.want, got)
		})
	}
}

func matchSchemaError(t *testing.T, got, want error) {
	t.Helper()

	wErr, ok := want.(*openapi3.SchemaError)
	if !ok {
		t.Errorf("want error is not a SchemaError")
		return
	}
	gErr, ok := got.(*openapi3.SchemaError)
	if !ok {
		t.Errorf("got error is not a SchemaError")
		return
	}
	assert.Equalf(t, wErr.SchemaField, gErr.SchemaField, "SchemaError SchemaField differs")
	assert.Equalf(t, wErr.Reason, gErr.Reason, "SchemaError Reason differs")

	if wErr.Schema != nil {
		assert.EqualValuesf(t, wErr.Schema, gErr.Schema, "SchemaError Schema differs")
	}
	if wErr.Value != nil {
		assert.EqualValuesf(t, wErr.Value, gErr.Value, "SchemaError Value differs")
	}
	if wErr.Origin != nil {
		switch wErrOrigin := wErr.Origin.(type) {
		case *openapi3.SchemaError:
			matchSchemaError(t, gErr, wErrOrigin)
		case *ParseError:
			matchParseError(t, gErr, wErrOrigin)
		default:
			t.Errorf("unknown origin error")
		}
	}
}

func TestValidateRequestExcludeQueryParams(t *testing.T) {
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
            type: integer
          required: true
      responses:
        '200':
          description: Ok
`
	req, err := http.NewRequest(http.MethodPost, "/category?category=foo", nil)
	require.NoError(t, err)
	router := setupTestRouter(t, spec)
	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	err = ValidateRequest(context.Background(), &RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
		Options: &Options{
			ExcludeRequestQueryParams: true,
		},
	})
	require.NoError(t, err)

	err = ValidateRequest(context.Background(), &RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
		Options: &Options{
			ExcludeRequestQueryParams: false,
		},
	})
	require.Error(t, err)
}
