package openapi3filter_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// jsonschema/v6 v6.0.2 panicked with a nil-pointer dereference in
// numValidate when validating math.NaN() / math.Inf() float64 values
// against numeric constraints. A YAML request body containing .nan,
// .inf, or -.inf against an OAS 3.1 spec (which routes through
// useJSONSchema2020, bypassing the built-in NaN/Inf guard in visitJSON)
// therefore crashed the request-serving goroutine. jsonschema v6.0.3
// returns a normal validation error instead.

const yamlNaNInfNumberSpec = `
openapi: 3.1.0
info: {title: t, version: 1.0.0}
paths:
  /n:
    post:
      requestBody:
        required: true
        content:
          application/yaml:
            schema:
              type: number
              minimum: 0
      responses: {"200": {description: ok}}
`

func TestYAMLNaNInfNoPanic(t *testing.T) {
	t.Parallel()

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(yamlNaNInfNumberSpec))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(loader.Context))
	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	for _, body := range []string{".nan", ".inf", "-.inf"} {
		body := body
		t.Run(body, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/n", strings.NewReader(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/yaml")
			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			var got error
			panicVal := catchPanicValue(func() {
				got = openapi3filter.ValidateRequest(t.Context(), &openapi3filter.RequestValidationInput{
					Request: req, PathParams: pathParams, Route: route,
					Options: &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc},
				})
			})
			require.Nil(t, panicVal, "must not panic")
			require.Error(t, got, "must return a clean validation error")
		})
	}
}
