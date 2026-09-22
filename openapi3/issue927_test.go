package openapi3_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue927(t *testing.T) {
	spec := `
openapi: '3.0'
info:
  title: title
  version: 0.0.0
paths: {}
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

	for _, openapi := range []string{"3.0", "3.1"} {
		t.Run(openapi, func(t *testing.T) {
			t.Parallel()
			spec := strings.ReplaceAll(spec, "3.0", openapi)

			sl := openapi3.NewLoader()
			doc, err := sl.LoadFromData([]byte(spec))
			require.NoError(t, err)

			err = doc.Validate(t.Context())
			require.NoError(t, err)

			require.False(t, doc.Components.Schemas["String"].Value.Nullable)
			require.True(t, doc.Components.Schemas["NullableString"].Value.Nullable)
			require.Equal(t, openapi != "3.0", doc.Components.Schemas["NullableRef"].Value.Nullable)
		})
	}
}
