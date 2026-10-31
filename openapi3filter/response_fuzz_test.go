package openapi3filter_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/internal/fuzzing"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// Validating a response never panics and reaches the same verdict every
// time, and turning checks off never rejects a response the full checks
// accept.
func FuzzValidateResponse(f *testing.F) {
	seeds := fuzzing.Testdata(f, "../openapi3/testdata", "testdata")
	for _, seed := range seeds {
		f.Add(seed, "GET", "/", 200, "application/json", []byte(`{}`))
	}

	const spec = `
openapi: 3.0.0
info: {title: t, version: "1"}
paths:
  /items/{id}:
    get:
      parameters:
      - {name: id, in: path, required: true, schema: {type: integer}}
      responses:
        "200":
          description: ok
          headers:
            X-Rate: {required: true, schema: {type: integer, minimum: 0}}
            X-Tags: {schema: {type: array, items: {type: string}}, style: simple}
          content:
            application/json:
              schema:
                type: object
                required: [n]
                properties:
                  n: {type: number, readOnly: true}
                  w: {type: string, writeOnly: true}
                  d: {type: string, format: date-time}
            text/plain: {schema: {type: string, maxLength: 3}}
            application/x-www-form-urlencoded: {schema: {type: object, properties: {a: {type: string}}}}
        4XX:
          description: client error
          content:
            application/problem+json: {schema: {type: object, properties: {title: {type: string}}}}
        default:
          description: other
`
	f.Add([]byte(spec), "GET", "/items/1", 200, "application/json", []byte(`{"n":1,"d":"2020-01-01T00:00:00Z"}`))
	f.Add([]byte(spec), "GET", "/items/1", 200, "application/json", []byte(`{"w":"x"}`))
	f.Add([]byte(spec), "GET", "/items/1", 200, "text/plain; charset=utf-8", []byte(`abcd`))
	f.Add([]byte(spec), "GET", "/items/1", 200, "application/x-www-form-urlencoded", []byte(`a=b&a=c`))
	f.Add([]byte(spec), "GET", "/items/1", 404, "application/problem+json", []byte(`{"title":1}`))
	f.Add([]byte(spec), "GET", "/items/1", 503, "", []byte(``))

	f.Fuzz(func(t *testing.T, data []byte, method, target string, status int, contentType string, body []byte) {
		if len(data) > maxFuzzInput || len(body) > maxFuzzInput {
			t.Skip()
		}
		if len(method) > 64 || len(target) > 4<<10 || len(contentType) > 1<<10 {
			t.Skip()
		}

		doc, err := newFuzzLoader().LoadFromData(data)
		if err != nil {
			t.Skip()
		}
		if err = doc.Validate(t.Context()); err != nil {
			t.Skip()
		}
		router, err := gorillamux.NewRouter(doc)
		if err != nil {
			return
		}
		req, err := http.NewRequest(method, target, nil)
		if err != nil {
			t.Skip()
		}
		route, pathParams, err := router.FindRoute(req)
		if err != nil {
			return
		}

		validate := func(options openapi3filter.Options) error {
			options.MultiError = true
			options.AuthenticationFunc = openapi3filter.NoopAuthenticationFunc
			input := &openapi3filter.ResponseValidationInput{
				RequestValidationInput: &openapi3filter.RequestValidationInput{
					Request:    req,
					PathParams: pathParams,
					Route:      route,
				},
				Status:  status,
				Header:  http.Header{"Content-Type": {contentType}, "X-Rate": {"1"}},
				Options: &options,
			}
			input.SetBodyBytes(bytes.Clone(body))
			return openapi3filter.ValidateResponse(t.Context(), input)
		}

		full := validate(openapi3filter.Options{IncludeResponseStatus: true})
		again := validate(openapi3filter.Options{IncludeResponseStatus: true})
		require.Equal(t, full == nil, again == nil, "first: %v\nagain: %v", full, again)
		if full != nil {
			return
		}
		for _, lax := range []openapi3filter.Options{
			{},
			{ExcludeResponseBody: true},
			{ExcludeReadOnlyValidations: true, ExcludeWriteOnlyValidations: true},
		} {
			require.NoError(t, validate(lax), "%+v", lax)
		}
	})
}
