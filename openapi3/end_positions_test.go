package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// A block ending in a multi-line scalar reaches the end of that scalar. The
// scalar is one node whose Line is where it starts, so the lines below it hang
// off nothing and a subtree walk alone stops short.
func TestEndPositions_BlockScalarAtEndOfBlock(t *testing.T) {
	spec := `
openapi: 3.0.0
info:
  title: t
  version: "1"
paths:
  /a:
    get:
      responses:
        "200":
          description: |
            first line
            second line

            after a blank line
  /b:
    get:
      responses:
        "200":
          description: ok
`[1:]

	loader := NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	// /a spans to the last line of the description, line 14, not to line 10
	// where the scalar starts.
	a := doc.Paths.Value("/a").Origin
	require.NotNil(t, a)
	require.NotNil(t, a.Key)
	require.Equal(t, 6, a.Key.Line)
	require.Equal(t, 14, a.Key.EndLine, "block must reach the end of the trailing block scalar")

	// The blank line inside the scalar is part of it; the one after is not.
	get := doc.Paths.Value("/a").Get.Origin
	require.NotNil(t, get)
	require.Equal(t, 14, get.Key.EndLine)

	// The following sibling is unaffected.
	b := doc.Paths.Value("/b").Origin
	require.NotNil(t, b)
	require.Equal(t, 15, b.Key.Line)
	require.Equal(t, 19, b.Key.EndLine)
}
