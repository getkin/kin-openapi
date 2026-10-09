package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestGHS_mh7x_f8wq_4jhx(t *testing.T) {
	t.Parallel()

	spec := `
openapi: 3.0.3
info: {title: t, version: "1.0.0"}
paths: {}
components:
  schemas:
    A:
      $ref: '#/components/schemas/A/additionalProperties'
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.EqualError(t, err, `invalid components: schema "A": found unresolved ref: "#/components/schemas/A/additionalProperties"`)
}
