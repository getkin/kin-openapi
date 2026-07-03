package openapi3

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMediaTypeJSON(t *testing.T) {
	t.Log("Marshal *openapi3.MediaType to JSON")
	data, err := json.Marshal(mediaType())
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Unmarshal *openapi3.MediaType from JSON")
	mt := &MediaType{}
	err = json.Unmarshal(mediaTypeJSON, &mt)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Validate *openapi3.MediaType")
	err = mt.Validate(t.Context())
	require.NoError(t, err)

	t.Log("Ensure representations match")
	dataA, err := json.Marshal(mt)
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

func TestMediaTypeItemSchemaJSON(t *testing.T) {
	data := []byte(`
{
   "itemSchema": {
      "type": "object",
      "properties": {
         "data": {
            "type": "string"
         }
      }
   }
}
`)

	mt := &MediaType{}
	err := json.Unmarshal(data, mt)
	require.NoError(t, err)
	require.NotNil(t, mt.ItemSchema)
	require.NotNil(t, mt.ItemSchema.Value)
	require.True(t, mt.ItemSchema.Value.Type.Is(TypeObject))

	err = mt.Validate(t.Context())
	require.Error(t, err)
	var fvm *FieldVersionMismatchError
	require.True(t, errors.As(err, &fvm))
	require.Equal(t, "itemSchema", fvm.Field)
	require.Equal(t, "3.2", fvm.MinVersion)

	err = mt.Validate(t.Context(), IsOpenAPI32OrLater())
	require.NoError(t, err)

	lookup, err := mt.JSONLookup("itemSchema")
	require.NoError(t, err)
	require.Same(t, mt.ItemSchema.Value, lookup)

	encoded, err := json.Marshal(mt)
	require.NoError(t, err)
	require.JSONEq(t, string(data), string(encoded))
}
