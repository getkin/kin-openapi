package openapi2conv_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

const v3ResponseHeaderRefs = `{
	"openapi": "3.0.3",
	"info": {"title": "Example", "version": "1"},
	"paths": {
		"/pets": {
			"get": {
				"responses": {
					"200": {
						"description": "OK",
						"headers": {"X-Rate-Limit": {"$ref": "#/components/headers/X-Rate-Limit"}}
					},
					"default": {"$ref": "#/components/responses/Error"}
				}
			}
		}
	},
	"components": {
		"headers": {
			"X-Rate-Limit": {"description": "Requests per hour", "schema": {"type": "integer"}}
		},
		"responses": {
			"Error": {
				"description": "Error",
				"headers": {"X-Rate-Limit": {"$ref": "#/components/headers/X-Rate-Limit"}}
			}
		}
	}
}`

func TestFromV3ResponseHeaderRefs(t *testing.T) {
	t.Parallel()

	loader := openapi3.NewLoader()
	doc3, err := loader.LoadFromData([]byte(v3ResponseHeaderRefs))
	require.NoError(t, err)
	require.NoError(t, doc3.Validate(loader.Context))

	doc2, err := openapi2conv.FromV3(doc3)
	require.NoError(t, err)

	header := `{"X-Rate-Limit": {"description": "Requests per hour", "type": "integer"}}`

	data, err := json.Marshal(doc2.Paths["/pets"].Get.Responses.Value("200").Headers)
	require.NoError(t, err)
	require.JSONEq(t, header, string(data))

	data, err = json.Marshal(doc2.Responses["Error"].Headers)
	require.NoError(t, err)
	require.JSONEq(t, header, string(data))
}

func TestFromV3UnresolvedResponseHeaderRef(t *testing.T) {
	t.Parallel()

	var doc3 openapi3.T
	require.NoError(t, json.Unmarshal([]byte(v3ResponseHeaderRefs), &doc3))

	_, err := openapi2conv.FromV3(&doc3)
	require.EqualError(t, err, `unresolved reference "#/components/headers/X-Rate-Limit" for header "X-Rate-Limit"`)
}
