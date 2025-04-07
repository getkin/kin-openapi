package openapi3filter_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRequestWithAnAuthenticatorFunc_CanConsumeTheRequestBody(t *testing.T) {
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
          text/plain:
            schema:
              type: string
      responses:
        '200':
          description: Created
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
security:
  - BearerAuth: [ ]
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader("something"))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "text/plain")

	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	err = openapi3filter.ValidateRequest(
		context.Background(),
		&openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
			Options: &openapi3filter.Options{
				AuthenticationFunc: func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
					defer req.Body.Close()

					// NOTE that reading from `req.Body` appears to trigger the underlying issue raised in https://github.com/getkin/kin-openapi/issues/743
					// this doesn't seem to occur when using `req.GetBody()`, but as that's less common to do, we should support both types
					body, err := io.ReadAll(ai.RequestValidationInput.Request.Body)
					assert.NoError(t, err)

					// and make sure the parsed body was correct
					assert.Equal(t, []byte("something"), body)

					return nil
				},
			},
		},
	)

	require.NoError(t, err)
}

func TestValidateRequestWithAnAuthenticatorFunc_CanConsumeTheRequestBodyAndThenBeParsedByARouter(t *testing.T) {
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
          text/plain:
            schema:
              type: string
      responses:
        '200':
          description: Created
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
security:
  - BearerAuth: [ ]
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	called := false
	authenticatorFunc := func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
		defer ai.RequestValidationInput.Request.Body.Close()

		// NOTE that reading from `req.Body` appears to trigger the underlying issue raised in https://github.com/getkin/kin-openapi/issues/743
		// this doesn't seem to occur when using `req.GetBody()`, but as that's less common to do, we should support both types
		body, err := io.ReadAll(ai.RequestValidationInput.Request.Body)
		assert.NoError(t, err)

		// and make sure the parsed body was correct
		assert.Equal(t, []byte("something"), body)

		return nil
	}

	use := func(r *http.ServeMux, middlewares ...func(next http.Handler) http.Handler) http.Handler {
		var s http.Handler
		s = r

		for _, mw := range middlewares {
			s = mw(s)
		}

		return s
	}

	validationMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route, pathParams, err := router.FindRoute(r)
			if err != nil {
				fmt.Println(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			err = openapi3filter.ValidateRequest(r.Context(), &openapi3filter.RequestValidationInput{
				Request:    r,
				PathParams: pathParams,
				Route:      route,
				Options: &openapi3filter.Options{
					AuthenticationFunc: authenticatorFunc,
				},
			})
			require.NoError(t, err)

			next.ServeHTTP(w, r)
		})
	}

	r := http.NewServeMux()
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		called = true
		defer r.Body.Close()

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		// and make sure the parsed body was correct
		assert.Equal(t, []byte("something"), body)

		w.WriteHeader(http.StatusNoContent)
	})

	ts := httptest.NewServer(use(r, validationMiddleware))

	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/test", strings.NewReader("something"))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.True(t, called, "The POST /test endpoint should have been called, but it wasn't")
}
