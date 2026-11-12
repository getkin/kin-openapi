package routers_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/getkin/kin-openapi/routers/legacy"
)

// A URL fragment is not part of the path: it must not leak into path
// parameters, with servers declared or not.
func TestRoutersIgnoreURLFragment(t *testing.T) {
	for name, servers := range map[string]string{
		"no servers":      ``,
		"relative server": "servers: [{url: /}]",
		"absolute server": "servers: [{url: 'http://example.com/api'}]",
	} {
		t.Run(name, func(t *testing.T) {
			loader := openapi3.NewLoader()
			doc, err := loader.LoadFromData([]byte(`
openapi: 3.0.0
info: {title: t, version: "1"}
` + servers + `
paths:
  /items/{id}: {get: {parameters: [{name: id, in: path, required: true, schema: {type: string}}], responses: {"200": {description: ok}}}}
`))
			require.NoError(t, err)
			mux, err := gorillamux.NewRouter(doc)
			require.NoError(t, err)
			leg, err := legacy.NewRouter(doc)
			require.NoError(t, err)

			target := "/items/42#frag/ment"
			if name == "absolute server" {
				target = "http://example.com/api" + target
			}
			req, err := http.NewRequest("GET", target, nil)
			require.NoError(t, err)

			_, muxParams, err := mux.FindRoute(req)
			require.NoError(t, err)
			_, legParams, err := leg.FindRoute(req)
			require.NoError(t, err)
			require.Equal(t, "42", muxParams["id"])
			require.Equal(t, muxParams["id"], legParams["id"])
		})
	}
}
