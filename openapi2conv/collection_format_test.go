package openapi2conv_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

func TestQueryCollectionFormat(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		collectionFormat string
		style            string
		explode          *bool
	}{
		{collectionFormat: "", style: "", explode: nil},
		{collectionFormat: "csv", style: openapi3.SerializationForm, explode: openapi3.Ptr(false)},
		{collectionFormat: "ssv", style: openapi3.SerializationSpaceDelimited, explode: openapi3.Ptr(false)},
		{collectionFormat: "pipes", style: openapi3.SerializationPipeDelimited, explode: openapi3.Ptr(false)},
		{collectionFormat: "multi", style: openapi3.SerializationForm, explode: openapi3.Ptr(true)},
		{collectionFormat: "tsv", style: "", explode: nil},
	} {
		t.Run("collectionFormat="+tc.collectionFormat, func(t *testing.T) {
			t.Parallel()
			spec := fmt.Sprintf(`{
	"swagger": "2.0",
	"info": {"title": "Example", "version": "1"},
	"paths": {
		"/pets": {
			"get": {
				"parameters": [{
					"in": "query",
					"name": "ids",
					"type": "array",
					"items": {"type": "integer"},
					"collectionFormat": %q
				}],
				"responses": {"200": {"description": "OK"}}
			}
		}
	}
}`, tc.collectionFormat)

			var doc2 openapi2.T
			err := json.Unmarshal([]byte(spec), &doc2)
			require.NoError(t, err)

			doc3, err := openapi2conv.ToV3(&doc2)
			require.NoError(t, err)
			err = doc3.Validate(t.Context())
			require.NoError(t, err)

			parameter := doc3.Paths.Value("/pets").Get.Parameters[0].Value
			require.Equal(t, tc.style, parameter.Style)
			require.Equal(t, tc.explode, parameter.Explode)

			back, err := openapi2conv.FromV3(doc3)
			require.NoError(t, err)

			want := tc.collectionFormat
			if want == "tsv" {
				want = ""
			}
			require.Equal(t, want, back.Paths["/pets"].Get.Parameters[0].CollectionFormat)
		})
	}
}
