package openapi3_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestInternalizeRefs(t *testing.T) {
	ctx := t.Context()

	regexpRef := regexp.MustCompile(`"\$ref":`)
	regexpRefInternal := regexp.MustCompile(`"\$ref":"#`)

	tests := []struct {
		filename string
	}{
		{"testdata/testref.openapi.yml"},
		{"testdata/recursiveRef/openapi.yml"},
		{"testdata/spec.yaml"},
		{"testdata/callbacks.yml"},
		{"testdata/issue831/testref.internalizepath.openapi.yml"},
		{"testdata/issue959/openapi.yml"},
		{"testdata/interalizationNameCollision/api.yml"},
		{"testdata/discriminator.yml"},
		{"testdata/discriminatorLocalMapping.yml"},
		{"testdata/issue1205/openapi.yml"},
		{"testdata/issue1205/mutual.yml"},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			// Load in the reference spec from the testdata
			sl := openapi3.NewLoader()
			sl.IsExternalRefsAllowed = true
			doc, err := sl.LoadFromFile(test.filename)
			require.NoError(t, err, "loading test file")
			err = doc.Validate(ctx)
			require.NoError(t, err, "validating spec")

			// Internalize the references
			doc.InternalizeRefs(ctx, nil)

			// Validate the internalized spec
			err = doc.Validate(ctx)
			require.NoError(t, err, "validating internalized spec")

			actual, err := doc.MarshalJSON()
			require.NoError(t, err, "marshaling internalized spec")

			// run a static check over the file, making sure each occurrence of a
			// reference is followed by a #
			numRefs := len(regexpRef.FindAll(actual, -1))
			numInternalRefs := len(regexpRefInternal.FindAll(actual, -1))
			require.Equal(t, numRefs, numInternalRefs, "checking all references are internal")

			// load from actual, but with the path set to the current directory
			doc2, err := sl.LoadFromData(actual)
			require.NoError(t, err, "reloading spec")
			err = doc2.Validate(ctx)
			require.NoError(t, err, "validating reloaded spec")

			// compare with expected
			expected, err := os.ReadFile(test.filename + ".internalized.yml")
			require.NoError(t, err)
			require.JSONEq(t, string(expected), string(actual))
		})
	}
}

func TestInternalizeRefsUpdatesSchemaRefSiblingFields(t *testing.T) {
	dir := t.TempDir()
	rootPath := filepath.Join(dir, "root.yaml")
	externalPath := filepath.Join(dir, "external.yaml")

	require.NoError(t, os.WriteFile(rootPath, []byte(`
openapi: 3.1.0
info:
  title: Internalized sibling refs
  version: "1.0"
paths: {}
components:
  schemas:
    Root:
      type: object
      properties:
        use:
          $ref: external.yaml#/Base
          items:
            $ref: external.yaml#/Item
`), 0o600))
	require.NoError(t, os.WriteFile(externalPath, []byte(`
Base:
  type: array
Item:
  type: string
`), 0o600))

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile(rootPath)
	require.NoError(t, err)

	doc.InternalizeRefs(t.Context(), nil)
	data, err := doc.MarshalJSON()
	require.NoError(t, err)
	require.NotContains(t, string(data), "external.yaml")

	use := doc.Components.Schemas["Root"].Value.Properties["use"]
	require.True(t, strings.HasPrefix(use.Ref, "#/components/schemas/"))
	targetName := strings.TrimPrefix(use.Ref, "#/components/schemas/")
	require.Nil(t, doc.Components.Schemas[targetName].Value.Items,
		"internalization must componentize the referenced target, not the effective sibling wrapper")
	var marshaledUse map[string]any
	require.NoError(t, json.Unmarshal(mustMarshalJSON(t, use), &marshaledUse))
	items := marshaledUse["items"].(map[string]any)
	require.True(t, strings.HasPrefix(items["$ref"].(string), "#/components/schemas/"))
}

func TestInternalizeRefsUpdatesSchemaRefSiblingFieldsInOpenAPI30(t *testing.T) {
	dir := t.TempDir()
	rootPath := filepath.Join(dir, "root.yaml")
	externalPath := filepath.Join(dir, "external.yaml")

	require.NoError(t, os.WriteFile(rootPath, []byte(`
openapi: 3.0.3
info:
  title: Internalized OpenAPI 3.0 sibling refs
  version: "1.0"
paths: {}
components:
  schemas:
    Root:
      type: object
      properties:
        use:
          $ref: external.yaml#/Base
          items:
            $ref: external.yaml#/Item
`), 0o600))
	require.NoError(t, os.WriteFile(externalPath, []byte(`
Base:
  type: array
Item:
  type: string
`), 0o600))

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile(rootPath)
	require.NoError(t, err)

	use := doc.Components.Schemas["Root"].Value.Properties["use"]
	require.Nil(t, use.Value.Items, "an OpenAPI 3.0 sibling must not affect the resolved target")
	before, err := json.Marshal(use)
	require.NoError(t, err)
	require.Contains(t, string(before), `external.yaml#/Item`)

	doc.InternalizeRefs(t.Context(), nil)
	data, err := doc.MarshalJSON()
	require.NoError(t, err)
	require.NotContains(t, string(data), "external.yaml")
	require.Nil(t, use.Value.Items, "internalization must not change OpenAPI 3.0 sibling semantics")

	var marshaledUse map[string]any
	require.NoError(t, json.Unmarshal(mustMarshalJSON(t, use), &marshaledUse))
	require.True(t, strings.HasPrefix(marshaledUse["$ref"].(string), "#/components/schemas/"))
	items := marshaledUse["items"].(map[string]any)
	require.True(t, strings.HasPrefix(items["$ref"].(string), "#/components/schemas/"))

	roundTripLoader := openapi3.NewLoader()
	roundTripped, err := roundTripLoader.LoadFromData(data)
	require.NoError(t, err, "the internalized document must not require external access")
	roundTrippedUse := roundTripped.Components.Schemas["Root"].Value.Properties["use"]
	require.Nil(t, roundTrippedUse.Value.Items)
}

func TestInternalizeRefsUpdatesSiblingFieldsWithFalseTarget(t *testing.T) {
	dir := t.TempDir()
	rootPath := filepath.Join(dir, "root.yaml")
	externalPath := filepath.Join(dir, "external.yaml")

	require.NoError(t, os.WriteFile(rootPath, []byte(`
openapi: 3.1.0
info:
  title: Internalized false-target sibling refs
  version: "1.0"
paths: {}
components:
  schemas:
    RejectEverything: false
    Root:
      type: object
      properties:
        use:
          $ref: '#/components/schemas/RejectEverything'
          items:
            $ref: external.yaml#/Item
`), 0o600))
	require.NoError(t, os.WriteFile(externalPath, []byte(`
Item:
  type: string
`), 0o600))

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile(rootPath)
	require.NoError(t, err)

	use := doc.Components.Schemas["Root"].Value.Properties["use"]
	require.NotNil(t, use.Value.Items)
	require.Equal(t, "external.yaml#/Item", use.Value.Items.Ref)
	require.Error(t, use.Value.VisitJSON([]any{"value"}, openapi3.EnableJSONSchema2020()))

	doc.InternalizeRefs(t.Context(), nil)
	data, err := doc.MarshalJSON()
	require.NoError(t, err)
	require.NotContains(t, string(data), "external.yaml")
	require.True(t, strings.HasPrefix(use.Value.Items.Ref, "#/components/schemas/"))
	require.Error(t, use.Value.VisitJSON([]any{"value"}, openapi3.EnableJSONSchema2020()))
}

func mustMarshalJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	return data
}
