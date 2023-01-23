package openapi3

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONSpecResponseDescriptionEmptiness(t *testing.T) {
	const spec = `
{
  "info": {
    "description": "A sample API to illustrate OpenAPI concepts",
    "title": "Sample API",
    "version": "1.0.0"
  },
  "openapi": "3.0.0",
  "paths": {
    "/path1": {
      "get": {
        "responses": {
          "200": {
            "description": ""
          }
        }
      }
    }
  }
}
`

	{
		spec := []byte(spec)
		loader := NewLoader()
		doc, err := loader.LoadFromData(spec)
		require.NoError(t, err)
		require.Equal(t, "", *doc.Paths.Value("/path1").Get.Responses.Value("200").Value.Description)
		t.Log("Empty description provided: valid spec")
		err = doc.Validate(loader.Context)
		require.NoError(t, err)
	}

	{
		spec := []byte(strings.Replace(spec, `"description": ""`, `"description": "My response"`, 1))
		loader := NewLoader()
		doc, err := loader.LoadFromData(spec)
		require.NoError(t, err)
		require.Equal(t, "My response", *doc.Paths.Value("/path1").Get.Responses.Value("200").Value.Description)
		t.Log("Non-empty description provided: valid spec")
		err = doc.Validate(loader.Context)
		require.NoError(t, err)
	}

	noDescriptionIsInvalid := func(data []byte) *T {
		loader := NewLoader()
		doc, err := loader.LoadFromData(data)
		require.NoError(t, err)
		require.Nil(t, doc.Paths.Value("/path1").Get.Responses.Value("200").Value.Description)
		t.Log("No description provided: invalid spec")
		err = doc.Validate(loader.Context)
		require.Error(t, err)
		return doc
	}

	var docWithNoResponseDescription *T
	{
		spec := []byte(strings.Replace(spec, `"description": ""`, ``, 1))
		docWithNoResponseDescription = noDescriptionIsInvalid(spec)
	}

	str, err := json.Marshal(docWithNoResponseDescription)
	require.NoError(t, err)
	require.NotEmpty(t, str)
	t.Log("Reserialization does not set description")
	_ = noDescriptionIsInvalid(str)
}
