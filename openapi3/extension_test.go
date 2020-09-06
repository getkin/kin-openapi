package openapi3

import (
	"testing"

	"github.com/getkin/kin-openapi/jsoninfo"
	"github.com/stretchr/testify/assert"
)

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
		assert.Nil(t, err)
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
		assert.Nil(t, err)
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
		assert.Nil(t, err)
		assert.Equal(t, 0, len(extensionProps.Extensions))
		assert.Equal(t, "value1", value.Field1)
		assert.Equal(t, "value2", value.Field2)
	})

	t.Run("successfully decode some of the fields", func(t *testing.T) {
		decoder, err := jsoninfo.NewObjectDecoder(data)
		assert.Nil(t, err)
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
		assert.Nil(t, err)
		assert.Equal(t, 1, len(extensionProps.Extensions))
		assert.Equal(t, "value1", value.Field1)
	})

	t.Run("successfully decode none of the fields", func(t *testing.T) {
		decoder, err := jsoninfo.NewObjectDecoder(data)
		assert.Nil(t, err)

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
		assert.Nil(t, err)
		assert.Equal(t, 2, len(extensionProps.Extensions))
		assert.Empty(t, value.Field3)
		assert.Empty(t, value.Field4)
	})
}
