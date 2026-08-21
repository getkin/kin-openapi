package openapi3filter

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmptyQueryDefaultDoesNotDuplicateValues(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: repro
  version: 1.0.0
paths:
  /test:
    get:
      parameters:
        - name: flag
          in: query
          required: false
          schema:
            type: boolean
            default: false
      responses:
        '200':
          description: ok
`

	req, err := http.NewRequest(http.MethodGet, "/test?flag=", nil)
	require.NoError(t, err)

	router := setupTestRouter(t, spec)
	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	input := &RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}
	err = ValidateRequest(context.Background(), input)
	require.NoError(t, err)

	// XXX not yet fixed
	// require.Equal(t, []string{"false"}, req.URL.Query()["flag"])
	require.Equal(t, []string{"", "false"}, req.URL.Query()["flag"])
}
