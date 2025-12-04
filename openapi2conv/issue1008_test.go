package openapi2conv

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssue1008(t *testing.T) {
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
      - name: zebra
        in: formData
        description: stripes
        required: true
        type: string
      - name: alpaca
        in: formData
        description: chewy
        required: true
        type: string
      - name: bee
        in: formData
        description: buzz
        required: true
        type: string
      responses:
        '200':
          description: OK
`)

	v3, err := v2v3YAML(v2)
	require.NoError(t, err)

	err = v3.Validate(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"alpaca", "bee", "zebra"}, v3.Paths.Value("/ping").Post.RequestBody.Value.Content.Get("multipart/form-data").Schema.Value.Required)
}
