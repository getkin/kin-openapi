package openapi2conv

import (
	"testing"

	"github.com/invopop/yaml"
	"github.com/stretchr/testify/require"
)

func TestPR558(t *testing.T) {
	spec := `
swagger: '2.0'
info:
  version: 1.0.0
  title: title
paths:
  /test:
    get:
      deprecated: true
      parameters:
      - in: body
        schema:
          type: object
      responses:
        '200':
          description: description
`
	doc3, err := v2v3YAML([]byte(spec))
	require.NoError(t, err)
	require.NotEmpty(t, doc3.Paths.Value("/test").Get.Deprecated)
	_, err = yaml.Marshal(doc3)
	require.NoError(t, err)

	doc2, err := FromV3(doc3)
	require.NoError(t, err)
	require.NotEmpty(t, doc2.Paths["/test"].Get.Deprecated)
}
