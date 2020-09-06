package jsoninfo

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewObjectDecoder(t *testing.T) {
	data := []byte(`
	{
		"field1": 1,
		"field2": 2
	}
`)
	t.Run("test new object decoder", func(t *testing.T) {
		decoder, err := NewObjectDecoder(data)
		assert.Nil(t, err)
		assert.NotNil(t, decoder)
		assert.Equal(t, data, decoder.Data)
		assert.Equal(t, 2, len(decoder.DecodeExtensionMap()))
	})
}

type mockStrictStruct struct {
	EncodeWithFn func(encoder *ObjectEncoder, value interface{}) error
	DecodeWithFn func(decoder *ObjectDecoder, value interface{}) error
}

func (m *mockStrictStruct) EncodeWith(encoder *ObjectEncoder, value interface{}) error {
	return m.EncodeWithFn(encoder, value)
}

func (m *mockStrictStruct) DecodeWith(decoder *ObjectDecoder, value interface{}) error {
	return m.DecodeWithFn(decoder, value)
}

func TestUnmarshalStrictStruct(t *testing.T) {
	data := []byte(`
			{
				"field1": 1,
				"field2": 2
			}
		`)

	t.Run("test unmarshal with StrictStruct without err", func(t *testing.T) {
		decodeWithFnCalled := 0
		mockStruct := &mockStrictStruct{
			EncodeWithFn: func(encoder *ObjectEncoder, value interface{}) error {
				return nil
			},
			DecodeWithFn: func(decoder *ObjectDecoder, value interface{}) error {
				decodeWithFnCalled++
				return nil
			},
		}
		err := UnmarshalStrictStruct(data, mockStruct)
		assert.Nil(t, err)
		assert.Equal(t, 1, decodeWithFnCalled)
	})

	t.Run("test unmarshal with StrictStruct with err", func(t *testing.T) {
		decodeWithFnCalled := 0
		mockStruct := &mockStrictStruct{
			EncodeWithFn: func(encoder *ObjectEncoder, value interface{}) error {
				return nil
			},
			DecodeWithFn: func(decoder *ObjectDecoder, value interface{}) error {
				decodeWithFnCalled++
				return errors.New("unable to decode the value")
			},
		}
		err := UnmarshalStrictStruct(data, mockStruct)
		assert.NotNil(t, err)
		assert.Equal(t, 1, decodeWithFnCalled)
	})
}

func TestDecodeStructFieldsAndExtensions(t *testing.T) {
	data := []byte(`
	{
		"field1": "field1",
		"field2": "field2"
	}
`)
	decoder, err := NewObjectDecoder(data)
	assert.Nil(t, err)
	assert.NotNil(t, decoder)

	t.Run("value is not pointer", func(t *testing.T) {
		var value interface{}
		assert.Panics(t, func() {
			_ = decoder.DecodeStructFieldsAndExtensions(value)
		}, "value is not a pointer")
	})

	t.Run("value is nil", func(t *testing.T) {
		var value *string = nil
		assert.Panics(t, func() {
			_ = decoder.DecodeStructFieldsAndExtensions(value)
		}, "value is nil")
	})

	t.Run("value is not struct", func(t *testing.T) {
		var value = "simple string"
		assert.Panics(t, func() {
			_ = decoder.DecodeStructFieldsAndExtensions(&value)
		}, "value is not struct")
	})

	t.Run("successfully decoded with all fields", func(t *testing.T) {
		d, err := NewObjectDecoder(data)
		assert.Nil(t, err)
		assert.NotNil(t, d)

		var value = struct {
			Field1 string `json:"field1"`
			Field2 string `json:"field2"`
		}{}
		err = d.DecodeStructFieldsAndExtensions(&value)
		assert.Nil(t, err)
		assert.Equal(t, "field1", value.Field1)
		assert.Equal(t, "field2", value.Field2)
		assert.Equal(t, 0, len(d.DecodeExtensionMap()))
	})

	t.Run("successfully decoded with renaming field", func(t *testing.T) {
		d, err := NewObjectDecoder(data)
		assert.Nil(t, err)
		assert.NotNil(t, d)

		var value = struct {
			Field1 string `json:"field1"`
		}{}
		err = d.DecodeStructFieldsAndExtensions(&value)
		assert.Nil(t, err)
		assert.Equal(t, "field1", value.Field1)
		assert.Equal(t, 1, len(d.DecodeExtensionMap()))
	})

	t.Run("un-successfully decoded due to data mismatch", func(t *testing.T) {
		d, err := NewObjectDecoder(data)
		assert.Nil(t, err)
		assert.NotNil(t, d)

		var value = struct {
			Field1 int `json:"field1"`
		}{}
		err = d.DecodeStructFieldsAndExtensions(&value)
		assert.NotNil(t, err)
		assert.EqualError(t, err, "Error while unmarshalling property 'field1' (*int): json: cannot unmarshal string into Go value of type int")
		assert.Equal(t, 0, value.Field1)
		assert.Equal(t, 2, len(d.DecodeExtensionMap()))
	})
}
