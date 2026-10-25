package openapi2conv_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

// A media type need not carry a schema.
func TestFromV3_MediaTypeWithoutSchema(t *testing.T) {
	var doc3 openapi3.T
	require.NoError(t, json.Unmarshal([]byte(`{
		"openapi": "3.0.0",
		"info": {"title": "t", "version": "1"},
		"paths": {"/a": {"get": {"responses": {"200": {
			"description": "ok",
			"content": {"application/json": {}}
		}}}}}
	}`), &doc3))

	doc2, err := openapi2conv.FromV3(&doc3)
	require.NoError(t, err)
	require.Nil(t, doc2.Paths["/a"].Get.Responses.Value("200").Schema)
}

// Converting a nullable schema either way leaves the source document as it
// was: converting it again gives the same result.
func TestSchemaConversion_LeavesSourceUnchanged(t *testing.T) {
	var doc2 openapi2.T
	require.NoError(t, json.Unmarshal([]byte(`{
		"swagger": "2.0",
		"info": {"title": "t", "version": "1"},
		"paths": {},
		"definitions": {"A": {"type": "string", "x-nullable": true, "x-other": 1}}
	}`), &doc2))

	doc3, err := openapi2conv.ToV3(&doc2)
	require.NoError(t, err)
	require.Equal(t, true, doc2.Definitions["A"].Value.Extensions["x-nullable"])
	require.True(t, doc3.Components.Schemas["A"].Value.Nullable)

	back, err := openapi2conv.FromV3(doc3)
	require.NoError(t, err)
	require.Equal(t, true, back.Definitions["A"].Value.Extensions["x-nullable"])
	require.True(t, doc3.Components.Schemas["A"].Value.Nullable)
	require.NotContains(t, doc3.Components.Schemas["A"].Value.Extensions, "x-nullable")

	again, err := openapi2conv.FromV3(doc3)
	require.NoError(t, err)
	require.Equal(t, back.Definitions, again.Definitions)
}
