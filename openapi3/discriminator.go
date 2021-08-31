package openapi3

import (
	"context"

	"github.com/getkin/kin-openapi/jsoninfo"
)

// Discriminator is specified by OpenAPI/Swagger standard version 3.0.
type Discriminator struct {
	ExtensionProps
	PropertyName string            `json:"propertyName" yaml:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty" yaml:"mapping,omitempty"`
}

func (discr *Discriminator) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(discr)
}

func (discr *Discriminator) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, discr)
}

// Validate goes through the receiver value and its descendants and errors on any non compliance to the OpenAPIv3 specification.
func (discr *Discriminator) Validate(ctx context.Context) error {
	return nil
}
