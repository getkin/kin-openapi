package openapi3filter_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// GHSA-mmfr-pmjx-hw9w: a nil-pointer dereference in convertParseError (reached
// via ConvertErrors / ValidationErrorEncoder) lets an unauthenticated client
// crash a server with a single multipart/form-data request.
//
// convertParseError dereferences e.Parameter.In in the "query" branch without a
// nil guard. For request-body errors e.Parameter is always nil, and only the
// multipart decoder produces the nested *ParseError shape that reaches that
// branch, so a malformed scalar part (e.g. a non-numeric value for an integer
// property) panics.
//
// Sending the crafted request through ValidateRequest and then ConvertErrors
// must not panic; it must return a 400 with a meaningful message. The
// application/json control cases confirm the JSON body path is unaffected.
func TestGHSA_mmfr_pmjx_hw9w_ConvertErrors_NoPanic(t *testing.T) {
	spec := `
openapi: '3.0.3'
info:
  title: t
  version: '1.0.0'
paths:
  /upload:
    post:
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              properties:
                age: {type: integer}
          application/json:
            schema:
              type: object
              properties:
                age: {type: integer}
      responses:
        '200': {description: ok}
`[1:]

	loader := openapi3.NewLoader()
	ctx := loader.Context

	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(ctx))

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	multipartBody := func() (string, string) {
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		require.NoError(t, w.WriteField("age", "notanumber"))
		require.NoError(t, w.Close())
		return w.FormDataContentType(), buf.String()
	}

	for _, tc := range []struct {
		name        string
		contentType string
		body        string
		wantStatus  int
	}{
		{
			// The crash: multipart scalar part fails primitive parsing.
			// e.Parameter == nil and the cause is a nested *ParseError.
			name:       "multipart malformed scalar (the crash)",
			body:       "notanumber",
			wantStatus: http.StatusBadRequest,
		},
		{
			// Control: malformed JSON body. *ParseError whose cause is an
			// encoding/json error, so the type assertion fails and the safe
			// fallback branch is taken.
			name:        "json malformed body",
			contentType: "application/json",
			body:        `{"age": `,
			wantStatus:  http.StatusBadRequest,
		},
		{
			// Control: wrong-type JSON body. Routed through convertSchemaError
			// (422), never reaching convertParseError.
			name:        "json wrong-type body",
			contentType: "application/json",
			body:        `{"age": "notanumber"}`,
			wantStatus:  http.StatusUnprocessableEntity,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			contentType, body := tc.contentType, tc.body
			if contentType == "" {
				contentType, body = multipartBody()
			}

			req, err := http.NewRequest(http.MethodPost, "/upload", strings.NewReader(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", contentType)

			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			reqErr := openapi3filter.ValidateRequest(ctx, &openapi3filter.RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
				Options:    &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc},
			})
			require.Error(t, reqErr)

			// This is what a typical error-rendering middleware calls; it must
			// not panic.
			var converted error
			require.NotPanics(t, func() {
				converted = openapi3filter.ConvertErrors(reqErr)
			})

			var validationErr *openapi3filter.ValidationError
			require.ErrorAs(t, converted, &validationErr)
			require.Equal(t, tc.wantStatus, validationErr.Status)
			require.NotEmpty(t, validationErr.Title, "converted error should carry a meaningful message")
		})
	}
}
