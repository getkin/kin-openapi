package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// Two $refs to the same schema, each with its own sibling const, must not
// share the first one's sibling under the JSON Schema 2020-12 validator.
func TestIssue1381(t *testing.T) {
	t.Parallel()

	spec := `
openapi: 3.1.0
info: {title: Repro, version: "1"}
paths: {}
components:
  schemas:
    Kind:
      type: string
      enum: [a, b]
    TwoProps:
      type: object
      properties:
        x: {$ref: '#/components/schemas/Kind', const: a}
        y: {$ref: '#/components/schemas/Kind', const: b}
`
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(loader.Context))
	schema := doc.Components.Schemas["TwoProps"].Value

	opt := openapi3.EnableJSONSchema2020()
	require.NoError(t, schema.VisitJSON(map[string]any{"x": "a", "y": "b"}, opt))
	require.Error(t, schema.VisitJSON(map[string]any{"x": "a", "y": "a"}, opt))
	require.Error(t, schema.VisitJSON(map[string]any{"x": "b", "y": "b"}, opt))
	require.Error(t, schema.VisitJSON(map[string]any{"x": "a", "y": "c"}, opt))
}
