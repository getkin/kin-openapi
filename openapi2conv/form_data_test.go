package openapi2conv_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

const formDataSpecV2 = `{
	"swagger": "2.0",
	"info": {"title": "Example", "version": "1"},
	"paths": {
		"/upload": {
			"post": {
				"consumes": ["multipart/form-data"],
				"parameters": [
					{"name": "file", "in": "formData", "type": "file", "required": true},
					{"name": "taken", "in": "formData", "type": "string", "format": "date"},
					{"name": "note", "in": "formData", "type": "string"}
				],
				"responses": {"200": {"description": "OK"}}
			}
		}
	}
}`

func formDataParams(t *testing.T, doc2 *openapi2.T) map[string]*openapi2.Parameter {
	t.Helper()

	byName := map[string]*openapi2.Parameter{}
	for _, parameter := range doc2.Paths["/upload"].Post.Parameters {
		require.Equal(t, "formData", parameter.In)
		byName[parameter.Name] = parameter
	}
	return byName
}

// TestFormDataRoundTrip covers the fields of a form field that the conversion
// back from OpenAPI 3 used to drop: whether the field is required, which is
// recorded on the object holding the properties, and its format.
func TestFormDataRoundTrip(t *testing.T) {
	t.Parallel()

	var doc2 openapi2.T
	require.NoError(t, json.Unmarshal([]byte(formDataSpecV2), &doc2))

	doc3, err := openapi2conv.ToV3(&doc2)
	require.NoError(t, err)

	// The required field is named on the object, as OpenAPI 3 requires.
	body := doc3.Paths.Value("/upload").Post.RequestBody.Value
	require.Equal(t, []string{"file"}, body.Content["multipart/form-data"].Schema.Value.Required)

	back, err := openapi2conv.FromV3(doc3)
	require.NoError(t, err)
	params := formDataParams(t, back)
	require.Len(t, params, 3)

	require.True(t, params["file"].Required, "a required upload stays required")
	require.Equal(t, &openapi3.Types{"file"}, params["file"].Type)
	require.Empty(t, params["file"].Format, "type file spends the format naming the type")

	require.Equal(t, "date", params["taken"].Format, "a format on a form field survives")
	require.Equal(t, &openapi3.Types{"string"}, params["taken"].Type)
	require.False(t, params["taken"].Required)

	require.Empty(t, params["note"].Format)
	require.False(t, params["note"].Required)
}

// TestFormDataRequiredOnTheObject pins that required is read from the object
// holding the properties, which is where a document not written by this
// package records it.
func TestFormDataRequiredOnTheObject(t *testing.T) {
	t.Parallel()

	mediaType := &openapi3.MediaType{Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{
		Type:     &openapi3.Types{"object"},
		Required: []string{"file"},
		Properties: openapi3.Schemas{
			"file": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "binary"}},
			"note": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
		},
	}}}

	byName := map[string]*openapi2.Parameter{}
	for _, parameter := range openapi2conv.FromV3RequestBodyFormData(mediaType) {
		byName[parameter.Name] = parameter
	}
	require.True(t, byName["file"].Required)
	require.False(t, byName["note"].Required)
}

// TestFormDataNameMarkerStaysInternal pins that x-formData-name, which carries
// the field name into OpenAPI 3 where it has nowhere else to live, does not
// come back out in the converted document, and that converting leaves the
// schema it was read from as it was.
func TestFormDataNameMarkerStaysInternal(t *testing.T) {
	t.Parallel()

	extensions := map[string]any{"x-formData-name": "file", "x-mimetype": "text/plain", "unknown": true}
	schema := &openapi3.Schema{
		Type:       &openapi3.Types{"string"},
		Format:     "binary",
		Extensions: extensions,
	}
	mediaType := &openapi3.MediaType{Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{
		Type:       &openapi3.Types{"object"},
		Properties: openapi3.Schemas{"file": {Value: schema}},
	}}}

	params := openapi2conv.FromV3RequestBodyFormData(mediaType)
	require.Len(t, params, 1)
	require.Equal(t, map[string]any{"x-mimetype": "text/plain"}, params[0].Extensions)

	require.Equal(t, extensions, schema.Extensions, "the schema keeps its own extensions")
}
