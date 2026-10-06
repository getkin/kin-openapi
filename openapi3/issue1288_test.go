package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue1288(t *testing.T) {
	const spec = `
openapi: 3.0.3
info: {title: t, version: "1"}
paths: {}
components:
  schemas:
    Node:
      type: object
      required: [name]
      properties:
        name: {type: string}
        children:
          type: array
          items: {$ref: "#/components/schemas/Node"}
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)
	node := doc.Components.Schemas["Node"].Value

	for _, tc := range []struct {
		value   map[string]any
		errPath string
	}{
		{map[string]any{}, `"/name"`},
		{map[string]any{"name": "a", "children": []any{map[string]any{}}}, `"/children/0/name"`},
		{map[string]any{"name": "a", "children": []any{map[string]any{"name": "b"}, map[string]any{}}}, `"/children/1/name"`},
		{map[string]any{"name": "a", "children": []any{map[string]any{"name": "b", "children": []any{map[string]any{}}}}}, `"/children/0/children/0/name"`},
	} {
		err := node.VisitJSON(tc.value)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Error at "+tc.errPath)
	}

	err = node.VisitJSON(map[string]any{"name": "a", "children": []any{map[string]any{"name": "b", "children": []any{}}}})
	require.NoError(t, err)
}
