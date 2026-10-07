package openapi3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue1288(t *testing.T) {
	spec := []byte(`
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
        parent: {$ref: "#/components/schemas/Node"}
        labels:
          type: object
          additionalProperties: {$ref: "#/components/schemas/Node"}
`[1:])

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)
	err = doc.Validate(context.Background())
	require.NoError(t, err)

	node := doc.Components.Schemas["Node"].Value

	err = node.VisitJSON(map[string]any{})
	require.ErrorContains(t, err, `property "name" is missing`)

	err = node.VisitJSON(map[string]any{
		"name":     "a",
		"children": []any{map[string]any{"name": "b"}},
		"parent":   map[string]any{"name": "c"},
		"labels":   map[string]any{"x": map[string]any{"name": "d"}},
	})
	require.NoError(t, err)

	for name, value := range map[string]map[string]any{
		"items":                {"name": "a", "children": []any{map[string]any{}}},
		"nested items":         {"name": "a", "children": []any{map[string]any{"name": "b", "children": []any{map[string]any{}}}}},
		"properties":           {"name": "a", "parent": map[string]any{}},
		"additionalProperties": {"name": "a", "labels": map[string]any{"x": map[string]any{}}},
	} {
		t.Run(name, func(t *testing.T) {
			err := node.VisitJSON(value)
			require.ErrorContains(t, err, `property "name" is missing`)
		})
	}
}
