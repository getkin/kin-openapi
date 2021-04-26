package openapi3

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/getkin/kin-openapi/jsoninfo"
	"github.com/stretchr/testify/require"
)

func ExampleExtensionProps_DecodeWith() {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	spec, err := loader.LoadFromFile("testdata/testref.openapi.json")
	if err != nil {
		panic(err)
	}

	dec, err := jsoninfo.NewObjectDecoder(spec.Info.Extensions["x-my-extension"].(json.RawMessage))
	if err != nil {
		panic(err)
	}
	var value struct {
		Key int `json:"k"`
	}
	if err = spec.Info.DecodeWith(dec, &value); err != nil {
		panic(err)
	}
	fmt.Println(value.Key)
	// Output: 42
}

func TestExtensionProps_EncodeWith(t *testing.T) {
	t.Run("successfully encoded", func(t *testing.T) {
		encoder := jsoninfo.NewObjectEncoder()
		var extensionProps = ExtensionProps{
			Extensions: map[string]interface{}{
				"field1": "value1",
			},
		}

		var value = struct {
			Field1 string `json:"field1"`
			Field2 string `json:"field2"`
		}{}

		err := extensionProps.EncodeWith(encoder, &value)
		require.NoError(t, err)
	})
}

func TestExtensionProps_DecodeWith(t *testing.T) {
	data := []byte(`
	{
		"field1": "value1",
		"field2": "value2"
	}
`)
	t.Run("successfully decode all the fields", func(t *testing.T) {
		decoder, err := jsoninfo.NewObjectDecoder(data)
		require.NoError(t, err)
		var extensionProps = &ExtensionProps{
			Extensions: map[string]interface{}{
				"field1": "value1",
				"field2": "value1",
			},
		}

		var value = struct {
			Field1 string `json:"field1"`
			Field2 string `json:"field2"`
		}{}

		err = extensionProps.DecodeWith(decoder, &value)
		require.NoError(t, err)
		require.Equal(t, 0, len(extensionProps.Extensions))
		require.Equal(t, "value1", value.Field1)
		require.Equal(t, "value2", value.Field2)
	})

	t.Run("successfully decode some of the fields", func(t *testing.T) {
		decoder, err := jsoninfo.NewObjectDecoder(data)
		require.NoError(t, err)
		var extensionProps = &ExtensionProps{
			Extensions: map[string]interface{}{
				"field1": "value1",
				"field2": "value2",
			},
		}

		var value = &struct {
			Field1 string `json:"field1"`
		}{}

		err = extensionProps.DecodeWith(decoder, value)
		require.NoError(t, err)
		require.Equal(t, 1, len(extensionProps.Extensions))
		require.Equal(t, "value1", value.Field1)
	})

	t.Run("successfully decode none of the fields", func(t *testing.T) {
		decoder, err := jsoninfo.NewObjectDecoder(data)
		require.NoError(t, err)

		var extensionProps = &ExtensionProps{
			Extensions: map[string]interface{}{
				"field1": "value1",
				"field2": "value2",
			},
		}

		var value = struct {
			Field3 string `json:"field3"`
			Field4 string `json:"field4"`
		}{}

		err = extensionProps.DecodeWith(decoder, &value)
		require.NoError(t, err)
		require.Equal(t, 2, len(extensionProps.Extensions))
		require.Empty(t, value.Field3)
		require.Empty(t, value.Field4)
	})
}
