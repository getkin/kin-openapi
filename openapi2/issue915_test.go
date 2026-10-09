package openapi2_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
)

func TestIssue915(t *testing.T) {
	t.Parallel()

	spec := []byte(`
{
  "swagger": "2.0",
  "info": {"title": "MyAPI", "version": "0.1"},
  "paths": {
    "/foo": {
      "get": {
        "responses": {
          "200": {"description": "OK"},
          "default": {"$ref": "#/responses/Error"},
          "x-string": "some extension",
          "x-object": {"a": 1}
        }
      }
    }
  },
  "responses": {
    "Error": {"description": "error"}
  }
}
`[1:])

	var doc openapi2.T
	err := json.Unmarshal(spec, &doc)
	require.NoError(t, err)

	responses := doc.Paths["/foo"].Get.Responses
	require.Equal(t, 2, responses.Len())
	require.Equal(t, "OK", responses.Value("200").Description)
	require.Equal(t, "#/responses/Error", responses.Value("default").Ref)
	require.Nil(t, responses.Value("x-string"))
	require.Equal(t, map[string]any{
		"x-string": "some extension",
		"x-object": map[string]any{"a": float64(1)},
	}, responses.Extensions)

	data, err := json.Marshal(doc)
	require.NoError(t, err)
	require.JSONEq(t, string(spec), string(data))
}
