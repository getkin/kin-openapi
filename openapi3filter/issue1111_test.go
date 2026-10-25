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

// issue1111Bodies are request bodies with keys that only differ in case from
// the "value" property, which only accepts "aniki".
var issue1111Bodies = map[string]string{
	"good":         `{"value": "aniki"}`,
	"bad":          `{"value": "beniki"}`,
	"evil":         `{"value": "aniki", "Value": "beniki"}`,
	"chaotic evil": `{"Value": "beniki", "vaLue": "chaos"}`,
}

// TestIssue1111 checks that validation matches property names to keys
// case-sensitively, as JSON does: keys that only differ in case from a
// property are other properties, which only additionalProperties: false rejects.
func TestIssue1111(t *testing.T) {
	t.Parallel()

	for _, closed := range []bool{false, true} {
		spec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                value:
                  type: string
                  enum:
                    - "aniki"
`
		if closed {
			spec += "              additionalProperties: false\n"
		}
		spec += `      responses:
        '204':
          description: OK
`

		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromData([]byte(spec))
		require.NoError(t, err)
		require.NoError(t, doc.Validate(t.Context()))
		router, err := gorillamux.NewRouter(doc)
		require.NoError(t, err)

		for name, body := range issue1111Bodies {
			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			err = openapi3filter.ValidateRequest(t.Context(), &openapi3filter.RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			})
			switch {
			case name == "good":
				require.NoError(t, err, name)
			case name == "bad" || closed:
				require.Error(t, err, name)
			default:
				require.NoError(t, err, name)
			}
		}
	}
}
