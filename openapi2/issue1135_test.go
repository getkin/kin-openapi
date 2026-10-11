package openapi2_test

import (
	"testing"

	yaml "github.com/oasdiff/yaml3"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
)

// Swagger 2 writes "type" as a scalar, so a scalar has to decode into
// openapi3.Types straight from YAML just like it does from JSON.
func TestIssue1135(t *testing.T) {
	t.Parallel()

	var doc openapi2.T
	err := yaml.Unmarshal([]byte(`
swagger: "2.0"
info: {title: MyAPI, version: "0.1"}
paths:
  /pets:
    get:
      operationId: findPets
      parameters:
        - name: tags
          in: query
          required: false
          type: array
          collectionFormat: csv
          items:
            type: string
        - name: limit
          in: query
          type: integer
      responses:
        "200": {description: OK}
`), &doc)
	require.NoError(t, err)

	item := doc.Paths["/pets"].Get.Parameters[0]
	require.True(t, item.Type.Is("array"))

	limit := doc.Paths["/pets"].Get.Parameters[1]
	require.True(t, limit.Type.Is("integer"))
}

// The scalar form must keep round-tripping back out as a scalar.
func TestIssue1135MarshalYAML(t *testing.T) {
	t.Parallel()

	types := &openapi3.Types{"string"}
	data, err := yaml.Marshal(types)
	require.NoError(t, err)
	require.Equal(t, "string\n", string(data))

	var back openapi3.Types
	require.NoError(t, yaml.Unmarshal(data, &back))
	require.Equal(t, openapi3.Types{"string"}, back)
}

// A sequence must still decode as a sequence.
func TestIssue1135Sequence(t *testing.T) {
	t.Parallel()

	var types openapi3.Types
	require.NoError(t, yaml.Unmarshal([]byte("[string, 'null']\n"), &types))
	require.Equal(t, openapi3.Types{"string", "null"}, types)
}
