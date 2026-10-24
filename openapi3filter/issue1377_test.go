package openapi3filter

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue1377(t *testing.T) {
	t.Parallel()

	const spec = `
openapi: 3.1.0
info: {title: repro, version: "1"}
paths:
  /config:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                interval:
                  type: string
                  enum: [day, week]
      responses:
        "204": {description: ok}
`

	req, err := http.NewRequest(http.MethodPut, "/config", strings.NewReader(`{"interval":"hour"}`))
	require.NoError(t, err)
	req.Header.Set(headerCT, "application/json")

	router := setupTestRouter(t, spec)
	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	err = ValidateRequest(t.Context(), &RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	})
	require.Error(t, err)

	verr := ConvertErrors(err).(*ValidationError)
	require.Equal(t, http.StatusUnprocessableEntity, verr.Status)
	require.Equal(t, &ValidationErrorSource{Pointer: "/interval"}, verr.Source)
	require.Equal(t, "value hour at /interval must be one of: day, week", verr.Detail)
}
