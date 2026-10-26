package openapi3_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// A $ref naming a whole document and a $ref with a fragment take different
// paths through the loader, and each applies IsExternalRefsAllowed on its own.
func TestDisallowedExternalRefError(t *testing.T) {
	t.Parallel()

	for _, ref := range []string{"other.yaml", "other.yaml#/components/schemas/Thing"} {
		t.Run(ref, func(t *testing.T) {
			t.Parallel()

			spec := fmt.Appendf(nil, `
openapi: 3.0.0
info:
  title: Test
  version: '1.0'
paths: {}
components:
  schemas:
    Thing:
      $ref: %q
`, ref)

			loader := openapi3.NewLoader()
			_, err := loader.LoadFromData(spec)
			require.Error(t, err)

			var refused *openapi3.DisallowedExternalRefError
			require.ErrorAs(t, err, &refused)
			require.Equal(t, ref, refused.Ref)
		})
	}
}
