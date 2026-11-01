package openapi3

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoaderWholeFileSchemaRefLocalFragment(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"3.0.3", "3.1.0"} {
		for _, format := range []string{"yaml", "json"} {
			for _, fragment := range []string{"", "#"} {
				for _, missing := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/alias%s/missing=%t", version, format, fragment, missing), func(t *testing.T) {
						t.Parallel()
						dir := t.TempDir()
						require.NoError(t, os.Mkdir(filepath.Join(dir, "schemas"), 0o700))
						root := fmt.Sprintf(`openapi: %s
info: {title: Local schema reference, version: '1'}
paths: {}
components:
  schemas:
    Alias: {$ref: './schemas/alias.%s%s'}
    Target: {type: integer}
`, version, format, fragment)
						name := "Target"
						if missing {
							name = "Other"
						}
						alias := fmt.Sprintf(`$ref: '#/components/schemas/Target'
components:
  schemas:
    %s: {type: string, minLength: 3}
`, name)
						if format == "json" {
							alias = fmt.Sprintf(`{"$ref":"#/components/schemas/Target","components":{"schemas":{"%s":{"type":"string","minLength":3}}}}`, name)
						}
						require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(root), 0o600))
						require.NoError(t, os.WriteFile(filepath.Join(dir, "schemas", "alias."+format), []byte(alias), 0o600))

						loader := NewLoader()
						loader.IsExternalRefsAllowed = true
						loader.IncludeOrigin = true
						doc, err := loader.LoadFromFile(filepath.Join(dir, "openapi.yaml"))
						if missing {
							require.ErrorContains(t, err, "Target")
							return
						}
						require.NoError(t, err)
						require.NoError(t, doc.Validate(loader.Context))
						schema := doc.Components.Schemas["Alias"].Value
						require.NotNil(t, schema)
						require.NoError(t, schema.VisitJSON("abc"))
						require.Error(t, schema.VisitJSON("ab"))
						require.Error(t, schema.VisitJSON(42))

						doc.InternalizeRefs(loader.Context, nil)
						data, err := json.Marshal(doc)
						require.NoError(t, err)
						bundledLoader := NewLoader()
						bundled, err := bundledLoader.LoadFromData(data)
						require.NoError(t, err)
						require.NoError(t, bundled.Validate(bundledLoader.Context))
						schema = bundled.Components.Schemas["Alias"].Value
						require.NotNil(t, schema)
						require.NoError(t, schema.VisitJSON("abc"))
						require.Error(t, schema.VisitJSON("ab"))
						require.Error(t, schema.VisitJSON(42))
					})
				}
			}
		}
	}
}
