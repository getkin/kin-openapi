package routers_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/internal/fuzzing"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/getkin/kin-openapi/routers/legacy"
)

// maxFuzzInput bounds fuzzer inputs so the fuzzer spends its budget on
// routing rather than on allocating huge documents.
const maxFuzzInput = 32 << 10

// Neither router panics, and when both find a route for a request they agree
// on which one it is and on the values of its path parameters.
func FuzzRoutersAgree(f *testing.F) {
	for _, seed := range fuzzing.Testdata(f, "../openapi3/testdata") {
		f.Add(seed, "GET", "/")
	}
	const spec = `
openapi: 3.0.0
info: {title: t, version: "1"}
servers:
- url: http://example.com/api/{v}
  variables: {v: {default: v1, enum: [v1, v2]}}
- url: /
paths:
  /items: {get: {responses: {"200": {description: ok}}}}
  /items/{id}: {get: {parameters: [{name: id, in: path, required: true, schema: {type: string}}], responses: {"200": {description: ok}}}}
  /items/{id}/tags/{tag}: {delete: {parameters: [{name: id, in: path, required: true, schema: {type: string}}, {name: tag, in: path, required: true, schema: {type: string}}], responses: {"200": {description: ok}}}}
  /files/{name}.{ext}: {get: {parameters: [{name: name, in: path, required: true, schema: {type: string}}, {name: ext, in: path, required: true, schema: {type: string}}], responses: {"200": {description: ok}}}}
`
	f.Add([]byte(spec), "GET", "http://example.com/api/v1/items/42")
	f.Add([]byte(spec), "GET", "http://example.com/api/v3/items")
	f.Add([]byte(spec), "DELETE", "/items/a%2Fb/tags/x")
	f.Add([]byte(spec), "GET", "/files/report.tar.gz")
	f.Add([]byte(spec), "GET", "/items/")

	f.Fuzz(func(t *testing.T, data []byte, method, target string) {
		if len(data) > maxFuzzInput || len(method) > 64 || len(target) > 4<<10 {
			t.Skip()
		}

		loader := openapi3.NewLoader()
		loader.IsExternalRefsAllowed = true
		loader.ReadFromURIFunc = func(*openapi3.Loader, *url.URL) ([]byte, error) { return []byte(`{}`), nil }
		doc, err := loader.LoadFromData(data)
		if err != nil {
			t.Skip()
		}
		if err := doc.Validate(t.Context()); err != nil {
			t.Skip()
		}
		req, err := http.NewRequest(method, target, nil)
		if err != nil {
			t.Skip()
		}

		mux, muxErr := gorillamux.NewRouter(doc)
		leg, legErr := legacy.NewRouter(doc)
		if muxErr != nil || legErr != nil {
			return
		}

		muxRoute, muxParams, muxErr := mux.FindRoute(req)
		legRoute, legParams, legErr := leg.FindRoute(req)
		if muxErr != nil || legErr != nil {
			return
		}
		require.Equal(t, muxRoute.Method, legRoute.Method)
		require.Equal(t, muxRoute.Path, legRoute.Path, "target: %s", target)
		for name, value := range legParams {
			// gorillamux leaves every value escaped, legacy only "%2F": compare
			// them decoded.
			muxValue, err := url.PathUnescape(muxParams[name])
			require.NoError(t, err)
			legValue, err := url.PathUnescape(value)
			require.NoError(t, err)
			require.Equal(t, legValue, muxValue, "parameter %q of %s for %s", name, muxRoute.Path, target)
		}
	})
}
