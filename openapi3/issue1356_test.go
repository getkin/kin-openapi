package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue1356(t *testing.T) {
	t.Parallel()

	// Author.Book is first reached while Book is still being resolved (through
	// Book.Author), so its value is filled in when the cycle unwinds. That has
	// to keep the readOnly sibling, the same as the other side of the cycle.
	const spec = `
openapi: 3.1.0
info: {title: t, version: "1"}
paths: {}
components:
  schemas:
    Author:
      type: object
      properties:
        Book:
          $ref: '#/components/schemas/Book'
          readOnly: true
    Book:
      type: object
      properties:
        Author:
          $ref: '#/components/schemas/Author'
          readOnly: true
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	author := doc.Components.Schemas["Author"].Value
	book := doc.Components.Schemas["Book"].Value

	require.True(t, author.Properties["Book"].Value.ReadOnly, "Author.Book")
	require.True(t, book.Properties["Author"].Value.ReadOnly, "Book.Author")

	// The siblings apply to the referencing property only, not to the shared components.
	require.False(t, author.ReadOnly)
	require.False(t, book.ReadOnly)
}
