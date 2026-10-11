package openapi3

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodingJSON(t *testing.T) {
	t.Parallel()

	t.Log("Marshal *openapi3.Encoding to JSON")
	data, err := json.Marshal(encoding())
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Unmarshal *openapi3.Encoding from JSON")
	enc := &Encoding{}
	err = json.Unmarshal(encodingJSON, &enc)
	require.NoError(t, err)
	require.NotEmpty(t, enc)

	t.Log("Validate *openapi3.Encoding")
	err = enc.Validate(t.Context())
	require.NoError(t, err)

	t.Log("Ensure representations match")
	dataA, err := json.Marshal(enc)
	require.NoError(t, err)
	require.JSONEq(t, string(data), string(encodingJSON))
	require.JSONEq(t, string(data), string(dataA))
}

var encodingJSON = []byte(`
{
  "contentType": "application/json",
  "headers": {
    "someHeader": {
      "schema": {"type": "string"}
    }
  },
  "style": "form",
  "explode": true,
  "allowReserved": true
}
`)

func encoding() *Encoding {
	return &Encoding{
		ContentType: "application/json",
		Headers: map[string]*HeaderRef{
			"someHeader": {
				Value: &Header{Parameter: Parameter{Schema: NewStringSchema().NewRef()}},
			},
		},
		Style:         "form",
		Explode:       Ptr(true),
		AllowReserved: true,
	}
}

func TestEncodingSerializationMethod(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		enc  *Encoding
		want *SerializationMethod
	}{
		{
			name: "default",
			want: &SerializationMethod{Style: SerializationForm, Explode: true},
		},
		{
			name: "encoding with style",
			enc:  &Encoding{Style: SerializationSpaceDelimited},
			want: &SerializationMethod{Style: SerializationSpaceDelimited, Explode: true},
		},
		{
			name: "encoding with explode",
			enc:  &Encoding{Explode: Ptr(true)},
			want: &SerializationMethod{Style: SerializationForm, Explode: true},
		},
		{
			name: "encoding with no explode",
			enc:  &Encoding{Explode: Ptr(false)},
			want: &SerializationMethod{Style: SerializationForm, Explode: false},
		},
		{
			name: "encoding with style and explode ",
			enc:  &Encoding{Style: SerializationSpaceDelimited, Explode: Ptr(false)},
			want: &SerializationMethod{Style: SerializationSpaceDelimited, Explode: false},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.enc.SerializationMethod()
			require.EqualValues(t, tc.want, got, "got %#v, want %#v", got, tc.want)
		})
	}
}

func TestEncodingValidateHeaders(t *testing.T) {
	t.Parallel()

	t.Run("invalid header name", func(t *testing.T) {
		enc := &Encoding{Headers: Headers{"bad header": {Value: &Header{}}}}
		err := enc.Validate(t.Context())
		require.ErrorContains(t, err, `identifier "bad header" is not supported`)
	})

	t.Run("invalid header", func(t *testing.T) {
		enc := &Encoding{Headers: Headers{"X-Rate": {Value: &Header{Parameter: Parameter{
			Schema:  NewStringSchema().NewRef(),
			Content: NewContentWithJSONSchema(NewStringSchema()),
		}}}}}
		err := enc.Validate(t.Context())
		require.Error(t, err)
	})
}

func TestEncodingValidateContentType(t *testing.T) {
	for _, tc := range []struct {
		contentType string
		valid       bool
	}{
		{contentType: "", valid: true},
		{contentType: "image/png", valid: true},
		{contentType: "image/*", valid: true},
		{contentType: "*/*", valid: true},
		{contentType: "application/xml; charset=utf-8", valid: true},
		{contentType: "image/png, image/jpeg", valid: true},
		{contentType: "application/xml,application/json,application/*", valid: true},
		{contentType: "png"},
		{contentType: "image/"},
		{contentType: "/png"},
		{contentType: "*/png"},
		{contentType: "image/png/x"},
		{contentType: "image /png"},
		{contentType: "image/png,"},
		{contentType: ",image/png"},
		{contentType: ",image/png,"},
		{contentType: "image/png,,image/jpeg"},
	} {
		t.Run(tc.contentType, func(t *testing.T) {
			encoding := Encoding{ContentType: tc.contentType}
			err := encoding.Validate(t.Context())
			if tc.valid {
				require.NoError(t, err)
				return
			}
			var e *InvalidEncodingContentTypeError
			require.ErrorAs(t, err, &e)
			require.Equal(t, tc.contentType, e.ContentType)
			require.Equal(t, "encoding-content-type-invalid", e.Code())
		})
	}
}
