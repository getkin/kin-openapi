package openapi3filter

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
)

func TestReadOnlyWriteOnlyPropertiesValidation(t *testing.T) {
	type testCase struct {
		name                string
		requestSchema       string
		responseSchema      string
		requestBody         string
		responseBody        string
		responseErrContains string
		requestErrContains  string
	}

	testCases := []testCase{
		{
			name: "valid_readonly_in_response_and_valid_writeonly_in_request",
			requestSchema: `
			"schema":{
				"type": "object",
				"required": ["_id"],
				"properties": {
					"_id": {
						"type": "string",
						"writeOnly": true
					}
				}
			}`,
			responseSchema: `
			"schema":{
				"type": "object",
				"required": ["access_token"],
				"properties": {
					"access_token": {
						"type": "string",
						"readOnly": true
					}
				}
			}`,
			requestBody:  `{"_id": "bt6kdc3d0cvp6u8u3ft0"}`,
			responseBody: `{"access_token": "abcd"}`,
		},
		{
			name: "valid_readonly_in_response_and_invalid_readonly_in_request",
			requestSchema: `
			"schema":{
				"type": "object",
				"required": ["_id"],
				"properties": {
					"_id": {
						"type": "string",
						"readOnly": true
					}
				}
			}`,
			responseSchema: `
			"schema":{
				"type": "object",
				"required": ["access_token"],
				"properties": {
					"access_token": {
						"type": "string",
						"readOnly": true
					}
				}
			}`,
			requestBody:        `{"_id": "bt6kdc3d0cvp6u8u3ft0"}`,
			responseBody:       `{"access_token": "abcd"}`,
			requestErrContains: `readOnly property "_id" in request`,
		},
		{
			name: "invalid_writeonly_in_response_and_valid_writeonly_in_request",
			requestSchema: `
			"schema":{
				"type": "object",
				"required": ["_id"],
				"properties": {
					"_id": {
						"type": "string",
						"writeOnly": true
					}
				}
			}`,
			responseSchema: `
			"schema":{
				"type": "object",
				"required": ["access_token"],
				"properties": {
					"access_token": {
						"type": "string",
						"writeOnly": true
					}
				}
			}`,
			requestBody:         `{"_id": "bt6kdc3d0cvp6u8u3ft0"}`,
			responseBody:        `{"access_token": "abcd"}`,
			responseErrContains: `writeOnly property "access_token" in response`,
		},
		{
			name: "invalid_writeonly_in_response_and_invalid_readonly_in_request",
			requestSchema: `
			"schema":{
				"type": "object",
				"required": ["_id"],
				"properties": {
					"_id": {
						"type": "string",
						"readOnly": true
					}
				}
			}`,
			responseSchema: `
			"schema":{
				"type": "object",
				"required": ["access_token"],
				"properties": {
					"access_token": {
						"type": "string",
						"writeOnly": true
					}
				}
			}`,
			requestBody:         `{"_id": "bt6kdc3d0cvp6u8u3ft0"}`,
			responseBody:        `{"access_token": "abcd"}`,
			responseErrContains: `writeOnly property "access_token" in response`,
			requestErrContains:  `readOnly property "_id" in request`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			spec := bytes.NewBufferString(`{
				"openapi": "3.0.3",
				"info": {
					"version": "1.0.0",
					"title": "title"
				},
				"paths": {
					"/accounts": {
						"post": {
							"description": "Create a new account",
							"requestBody": {
								"required": true,
								"content": {
									"application/json": {`)
			spec.WriteString(tc.requestSchema)
			spec.WriteString(`}
								}
							},
							"responses": {
								"201": {
									"description": "Successfully created a new account",
									"content": {
										"application/json": {`)
			spec.WriteString(tc.responseSchema)
			spec.WriteString(`}
									}
								},
								"400": {
									"description": "The server could not understand the request due to invalid syntax",
								}
							}
						}
					}
				}
				}`)

			sl := openapi3.NewLoader()
			doc, err := sl.LoadFromData(spec.Bytes())
			require.NoError(t, err)
			err = doc.Validate(sl.Context)
			require.NoError(t, err)
			router, err := legacyrouter.NewRouter(doc)
			require.NoError(t, err)

			httpReq, err := http.NewRequest(http.MethodPost, "/accounts", strings.NewReader(tc.requestBody))
			require.NoError(t, err)
			httpReq.Header.Add(headerCT, "application/json")

			route, pathParams, err := router.FindRoute(httpReq)
			require.NoError(t, err)

			reqValidationInput := &RequestValidationInput{
				Request:    httpReq,
				PathParams: pathParams,
				Route:      route,
			}

			if tc.requestSchema != "" {
				err = ValidateRequest(sl.Context, reqValidationInput)

				if tc.requestErrContains != "" {
					require.Error(t, err)
					require.ErrorContains(t, err, tc.requestErrContains)
				} else {
					require.NoError(t, err)
				}
			}

			if tc.responseSchema != "" {
				err = ValidateResponse(sl.Context, &ResponseValidationInput{
					RequestValidationInput: reqValidationInput,
					Status:                 201,
					Header:                 httpReq.Header,
					Body:                   io.NopCloser(strings.NewReader(tc.responseBody)),
				})

				if tc.responseErrContains != "" {
					require.Error(t, err)
					require.ErrorContains(t, err, tc.responseErrContains)
				} else {
					require.NoError(t, err)
				}
			}
		})
	}
}
