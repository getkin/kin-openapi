package openapi3filter

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
)

func TestReadOnlyWriteOnlyPropertiesValidation(t *testing.T) {
	type testCase struct {
		name           string
		requestSchema  string
		responseSchema string
		errContains    string
	}

	testCases := []testCase{
		{
			name: "invalid_readonly_in_request",
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
			errContains: "readOnly property \"_id\" in request",
		},
		{
			name: "valid_writeonly_in_request",
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
		},
		{
			name: "invalid_writeonly_in_response",
			responseSchema: `
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
			errContains: "writeOnly property \"_id\" in response",
		},
		{
			name: "valid_readonly_in_response",
			responseSchema: `
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
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			spec := bytes.Buffer{}
			spec.WriteString(`{
				"openapi": "3.0.3",
				"info": {
					"version": "1.0.0",
					"title": "title",
					"description": "desc",
					"contact": {
						"email": "email"
					}
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

			type Request struct {
				ID string `json:"_id"`
			}

			sl := openapi3.NewLoader()
			doc, err := sl.LoadFromData(spec.Bytes())
			require.NoError(t, err)
			err = doc.Validate(sl.Context)
			require.NoError(t, err)
			router, err := legacyrouter.NewRouter(doc)
			require.NoError(t, err)

			b, err := json.Marshal(Request{ID: "bt6kdc3d0cvp6u8u3ft0"})
			require.NoError(t, err)

			httpReq, err := http.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(b))
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
			} else if tc.responseSchema != "" {
				err = ValidateResponse(sl.Context, &ResponseValidationInput{
					RequestValidationInput: reqValidationInput,
					Status:                 201,
					Header:                 httpReq.Header,
					Body:                   httpReq.Body,
				})
			}

			if tc.errContains != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
