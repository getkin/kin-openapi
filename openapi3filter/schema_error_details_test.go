package openapi3filter

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func schemaErrorDetailsSpec(version string) string {
	return `
openapi: ` + version + `
info:
  title: repro
  version: "1"
paths:
  /probe:
    post:
      parameters:
        - name: token
          in: query
          schema: {type: string, pattern: "^[A-Z]+$"}
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                token: {type: string, pattern: "^[A-Z]+$"}
      responses:
        "200":
          description: ok
`
}

func TestDisableSchemaErrorDetails(t *testing.T) {
	router := setupTestRouter(t, schemaErrorDetailsSpec("3.0.3"))

	validate := func(query, body string, options *Options) error {
		req, err := http.NewRequest(http.MethodPost, "/probe?"+query, strings.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		route, pathParams, err := router.FindRoute(req)
		require.NoError(t, err)
		return ValidateRequest(t.Context(), &RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
			Options:    options,
		})
	}
	disabled := &Options{
		SchemaValidationOptions: []openapi3.SchemaValidationOption{openapi3.DisableSchemaErrorDetails()},
	}

	t.Run("parameter", func(t *testing.T) {
		err := validate("token=secret", `{}`, &Options{})
		require.ErrorContains(t, err, "secret")
		e := validate("token=secret", `{}`, disabled)
		require.Error(t, e)
		msg := e.Error()
		require.NotContains(t, msg, "secret")
	})

	t.Run("body", func(t *testing.T) {
		err := validate("", `{"token":"secret"}`, &Options{})
		require.ErrorContains(t, err, "secret")
		e := validate("", `{"token":"secret"}`, disabled)
		require.Error(t, e)
		msg := e.Error()
		require.NotContains(t, msg, "secret")
	})
}

func TestDisableSchemaErrorDetailsOnOpenAPI31(t *testing.T) {
	router := setupTestRouter(t, schemaErrorDetailsSpec("3.1.0"))

	req, err := http.NewRequest(http.MethodPost, "/probe?token=VALID", strings.NewReader(`{"token":"VALID"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)
	input := &RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
		Options: &Options{
			SchemaValidationOptions: []openapi3.SchemaValidationOption{openapi3.DisableSchemaErrorDetails()},
		},
	}

	err = ValidateRequest(t.Context(), input)
	require.ErrorIs(t, err, openapi3.ErrSchemaErrorDetailsWithJSONSchema2020)
}
