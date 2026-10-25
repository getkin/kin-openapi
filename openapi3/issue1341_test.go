package openapi3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue1341(t *testing.T) {
	t.Parallel()

	// VisitJSON receives a subtree rather than the whole document, so every
	// link of the reference chain has to be copied into $defs. Reproducing the
	// report takes several iterations: the rewrite descends into the entries
	// its map iteration happens to reach.
	const spec = `
openapi: 3.1.0
info: {title: t, version: "1"}
paths: {}
components:
  schemas:
    Root:
      type: object
      properties:
        version: {type: integer, const: 1}
        child: {$ref: '#/components/schemas/A'}
    A:
      type: object
      properties:
        child: {$ref: '#/components/schemas/B'}
    B:
      type: object
      properties:
        child: {$ref: '#/components/schemas/C'}
    C:
      type: object
      properties:
        child: {$ref: '#/components/schemas/D'}
    D:
      type: string
`
	opt := openapi3.EnableJSONSchema2020()

	for range 50 {
		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromData([]byte(spec))
		require.NoError(t, err)
		require.NoError(t, doc.Validate(context.Background()))

		err = doc.Components.Schemas["Root"].Value.VisitJSON(map[string]any{"version": float64(1)}, opt)
		require.NoError(t, err)
	}
}
