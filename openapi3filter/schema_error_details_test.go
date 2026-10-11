package openapi3filter

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestDisableSchemaErrorDetails(t *testing.T) {
	router := setupTestRouter(t, `
openapi: 3.0.3
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
`)

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
		require.NotContains(t, e.Error(), "secret")
	})

	t.Run("body", func(t *testing.T) {
		err := validate("", `{"token":"secret"}`, &Options{})
		require.ErrorContains(t, err, "secret")
		e := validate("", `{"token":"secret"}`, disabled)
		require.Error(t, e)
		require.NotContains(t, e.Error(), "secret")
	})
}
