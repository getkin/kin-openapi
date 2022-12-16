package openapi3

import (
	"context"

	"github.com/getkin/kin-openapi/jsoninfo"
)

// Discriminator is specified by OpenAPI/Swagger standard version 3.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#discriminator-object
type Discriminator struct {
	ExtensionProps `json:"-" yaml:"-"`

	PropertyName string            `json:"propertyName" yaml:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty" yaml:"mapping,omitempty"`
}

// MarshalJSON returns the JSON encoding of Discriminator.
func (discriminator *Discriminator) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(discriminator)
}

// UnmarshalJSON sets Discriminator to a copy of data.
func (discriminator *Discriminator) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, discriminator)
}

// Validate returns an error if Discriminator does not comply with the OpenAPI spec.
func (discriminator *Discriminator) Validate(ctx context.Context, opts ...ValidationOption) error {
	// ctx = WithValidationOptions(ctx, opts...)

	return nil
}
