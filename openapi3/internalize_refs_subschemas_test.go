package openapi3_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// InternalizeRefs reaches an external ref under every keyword that holds a
// subschema, not only the ones 3.0 has.
func TestInternalizeRefsSubschemaKeywords(t *testing.T) {
	other := []byte(`{"components":{"schemas":{"Ext":{"type":"string"}}}}`)
	const ext = `{"$ref":"ext.json#/components/schemas/Ext"}`

	for keyword, value := range map[string]string{
		"prefixItems":           `[` + ext + `]`,
		"contains":              ext,
		"propertyNames":         ext,
		"if":                    ext,
		"then":                  ext,
		"else":                  ext,
		"contentSchema":         ext,
		"unevaluatedItems":      ext,
		"unevaluatedProperties": ext,
		"patternProperties":     `{"^a":` + ext + `}`,
		"dependentSchemas":      `{"a":` + ext + `}`,
		"$defs":                 `{"D":` + ext + `}`,
	} {
		t.Run(keyword, func(t *testing.T) {
			root := []byte(`{"openapi":"3.1.0","info":{"title":"t","version":"1"},"paths":{},` +
				`"components":{"schemas":{"S":{"` + keyword + `":` + value + `}}}}`)

			loader := newTwoDocLoader(root, other)
			doc, err := loader.LoadFromDataWithPath(root, fuzzRootURL)
			require.NoError(t, err)

			doc.InternalizeRefs(t.Context(), nil)

			data, err := json.Marshal(doc)
			require.NoError(t, err)
			require.NotContains(t, string(data), "ext.json")
			require.Len(t, doc.Components.Schemas, 2)
		})
	}
}
