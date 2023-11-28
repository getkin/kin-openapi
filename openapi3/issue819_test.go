package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue819ResponsesGetPatternedFields(t *testing.T) {
	spec := `
openapi: "3.0.3"
info:
  title: 'My app'
  version: 1.0.0
  description: 'An API'

paths:
  /v1/operation:
    get:
      summary: Fetch something
      responses:
        2XX:
          description: Success
          content:
            application/json:
              schema:
                type: object
                description: An error response body.
`[1:]
	sl := NewLoader()
	doc, err := sl.LoadFromData([]byte(spec))
	require.NoError(t, err)
	err = doc.Validate(sl.Context)
	require.NoError(t, err)

	require.NotNil(t, doc.Paths.Value("/v1/operation").Get.Responses.Status(201))
	require.Nil(t, doc.Paths.Value("/v1/operation").Get.Responses.Status(404))
	require.Nil(t, doc.Paths.Value("/v1/operation").Get.Responses.Status(999))
}
