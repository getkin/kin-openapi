package openapi3_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestIssue341(t *testing.T) {
	sl := openapi3.NewLoader()
	sl.IsExternalRefsAllowed = true
	doc, err := sl.LoadFromFile("testdata/main.yaml")
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)

	err = sl.ResolveRefsIn(doc, nil)
	require.NoError(t, err)

	bs, err := doc.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, `{
	"info": {
		"title": "test file",
		"version": "n/a"
	},
	"openapi": "3.0.0",
	"paths": {
		"/testpath": {
			"$ref": "testpath.yaml#/paths/~1testpath"
		}
	}
}`, string(bs))

	require.Equal(t, &openapi3.Types{"string"}, doc.
		Paths.Value("/testpath").
		Get.
		Responses.Value("200").Value.
		Content["application/json"].
		Schema.Value.
		Type)

	doc.InternalizeRefs(t.Context(), nil)
	bs, err = doc.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, `{
		"components": {
		  "responses": {
			"testpath_testpath_200_response": {
			  "content": {
				"application/json": {
				  "schema": {
					"type": "string"
				  }
				}
			  },
			  "description": "a custom response"
			}
		  }
		},
		"info": {
		  "title": "test file",
		  "version": "n/a"
		},
		"openapi": "3.0.0",
		"paths": {
		  "/testpath": {
			"get": {
			  "responses": {
				"200": {
				  "$ref": "#/components/responses/testpath_testpath_200_response"
				}
			  }
			}
		  }
		}
	  }`, string(bs))
}
