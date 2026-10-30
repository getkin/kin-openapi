package openapi3_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestInternalizeRefsDereferencesComponentRefsToFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	for name, content := range map[string]string{
		"root.yaml": `openapi: 3.0.3
info: {title: t, version: '1'}
security: [{bearerAuth: []}]
paths:
  /a:
    get:
      operationId: getA
      responses:
        '200':
          description: OK
          headers:
            X-Header: {$ref: '#/components/headers/X-Header'}
          content:
            application/json:
              schema: {type: integer}
              examples:
                example: {$ref: '#/components/examples/Example'}
          links:
            link: {$ref: '#/components/links/Link'}
        '404': {$ref: '#/components/responses/NotFound'}
components:
  headers:
    X-Header: {$ref: ./header.yaml}
  responses:
    NotFound: {$ref: ./response.yaml}
  securitySchemes:
    bearerAuth: {$ref: ./securityScheme.yaml}
  examples:
    Example: {$ref: ./example.yaml}
  links:
    Link: {$ref: ./link.yaml}
`,
		"header.yaml":         "schema: {type: string}\n",
		"response.yaml":       "description: Not found\n",
		"securityScheme.yaml": "type: http\nscheme: bearer\n",
		"example.yaml":        "value: 42\n",
		"link.yaml":           "operationId: getA\n",
	} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile(filepath.Join(dir, "root.yaml"))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(t.Context()))

	doc.InternalizeRefs(t.Context(), nil)
	require.NoError(t, doc.Validate(t.Context()))

	components := doc.Components
	require.Empty(t, components.Headers["X-Header"].Ref)
	require.Empty(t, components.Responses["NotFound"].Ref)
	require.Empty(t, components.SecuritySchemes["bearerAuth"].Ref)
	require.Empty(t, components.Examples["Example"].Ref)
	require.Empty(t, components.Links["Link"].Ref)
	require.Equal(t, "Not found", *components.Responses["NotFound"].Value.Description)

	operation := doc.Paths.Value("/a").Get
	require.Equal(t, "#/components/headers/X-Header", operation.Responses.Value("200").Value.Headers["X-Header"].Ref)
	require.Equal(t, "#/components/examples/Example", operation.Responses.Value("200").Value.Content["application/json"].Examples["example"].Ref)
	require.Equal(t, "#/components/links/Link", operation.Responses.Value("200").Value.Links["link"].Ref)
	require.Equal(t, "#/components/responses/NotFound", operation.Responses.Value("404").Ref)

	data, err := doc.MarshalJSON()
	require.NoError(t, err)
	loader = openapi3.NewLoader()
	reloaded, err := loader.LoadFromData(data)
	require.NoError(t, err)
	require.NoError(t, reloaded.Validate(t.Context()))
}
