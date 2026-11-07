package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestSchemaRefSiblingConstraintsStayAtTheirUseSites(t *testing.T) {
	t.Parallel()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.1.0
info: {title: independent reference constraints, version: "1"}
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
      required: [x, y]
`))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(loader.Context))

	schema := doc.Components.Schemas["TwoProps"].Value
	require.NoError(t, schema.VisitJSON(map[string]any{"x": "a", "y": "b"}, openapi3.EnableJSONSchema2020()))
	require.Error(t, schema.VisitJSON(map[string]any{"x": "b", "y": "b"}, openapi3.EnableJSONSchema2020()))
	require.Error(t, schema.VisitJSON(map[string]any{"x": "a", "y": "a"}, openapi3.EnableJSONSchema2020()))

	base := doc.Components.Schemas["Kind"].Value
	for _, value := range []string{"a", "b"} {
		require.NoError(t, base.VisitJSON(value, openapi3.EnableJSONSchema2020()))
	}
}
