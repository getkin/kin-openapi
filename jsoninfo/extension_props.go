package jsoninfo

import (
	"encoding/json"
)

type Extension interface {
	UnmarshalJSONExtension(data []byte, holder ExtensionHolder) error
}

// ExtensionProps provides support for OpenAPI extensions.
// It reads/writes all properties that begin with "x-".
type ExtensionProps struct {
	Extensions            []Extension                `json:"-"`
	UnsupportedExtensions map[string]json.RawMessage `json:"-"`
}

func (extensionProps *ExtensionProps) AddExtension(value Extension) {
	extensionProps.Extensions = append(extensionProps.Extensions, value)
}

func (extensionProps *ExtensionProps) GetExtensionProps() *ExtensionProps {
	return extensionProps
}

// ExtensionHolder interface is implemented by ExtensionProps.
//
// You probably shouldn't implement this interface yourself.
type ExtensionHolder interface {
	GetExtensionProps() *ExtensionProps
}

// ExtensionPrefixesExpectedToBeSupported contains JSON property prefixes.
//
// If an unmarshalled JSON property starts with any of these strings, the property
// must be accepted either by the struct or one of its extension handlers.
var ExtensionPrefixesExpectedToBeSupported = []string{}
