package openapi2conv

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue847(t *testing.T) {
	v2 := []byte(`
swagger: '2.0'
info:
  version: '1.10'
  title: title
paths:
  "/ping":
    post:
      consumes:
      - multipart/form-data
      parameters:
      - name: file
        in: formData
        description: file
        required: true
        type: file
      responses:
        '200':
          description: OK
`)

	v3, err := v2v3YAML(v2)
	require.NoError(t, err)

	err = v3.Validate(context.Background())
	require.NoError(t, err)

	require.Equal(t, []string{"file"}, v3.Paths.Value("/ping").Post.RequestBody.Value.Content["multipart/form-data"].Schema.Value.Required)

	require.Nil(t, v3.Paths.Value("/ping").Post.RequestBody.Value.Content["multipart/form-data"].Schema.Value.Properties["file"].Value.Required)
}
