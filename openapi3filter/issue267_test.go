package openapi3filter

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestIssue267(t *testing.T) {
	spec := `
openapi: 3.0.0
info:
  description: This is a sample of the API
  version: 1.0.0
  title: sample API
tags:
  - name: authorization
    description: Create and validate authorization tokens using oauth
paths:
  /oauth2/token:
    post:
      tags:
      - authorization
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/AccessTokenRequest'
            examples:
              ClientCredentialsTokenRequest:
                $ref: '#/components/examples/ClientCredentialsTokenRequest'
              RefreshTokenRequest:
                $ref: '#/components/examples/RefreshTokenRequest'
          application/x-www-form-urlencoded:
            schema:
              $ref: '#/components/schemas/AccessTokenRequest'
            examples:
              ClientCredentialsTokenRequest:
                $ref: '#/components/examples/ClientCredentialsTokenRequest'
              RefreshTokenRequest:
                $ref: '#/components/examples/RefreshTokenRequest'
      responses:
        '200':
          description: 'The request was successful and a token was issued.'

components:
  examples:
    ClientCredentialsTokenRequest:
      value:
        grant_type: client_credentials
        scope: 'member:read member:write'
    RefreshTokenRequest:
      value:
        grant_type: refresh_token
        client_id: '3fa85f64-5717-4562-b3fc-2c963f66afa6'
        refresh_token: '2fbd6ad96acc4fa99ef36a3e803b010b'
  schemas:
    AccessTokenRequest:
      description: 'Describes all of the potential access token requests that can be received'
      type: object
      oneOf:
      - $ref: '#/components/schemas/ClientCredentialsTokenRequest'
      - $ref: '#/components/schemas/RefreshTokenRequest'
    ClientCredentialsTokenRequest:
      description: 'The client_id and client_secret properties should only be sent in form data if the client does not support basic authentication for sending client credentials.'
      properties:
        grant_type:
          type: string
          enum:
          - client_credentials
          example: 'client_credentials'
        scope:
          description: 'A space separated list of scopes requested for the token'
          type: string
          example: 'member:read member:write'
        client_id:
          description: 'The ID provided when the client application was registered'
          type: string
          example: '3fa85f64-5717-4562-b3fc-2c963f66afa6'
        client_secret:
          description: 'A secret code that would be setup for the client to exchange for an access token.'
          type: string
          example: 'fac663c0-e8b5-4c02-9ad3-ddbd1bbc6964'
      required:
      - grant_type
      - scope
    RefreshTokenRequest:
      type: object
      properties:
        grant_type:
          type: string
          enum:
          - refresh_token
          example: 'refresh_token'
        client_id:
          description: 'The ID provided when the client application was registered'
          type: string
          example: '3fa85f64-5717-4562-b3fc-2c963f66afa6'
        refresh_token:
          description: 'A long lived one time use token that is issued only in cases where the client can be offline or restarted and where the authorization should persist.'
          type: string
          minLength: 32
          example: '2fbd6ad96acc4fa99ef36a3e803b010b'
      required:
      - grant_type
      - client_id
      - refresh_token
`[1:]

	loader := openapi3.NewLoader()

	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	for _, testcase := range []struct {
		ct, data string
	}{
		{
			ct:   "application/json",
			data: `{"grant_type":"client_credentials", "scope":"testscope", "client_id":"myclient", "client_secret":"mypass"}`,
		},
		{
			ct:   "application/x-www-form-urlencoded",
			data: "grant_type=client_credentials&scope=testscope&client_id=myclient&client_secret=mypass",
		},
	} {
		t.Run(testcase.ct, func(t *testing.T) {
			data := strings.NewReader(testcase.data)
			req, err := http.NewRequest("POST", "/oauth2/token", data)
			require.NoError(t, err)
			req.Header.Add("Content-Type", testcase.ct)

			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			validationInput := &RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			}
			err = ValidateRequest(loader.Context, validationInput)
			require.NoError(t, err)
		})
	}
}
