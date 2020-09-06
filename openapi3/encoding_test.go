package openapi3

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodingJSON(t *testing.T) {
	t.Log("Marshal *openapi3.Encoding to JSON")
	data, err := json.Marshal(encoding())
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Unmarshal *openapi3.Encoding from JSON")
	docA := &Encoding{}
	err = json.Unmarshal(encodingJSON, &docA)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Validate *openapi3.Encoding")
	err = docA.Validate(context.Background())
	require.NoError(t, err)

	t.Log("Ensure representations match")
	dataA, err := json.Marshal(docA)
	require.NoError(t, err)
	require.JSONEq(t, string(data), string(encodingJSON))
	require.JSONEq(t, string(data), string(dataA))
}

var encodingJSON = []byte(`
{
  "contentType": "application/json",
  "headers": {
    "someHeader": {}
  },
  "style": "form",
  "explode": true,
  "allowReserved": true
}
`)

func encoding() *Encoding {
	explode := true
	return &Encoding{
		ContentType: "application/json",
		Headers: map[string]*HeaderRef{
			"someHeader": {
				Value: &Header{},
			},
		},
		Style:         "form",
		Explode:       &explode,
		AllowReserved: true,
	}
}

func TestEncodingSerializationMethod(t *testing.T) {
	boolPtr := func(b bool) *bool { return &b }
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
			enc:  &Encoding{Explode: boolPtr(true)},
			want: &SerializationMethod{Style: SerializationForm, Explode: true},
		},
		{
			name: "encoding with no explode",
			enc:  &Encoding{Explode: boolPtr(false)},
			want: &SerializationMethod{Style: SerializationForm, Explode: false},
		},
		{
			name: "encoding with style and explode ",
			enc:  &Encoding{Style: SerializationSpaceDelimited, Explode: boolPtr(false)},
			want: &SerializationMethod{Style: SerializationSpaceDelimited, Explode: false},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.enc.SerializationMethod()
			require.True(t, reflect.DeepEqual(got, tc.want), "got %#v, want %#v", got, tc.want)
		})
	}
}
