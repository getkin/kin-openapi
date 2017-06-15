package openapi3

import (
	"encoding/json"
	"github.com/jban332/kin-openapi/jsoninfo"
)

// ExtensionProps provides support for OpenAPI extensions.
// It reads/writes all properties that begin with "x-".
type ExtensionProps struct {
	Extensions map[string]json.RawMessage `json:"-"`
}

// Assert that the type implements the interface
var _ jsoninfo.StrictStruct = &ExtensionProps{}

// EncodeWith will be invoked by package "jsoninfo"
func (props *ExtensionProps) EncodeWith(encoder *jsoninfo.ObjectEncoder, value interface{}) error {
	encoder.EncodeExtensionMap(props.Extensions)
	return encoder.EncodeStructFieldsAndExtensions(value)
}

// DecodeWith will be invoked by package "jsoninfo"
func (props *ExtensionProps) DecodeWith(decoder *jsoninfo.ObjectDecoder, value interface{}) error {
	props.Extensions = decoder.DecodeExtensionMap()
	return decoder.DecodeStructFieldsAndExtensions(value)
}
