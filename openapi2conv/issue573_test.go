package openapi2conv

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue573(t *testing.T) {
	spec := []byte(`paths:
  /ping:
    get:
      produces:
        - application/toml
        - application/xml
      responses:
        200:
          schema:
            type: object
            properties:
              username:
                type: string
                description: The user name.
    post:
      responses:
        200:
          schema:
            type: object
            properties:
              username:
                type: string
                description: The user name.`)

	v3, err := v2v3YAML(spec)
	require.NoError(t, err)

	// Make sure the response content appears for each mime-type originally
	// appeared in "produces".
	pingGetContent := v3.Paths.Value("/ping").Get.Responses.Value("200").Value.Content
	require.Len(t, pingGetContent, 2)
	require.Contains(t, pingGetContent, "application/toml")
	require.Contains(t, pingGetContent, "application/xml")

	// Is "produces" is not explicitly specified, default to "application/json".
	pingPostContent := v3.Paths.Value("/ping").Post.Responses.Value("200").Value.Content
	require.Len(t, pingPostContent, 1)
	require.Contains(t, pingPostContent, "application/json")
}
