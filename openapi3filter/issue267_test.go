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
          application/x-www-form-urlencoded:
            schema:
              $ref: '#/components/schemas/AccessTokenRequest'
      responses:
        '200':
          description: 'The request was successful and a token was issued.'
  /oauth2/any-token:
    post:
      tags:
        - authorization
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/AnyTokenRequest'
          application/x-www-form-urlencoded:
            schema:
              $ref: '#/components/schemas/AnyTokenRequest'
      responses:
        '200':
          description: 'Any type of token request was successful.'
  /oauth2/all-token:
    post:
      tags:
        - authorization
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/AllTokenRequest'
          application/x-www-form-urlencoded:
            schema:
              $ref: '#/components/schemas/AllTokenRequest'
      responses:
        '200':
          description: 'All type of token request was successful.'
components:
  schemas:
    AccessTokenRequest:
      description: 'Describes all of the potential access token requests that can be received'
      type: object
      oneOf:
        - $ref: '#/components/schemas/ClientCredentialsTokenRequest'
        - $ref: '#/components/schemas/RefreshTokenRequest'
    ClientCredentialsTokenRequest:
      description: 'The client_id and client_secret properties should only be sent in form data if the client does not support basic authentication for sending client credentials.'
      type: object
      properties:
        grant_type:
          type: string
          enum:
            - client_credentials
        scope:
          type: string
        client_id:
          type: string
        client_secret:
          type: string
      required:
        - grant_type
        - scope
        - client_id
        - client_secret
    RefreshTokenRequest:
      type: object
      properties:
        grant_type:
          type: string
          enum:
            - refresh_token
        client_id:
          type: string
        refresh_token:
          type: string
      required:
        - grant_type
        - client_id
        - refresh_token
    AnyTokenRequest:
      type: object
      anyOf:
        - $ref: '#/components/schemas/ClientCredentialsTokenRequest'
        - $ref: '#/components/schemas/RefreshTokenRequest'
        - $ref: '#/components/schemas/AdditionalTokenRequest'
    AdditionalTokenRequest:
      type: object
      properties:
        grant_type:
          type: string
          enum:
            - additional_grant
        additional_info:
          type: string
      required:
        - grant_type
        - additional_info
    AllTokenRequest:
      type: object
      allOf:
        - $ref: '#/components/schemas/ClientCredentialsTokenRequest'
        - $ref: '#/components/schemas/TrackingInfo'
    TrackingInfo:
      type: object
      properties:
        tracking_id:
          type: string
      required:
        - tracking_id
    `[1:]

	loader := openapi3.NewLoader()

	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	for _, testcase := range []struct {
		endpoint   string
		ct         string
		data       string
		shouldFail bool
	}{
		// application/json
		{
			endpoint:   "/oauth2/token",
			ct:         "application/json",
			data:       `{"grant_type":"client_credentials", "scope":"testscope", "client_id":"myclient", "client_secret":"mypass"}`,
			shouldFail: false,
		},
		{
			endpoint:   "/oauth2/token",
			ct:         "application/json",
			data:       `{"grant_type":"client_credentials", "scope":"testscope", "client_id":"myclient", "client_secret":"mypass","request":1}`,
			shouldFail: false,
		},
		{
			endpoint:   "/oauth2/any-token",
			ct:         "application/json",
			data:       `{"grant_type":"client_credentials", "scope":"testscope", "client_id":"myclient", "client_secret":"mypass"}`,
			shouldFail: false,
		},
		{
			endpoint:   "/oauth2/any-token",
			ct:         "application/json",
			data:       `{"grant_type":"refresh_token", "client_id":"myclient", "refresh_token":"someRefreshToken"}`,
			shouldFail: false,
		},
		{
			endpoint:   "/oauth2/any-token",
			ct:         "application/json",
			data:       `{"grant_type":"additional_grant", "additional_info":"extraInfo"}`,
			shouldFail: false,
		},
		{
			endpoint:   "/oauth2/any-token",
			ct:         "application/json",
			data:       `{"grant_type":"invalid_grant", "extra_field":"extraValue"}`,
			shouldFail: true,
		},
		{
			endpoint: "/oauth2/all-token",
			ct:       "application/json",
			data: `{
		      "grant_type": "client_credentials",
		      "scope": "testscope",
		      "client_id": "myclient",
		      "client_secret": "mypass",
		      "tracking_id": "123456"
		  }`,
			shouldFail: false,
		},

		{
			endpoint:   "/oauth2/all-token",
			ct:         "application/json",
			data:       `{"grant_type":"invalid", "client_id":"myclient", "extra_field":"extraValue"}`,
			shouldFail: true,
		},

		// application/x-www-form-urlencoded
		{
			endpoint:   "/oauth2/token",
			ct:         "application/x-www-form-urlencoded",
			data:       "grant_type=client_credentials&scope=testscope&client_id=myclient&client_secret=mypass",
			shouldFail: false,
		},
		{
			endpoint:   "/oauth2/token",
			ct:         "application/x-www-form-urlencoded",
			data:       "grant_type=client_credentials&scope=testscope&client_id=myclient&client_secret=mypass&request=1",
			shouldFail: false,
		},
		{
			endpoint:   "/oauth2/token",
			ct:         "application/x-www-form-urlencoded",
			data:       "invalid_field=invalid_value",
			shouldFail: true,
		},
		{
			endpoint:   "/oauth2/any-token",
			ct:         "application/x-www-form-urlencoded",
			data:       "grant_type=client_credentials&scope=testscope&client_id=myclient&client_secret=mypass",
			shouldFail: false,
		},
		{
			endpoint:   "/oauth2/any-token",
			ct:         "application/x-www-form-urlencoded",
			data:       "grant_type=refresh_token&client_id=myclient&refresh_token=someRefreshToken",
			shouldFail: false,
		},
		{
			endpoint:   "/oauth2/any-token",
			ct:         "application/x-www-form-urlencoded",
			data:       "grant_type=additional_grant&additional_info=extraInfo",
			shouldFail: false,
		},
		{
			endpoint:   "/oauth2/any-token",
			ct:         "application/x-www-form-urlencoded",
			data:       "grant_type=invalid_grant&extra_field=extraValue",
			shouldFail: true,
		},
		{
			endpoint:   "/oauth2/all-token",
			ct:         "application/x-www-form-urlencoded",
			data:       "grant_type=client_credentials&scope=testscope&client_id=myclient&client_secret=mypass&tracking_id=123456",
			shouldFail: false,
		},
		{
			endpoint:   "/oauth2/all-token",
			ct:         "application/x-www-form-urlencoded",
			data:       "grant_type=invalid&client_id=myclient&extra_field=extraValue",
			shouldFail: true,
		},
	} {
		t.Run(testcase.ct, func(t *testing.T) {
			data := strings.NewReader(testcase.data)
			req, err := http.NewRequest("POST", testcase.endpoint, data)
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
			if testcase.shouldFail {
				require.Error(t, err, "This test case should fail "+testcase.data)
			} else {
				require.NoError(t, err, "This test case should pass "+testcase.data)
			}
		})
	}
}
