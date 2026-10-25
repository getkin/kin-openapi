package openapi3filter_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

const jsonSchema2020ErrorSpec = `
openapi: 3.1.0
info:
  title: repro
  version: "1"
paths:
  /things:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                n:
                  type: integer
      responses:
        "204":
          description: done
`

// A request body refused by a 3.1 document names the failing value and keyword
// in its SchemaError, as one refused by a 3.0 document does: a caller answers
// 400 for a body that is not the object and 422 for a broken field by them.
func TestJSONSchema2020_RequestBodyErrorCarriesItsStructure(t *testing.T) {
	t.Parallel()

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(jsonSchema2020ErrorSpec))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(loader.Context))

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	tests := []struct {
		name    string
		body    string
		pointer []string
		field   string
	}{
		{"a field of the wrong type", `{"n":"x"}`, []string{"n"}, "type"},
		{"a body that is not an object", `[]`, nil, "type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/things", strings.NewReader(tt.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			err = openapi3filter.ValidateRequest(t.Context(), &openapi3filter.RequestValidationInput{
				Request: req, PathParams: pathParams, Route: route,
			})
			require.Error(t, err)

			var schemaErr *openapi3.SchemaError
			require.True(t, errors.As(err, &schemaErr))
			pointer := schemaErr.JSONPointer()
			if len(pointer) == 0 {
				pointer = nil
			}
			require.Equal(t, tt.pointer, pointer)
			require.Equal(t, tt.field, schemaErr.SchemaField)
		})
	}
}
