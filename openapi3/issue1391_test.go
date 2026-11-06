package openapi3_test

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// A file reached through whole-file $refs along many paths is read and
// resolved once, not once per path: https://github.com/getkin/kin-openapi/issues/1391
func TestIssue1391WholeFileRefReadOncePerFile(t *testing.T) {
	t.Parallel()

	// A ladder of layers: both schemas of a layer refer to both schemas of the
	// next one, so the number of paths to the last layer doubles with each layer.
	const layers = 12
	dir := t.TempDir()
	for k := range layers {
		for _, s := range []string{"a", "b"} {
			content := "type: string\n"
			if k < layers-1 {
				content = fmt.Sprintf("type: object\nproperties:\n  a:\n    $ref: './layer%d_a.yaml'\n  b:\n    $ref: './layer%d_b.yaml'\n", k+1, k+1)
			}
			require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("layer%d_%s.yaml", k, s)), []byte(content), 0o644))
		}
	}
	root := `openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
components:
  schemas:
    Root:
      $ref: './layer0_a.yaml'
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(root), 0o644))

	reads := make(map[string]int)
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, location *url.URL) ([]byte, error) {
		reads[location.Path]++
		return openapi3.ReadFromFile(loader, location)
	}
	doc, err := loader.LoadFromFile(filepath.Join(dir, "openapi.yaml"))
	require.NoError(t, err)

	// The root and every schema file but layer0_b, which nothing refers to.
	require.Len(t, reads, 2*layers)
	for location, n := range reads {
		require.Equal(t, 1, n, "%s read %d times", location, n)
	}

	// Every path still ends at the last layer's schema.
	schema := doc.Components.Schemas["Root"].Value
	for range layers - 1 {
		require.NotNil(t, schema.Properties["a"].Value)
		require.NotNil(t, schema.Properties["b"].Value)
		schema = schema.Properties["b"].Value
	}
	require.True(t, schema.Type.Is("string"))
}
