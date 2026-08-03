package openapi3filter_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
)

const specBody = `
openapi: 3.0.3
info: {title: poc, version: "1.0.0"}
paths:
  /x:
    get:
      responses:
        "200":
          description: ok
          content:
            application/json: {}
`

const specHeader = `
openapi: 3.0.3
info: {title: poc, version: "1.0.0"}
paths:
  /x:
    get:
      responses:
        "200":
          description: ok
          headers:
            X-Thing:
              content:
                application/json: {}
`

func validatedInput(t *testing.T, spec string, hdr http.Header) *openapi3filter.ResponseValidationInput {
	t.Helper()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)
	err = doc.Validate(t.Context())
	require.NoError(t, err)
	op := doc.Paths.Find("/x").Get
	return &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request: httptest.NewRequest(http.MethodGet, "/x", nil),
			Route:   &routers.Route{Spec: doc, Operation: op, Method: http.MethodGet, Path: "/x"},
		},
		Status: 200,
		Header: hdr,
		Body:   io.NopCloser(strings.NewReader(`{}`)),
	}
}

func TestControl_ResponseBodyNilSchema(t *testing.T) {
	in := validatedInput(t, specBody, http.Header{"Content-Type": {"application/json"}})
	err := openapi3filter.ValidateResponse(t.Context(), in)
	require.NoError(t, err)
}

func TestResponseHeaderNilSchema(t *testing.T) {
	in := validatedInput(t, specHeader, http.Header{"Content-Type": {"application/json"}})
	err := openapi3filter.ValidateResponse(t.Context(), in)
	require.NoError(t, err)
}
