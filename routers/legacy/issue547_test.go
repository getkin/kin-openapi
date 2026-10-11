package legacy_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
)

func TestIssue547(t *testing.T) {
	t.Parallel()

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths:
  /files/{fileId}:
    parameters:
    - name: fileId
      in: path
      required: true
      schema:
        type: string
    get:
      responses:
        '200':
          description: x
  /files/{fileId}:undelete:
    parameters:
    - name: fileId
      in: path
      required: true
      schema:
        type: string
    post:
      responses:
        '200':
          description: x
  /reports/{name}.{format}:
    parameters:
    - name: name
      in: path
      required: true
      schema:
        type: string
    - name: format
      in: path
      required: true
      schema:
        type: string
    get:
      responses:
        '200':
          description: x
`[1:]))
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	router, err := legacyrouter.NewRouter(doc)
	require.NoError(t, err)

	for _, tc := range []struct {
		method, path, route string
		params              map[string]string
	}{
		{http.MethodGet, "/files/123", "/files/{fileId}", map[string]string{"fileId": "123"}},
		{http.MethodPost, "/files/123:undelete", "/files/{fileId}:undelete", map[string]string{"fileId": "123"}},
		{http.MethodGet, "/reports/2026.10.csv", "/reports/{name}.{format}", map[string]string{"name": "2026.10", "format": "csv"}},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.path, nil)
			require.NoError(t, err)

			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)
			require.Equal(t, tc.route, route.Path)
			require.Equal(t, tc.params, pathParams)
		})
	}

	req, err := http.NewRequest(http.MethodPost, "/files/:undelete", nil)
	require.NoError(t, err)
	_, _, err = router.FindRoute(req)
	require.Error(t, err)
}
