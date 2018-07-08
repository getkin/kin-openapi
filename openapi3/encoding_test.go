package openapi3_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestEncodingJSON(t *testing.T) {
	t.Log("Marshal *openapi3.Encoding to JSON")
	data, err := json.Marshal(encoding())
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Unmarshal *openapi3.Encoding from JSON")
	docA := &openapi3.Encoding{}
	err = json.Unmarshal(encodingJSON, &docA)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Validate *openapi3.Encoding")
	err = docA.Validate(context.TODO())
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
  "style": "simple",
  "explode": true,
  "allowReserved": true
}
`)

func encoding() *openapi3.Encoding {
	return &openapi3.Encoding{
		ContentType: "application/json",
		Headers: map[string]*openapi3.HeaderRef{
			"someHeader": {
				Value: &openapi3.Header{},
			},
		},
		Style:         "simple",
		Explode:       true,
		AllowReserved: true,
	}
}
