package openapi2conv_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToV3GlobalProduces(t *testing.T) {
	t.Parallel()

	spec := []byte(`swagger: '2.0'
info:
  title: title
  version: '1'
produces:
  - application/xml
  - text/csv
paths:
  /inherits:
    get:
      responses:
        '200':
          description: OK
          schema:
            type: string
  /overrides:
    get:
      produces:
        - application/pdf
      responses:
        '200':
          description: OK
          schema:
            type: string
  /clears:
    get:
      produces: []
      responses:
        '200':
          description: OK
          schema:
            type: string
`)

	doc3, err := v2v3YAML(spec)
	require.NoError(t, err)
	require.NoError(t, doc3.Validate(t.Context()))

	for path, mediaTypes := range map[string][]string{
		"/inherits":  {"application/xml", "text/csv"},
		"/overrides": {"application/pdf"},
		// An empty list clears the global one, leaving the default.
		"/clears": {"application/json"},
	} {
		content := doc3.Paths.Value(path).Get.Responses.Value("200").Value.Content
		require.ElementsMatch(t, mediaTypes, slices.Collect(maps.Keys(content)), path)
	}
}
