package openapi3

import (
	"github.com/getkin/kin-openapi/jsoninfo"
)

// Extensions provides support for OpenAPI extensions.
// It reads/writes all properties that begin with "x-".
type Extensions map[string]interface{}

// AsMap returns the underlying map structure
func (props Extensions) AsMap() map[string]interface{} { return props }

// Assert that the type implements the interface
var _ jsoninfo.StrictStruct = &Extensions{}

// EncodeWith will be invoked by package "jsoninfo"
func (props Extensions) EncodeWith(encoder *jsoninfo.ObjectEncoder, value interface{}) error {
	for k, v := range props {
		if err := encoder.EncodeExtension(k, v); err != nil {
			return err
		}
	}
	return encoder.EncodeStructFieldsAndExtensions(value)
}

// DecodeWith will be invoked by package "jsoninfo"
func (props *Extensions) DecodeWith(decoder *jsoninfo.ObjectDecoder, value interface{}) error {
	if err := decoder.DecodeStructFieldsAndExtensions(value); err != nil {
		return err
	}
	source := decoder.DecodeExtensionMap()
	result := make(map[string]interface{}, len(source))
	for k, v := range source {
		result[k] = v
	}
	*props = result
	return nil
}
