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

// MarshalJSONUnsupportedFields will be invoked by package "jsoninfo"
func (extensionProps *ExtensionProps) MarshalJSONUnsupportedFields(dest map[string]json.RawMessage) error {
	extensions := extensionProps.Extensions
	if extensions != nil {
		for k, v := range extensions {
			dest[k] = v
		}
	}
	return nil
}

// UnmarshalJSONUnsupportedFields will be invoked by package "jsoninfo"
func (extensionProps *ExtensionProps) UnmarshalJSONUnsupportedFields(data []byte, extensions map[string]json.RawMessage) error {
	extensionProps.Extensions = extensions
	return nil
}
