package openapi3filter_test

import (
	"bytes"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// maxFuzzInput bounds fuzzer inputs so the fuzzer spends its budget on
// conversion logic rather than on allocating huge documents.
const maxFuzzInput = 32 << 10

func newFuzzLoader() *openapi3.Loader {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(*openapi3.Loader, *url.URL) ([]byte, error) { return []byte(`{}`), nil } // deterministic
	return loader
}

func FuzzValidateRequest(f *testing.F) {
	paths, err := filepath.Glob("../openapi3/testdata/*.yml")
	require.NoError(f, err)
	for _, path := range paths {
		data, err := os.ReadFile(path)
		require.NoError(f, err)
		if len(data) <= 16<<10 {
			f.Add(data, "GET", "/", "application/json", []byte(`{}`))
		}
	}

	const spec = `
openapi: 3.0.0
info: {title: t, version: "1"}
paths:
  /items/{id}:
    post:
      parameters:
      - {name: id, in: path, required: true, schema: {type: integer}}
      - {name: q, in: query, schema: {type: array, items: {type: string}}}
      requestBody:
        content:
          application/json: {schema: {type: object, required: [n], properties: {n: {type: number}}}}
          multipart/form-data: {schema: {type: object, properties: {f: {type: string, format: binary}}}}
          application/x-www-form-urlencoded: {schema: {type: object, properties: {a: {type: string}}}}
      responses: {"200": {description: ok}}
`
	f.Add([]byte(spec), "POST", "/items/1?q=a&q=b", "application/json", []byte(`{"n":1}`))
	f.Add([]byte(spec), "POST", "/items/x", "application/json", []byte(`{"n":"nope"}`))
	f.Add([]byte(spec), "POST", "/items/1", "application/x-www-form-urlencoded", []byte(`a=b&a=c`))
	f.Add([]byte(spec), "POST", "/items/1", "multipart/form-data; boundary=b", []byte("--b\r\nContent-Disposition: form-data; name=\"f\"\r\n\r\nx\r\n--b--\r\n"))

	f.Fuzz(func(t *testing.T, data []byte, method, target, contentType string, body []byte) {
		if len(data) > maxFuzzInput || len(body) > maxFuzzInput {
			t.Skip()
		}
		if len(method) > 64 || len(target) > 4<<10 || len(contentType) > 1<<10 {
			t.Skip()
		}

		loader := newFuzzLoader()
		doc, err := loader.LoadFromData(data)
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

		req, err := http.NewRequest(method, target, bytes.NewReader(body))
		if err != nil {
			t.Skip()
		}
		req.Header.Set("Content-Type", contentType)

		route, pathParams, err := router.FindRoute(req)
		if err != nil {
			return
		}

		// A matched route must be usable by callers: every field they dereference
		// has to be populated, and every path parameter of the matched template
		// has to be present.
		require.NotNil(t, route)
		require.NotNil(t, route.Spec)
		require.NotNil(t, route.PathItem)
		require.NotNil(t, route.Operation)
		require.Equal(t, req.Method, route.Method)
		requirePathParamsCoverTemplate(t, route, pathParams)

		// No crash in here
		_ = openapi3filter.ValidateRequest(t.Context(), &openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
			Options: &openapi3filter.Options{
				MultiError:         true,
				AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
			},
		})
	})
}

func requirePathParamsCoverTemplate(t *testing.T, route *routers.Route, pathParams map[string]string) {
	t.Helper()
	for _, name := range muxVarNames(route.Path) {
		require.Contains(t, pathParams, name, "path %q", route.Path)
	}
}

// muxVarNames extracts the variable names gorilla/mux reads out of a route
// template. It must stay in sync with mux's own parsing: braces nest, so
// "{{a}}" holds the single variable "{a}", and everything after the first
// colon is a match pattern rather than part of the name.
func muxVarNames(tpl string) (names []string) {
	level, idx := 0, 0
	for i := 0; i < len(tpl); i++ {
		switch tpl[i] {
		case '{':
			if level++; level == 1 {
				idx = i
			}
		case '}':
			if level--; level == 0 {
				name, _, _ := strings.Cut(tpl[idx+1:i], ":")
				names = append(names, name)
			} else if level < 0 {
				names = nil
				return
			}
		}
	}
	if level != 0 {
		names = nil
	}
	return
}
