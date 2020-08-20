package openapi3_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
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
		loader := openapi3.NewSwaggerLoader()
		doc, err := loader.LoadSwaggerFromData(spec)
		require.NoError(t, err)
		got := doc.Paths["/path1"].Get.Responses["200"].Value.Description
		expected := ""
		require.Equal(t, &expected, got)
		t.Log("Empty description provided: valid spec")
		err = doc.Validate(loader.Context)
		require.NoError(t, err)
	}

	{
		spec := []byte(strings.Replace(spec, `"description": ""`, `"description": "My response"`, 1))
		loader := openapi3.NewSwaggerLoader()
		doc, err := loader.LoadSwaggerFromData(spec)
		require.NoError(t, err)
		got := doc.Paths["/path1"].Get.Responses["200"].Value.Description
		expected := "My response"
		require.Equal(t, &expected, got)
		t.Log("Non-empty description provided: valid spec")
		err = doc.Validate(loader.Context)
		require.NoError(t, err)
	}

	noDescriptionIsInvalid := func(data []byte) *openapi3.Swagger {
		loader := openapi3.NewSwaggerLoader()
		doc, err := loader.LoadSwaggerFromData(data)
		require.NoError(t, err)
		got := doc.Paths["/path1"].Get.Responses["200"].Value.Description
		require.Nil(t, got)
		t.Log("No description provided: invalid spec")
		err = doc.Validate(loader.Context)
		require.Error(t, err)
		return doc
	}

	var docWithNoResponseDescription *openapi3.Swagger
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
