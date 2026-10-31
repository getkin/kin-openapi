package openapi2conv_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

// A binary schema outside formData used to come back from FromV3 as an
// unnamed formData file, so a header became a form field and a round trip
// lost the parameter altogether.
func TestBinaryParameterOutsideFormData(t *testing.T) {
	t.Parallel()

	t.Run("v3 header and response header", func(t *testing.T) {
		t.Parallel()

		spec := []byte(`{
  "openapi": "3.0.3",
  "info": {"title": "t", "version": "1"},
  "paths": {"/f": {"get": {
    "parameters": [{"in": "header", "name": "X-Blob", "required": true, "schema": {"type": "string", "format": "binary"}}],
    "responses": {"200": {
      "description": "ok",
      "headers": {"X-Blob": {"schema": {"type": "string", "format": "binary"}}}
    }}
  }}}
}`)
		loader := openapi3.NewLoader()
		doc3, err := loader.LoadFromData(spec)
		require.NoError(t, err)

		doc2, err := openapi2conv.FromV3(doc3)
		require.NoError(t, err)
		op := doc2.Paths["/f"].Get
		require.Len(t, op.Parameters, 1)
		require.Equal(t, &openapi2.Parameter{
			In:       "header",
			Name:     "X-Blob",
			Required: true,
			Type:     &openapi3.Types{"string"},
			Format:   "binary",
		}, op.Parameters[0])
		header := op.Responses.Value("200").Headers["X-Blob"]
		require.NotNil(t, header)
		require.Equal(t, &openapi3.Types{"string"}, header.Type)
		require.Equal(t, "binary", header.Format)

		back3, err := openapi2conv.ToV3(doc2)
		require.NoError(t, err)
		param := back3.Paths.Value("/f").Get.Parameters[0].Value
		require.Equal(t, "header", param.In)
		require.Equal(t, "X-Blob", param.Name)
		require.True(t, param.Required)
		require.Equal(t, "binary", param.Schema.Value.Format)
		require.Nil(t, back3.Paths.Value("/f").Get.RequestBody)
	})

	t.Run("v2 file outside formData round-trips", func(t *testing.T) {
		t.Parallel()

		spec := []byte(`{"swagger":"2.0","paths":{"/f":{"post":{
  "consumes":["multipart/form-data"],
  "parameters":[{"in":"firmData","name":"file","type":"file"}],
  "responses":{"200":{"description":"ok"}}
}}}}`)
		var doc2 openapi2.T
		require.NoError(t, json.Unmarshal(spec, &doc2))

		roundTrip := func(doc *openapi3.T) []byte {
			back2, err := openapi2conv.FromV3(doc)
			require.NoError(t, err)
			next3, err := openapi2conv.ToV3(back2)
			require.NoError(t, err)
			data, err := json.Marshal(next3)
			require.NoError(t, err)
			return data
		}

		doc3, err := openapi2conv.ToV3(&doc2)
		require.NoError(t, err)
		first, err := json.Marshal(doc3)
		require.NoError(t, err)
		once := roundTrip(doc3)
		require.JSONEq(t, string(first), string(once))
	})
}
