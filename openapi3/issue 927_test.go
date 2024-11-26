package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue927(t *testing.T) {
	spec := `
openapi: 3.0.0
components:
  schemas:
    NullableString:
      type: string
      nullable: true
    NullableRef:
      $ref: "#/components/schemas/String"
      nullable: true
    String:
      type: string
`

	sl := NewLoader()
	doc, err := sl.LoadFromData([]byte(spec))
	require.NoError(t, err)

	require.False(t, doc.Components.Schemas["String"].Value.Nullable)
	require.True(t, doc.Components.Schemas["NullableString"].Value.Nullable)
	require.True(t, doc.Components.Schemas["NullableRef"].Value.Nullable)
}
