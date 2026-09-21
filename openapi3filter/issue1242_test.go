package openapi3filter

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue1242ResponseRejectsUnevaluatedProperties(t *testing.T) {
	const spec = `
openapi: 3.1.0
info: {title: repro, version: "1"}
paths:
  /probe:
    get:
      operationId: probe
      responses:
        "502":
          description: failure
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Composed"
components:
  schemas:
    Composed:
      unevaluatedProperties: false
      allOf:
        - $ref: "#/components/schemas/Base"
        - type: object
          required: [extra]
          properties:
            extra: {type: integer}
    Base:
      type: object
      required: [code]
      properties:
        code: {type: string}
`
	router := setupTestRouter(t, spec)
	req, err := http.NewRequest(http.MethodGet, "/probe", nil)
	require.NoError(t, err)
	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)
	for _, test := range []struct {
		body string
		bad  bool
	}{
		{`{"code":"c","extra":1,"undeclared":"surprise"}`, true},
		{`{"code":"c","extra":1}`, false},
	} {
		t.Run(test.body, func(t *testing.T) {
			err := ValidateResponse(t.Context(), &ResponseValidationInput{
				RequestValidationInput: &RequestValidationInput{Request: req, PathParams: pathParams, Route: route},
				Status:                 http.StatusBadGateway,
				Header:                 http.Header{"Content-Type": []string{"application/json"}},
				Body:                   io.NopCloser(strings.NewReader(test.body)),
			})
			if test.bad {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
