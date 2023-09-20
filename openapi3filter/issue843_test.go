package openapi3filter_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestIssue843(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: 'Validator'
  version: 0.0.1
paths:
  /test:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                description:
                  type: string
      responses:
        '200':
          description: Created
`

	loader := openapi3.NewLoader()

	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte(`{"description":"some description"}`)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/html")

	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	err = openapi3filter.ValidateRequest(
		context.Background(),
		&openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
		})

	if assert.Error(t, err) {
		var serr *openapi3filter.RequestError
		if assert.ErrorAs(t, err, &serr) {
			message := serr.Error()
			assert.NotEqual(t, `request body has an error: header Content-Type has unexpected value ""`, message)
			assert.NotNil(t, serr.Parameter)
			if serr.Parameter != nil {
				assert.Equal(t, "header", serr.Parameter.In)
				assert.Equal(t, "content-type", serr.Parameter.Name)
				assert.Equal(t, "invalid content-type", serr.Reason)
			}
		}
	}
}
