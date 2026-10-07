package openapi2conv_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
)

func TestIssue915(t *testing.T) {
	spec := []byte(`
{
  "swagger": "2.0",
  "info": {"title": "MyAPI", "version": "0.1"},
  "paths": {
    "/foo": {
      "get": {
        "responses": {
          "200": {"description": "OK"},
          "x-responses-ext": "value"
        }
      }
    }
  }
}
`[1:])

	var doc2 openapi2.T
	err := json.Unmarshal(spec, &doc2)
	require.NoError(t, err)

	doc3, err := openapi2conv.ToV3(&doc2)
	require.NoError(t, err)
	err = doc3.Validate(t.Context())
	require.NoError(t, err)

	responses3 := doc3.Paths.Value("/foo").Get.Responses
	require.Equal(t, 1, responses3.Len())
	require.Equal(t, "value", responses3.Extensions["x-responses-ext"])

	doc2back, err := openapi2conv.FromV3(doc3)
	require.NoError(t, err)

	responses2 := doc2back.Paths["/foo"].Get.Responses
	require.Equal(t, 1, responses2.Len())
	require.Equal(t, "OK", responses2.Value("200").Description)
	require.Equal(t, "value", responses2.Extensions["x-responses-ext"])
}
