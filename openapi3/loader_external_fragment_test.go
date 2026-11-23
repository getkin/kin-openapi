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

// A schema in a file of bare schemas (not an OpenAPI document), reached through
// $refs to its fragment along many paths, is read and resolved a fixed number of
// times, not once per path.
func TestExternalFragmentRefToBareSchemaFileResolvedOnce(t *testing.T) {
	t.Parallel()

	// A ladder of files: both schemas of a file refer to both schemas of the
	// next one, so the number of paths to the last file doubles with each file.
	const layers = 12
	dir := t.TempDir()
	for k := range layers {
		content := ""
		for _, s := range []string{"A", "B"} {
			schema := "type: string"
			if k < layers-1 {
				schema = fmt.Sprintf("type: object\n  properties:\n    a: {$ref: './layer%d.yaml#/A'}\n    b: {$ref: './layer%d.yaml#/B'}", k+1, k+1)
			}
			content += s + ":\n  " + schema + "\n"
		}
		require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("layer%d.yaml", k)), []byte(content), 0o644))
	}
	root := `openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
components:
  schemas:
    Root:
      $ref: './layer0.yaml#/A'
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

	require.Len(t, reads, layers+1)
	for location, n := range reads {
		// A file of bare schemas is read once as a document and once more for
		// its fragments, however many paths lead to it.
		require.LessOrEqual(t, n, 2, "%s read %d times", location, n)
	}

	// Every path still ends at the last file's schema.
	schema := doc.Components.Schemas["Root"].Value
	for range layers - 1 {
		require.NotNil(t, schema.Properties["a"].Value)
		require.NotNil(t, schema.Properties["b"].Value)
		schema = schema.Properties["b"].Value
	}
	require.True(t, schema.Type.Is("string"))
}
