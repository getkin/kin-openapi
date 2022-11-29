package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue618(t *testing.T) {
	spec := `
openapi: 3.0.0
info:
  title: foo
  version: 0.0.0
paths:
  /foo:
    get:
      responses:
        '200':
          description: Some description value text
          content:
            application/json:
              schema:
                $ref: ./testdata/schema618.yml#/components/schemas/JournalEntry
`[1:]

	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	ctx := loader.Context

	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	doc.InternalizeRefs(ctx, nil)

	require.Contains(t, doc.Components.Schemas, "JournalEntry")
	require.Contains(t, doc.Components.Schemas, "Record")
	require.Contains(t, doc.Components.Schemas, "Account")
}
