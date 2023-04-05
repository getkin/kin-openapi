package openapi3

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMediaTypeJSON(t *testing.T) {
	t.Log("Marshal *openapi3.MediaType to JSON")
	data, err := json.Marshal(mediaType())
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Unmarshal *openapi3.MediaType from JSON")
	docA := &MediaType{}
	err = json.Unmarshal(mediaTypeJSON, &docA)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Validate *openapi3.MediaType")
	err = docA.Validate(context.Background())
	require.NoError(t, err)

	t.Log("Ensure representations match")
	dataA, err := json.Marshal(docA)
	require.NoError(t, err)
	require.JSONEq(t, string(data), string(mediaTypeJSON))
	require.JSONEq(t, string(data), string(dataA))
}

var mediaTypeJSON = []byte(`
{
   "schema": {
      "description": "Some schema"
   },
   "encoding": {
      "someEncoding": {
         "contentType": "application/xml; charset=utf-8"
      }
   },
   "examples": {
      "someExample": {
         "value": {
            "name": "Some example"
         }
      }
   }
}
`)

func mediaType() *MediaType {
	example := map[string]string{"name": "Some example"}
	return &MediaType{
		Schema: &SchemaRef{
			Value: &Schema{
				Description: "Some schema",
			},
		},
		Encoding: map[string]*Encoding{
			"someEncoding": {
				ContentType: "application/xml; charset=utf-8",
			},
		},
		Examples: map[string]*ExampleRef{
			"someExample": {
				Value: NewExample(example),
			},
		},
	}
}
