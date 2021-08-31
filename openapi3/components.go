package openapi3

import (
	"context"

	"github.com/getkin/kin-openapi/jsoninfo"
)

// Components is specified by OpenAPI/Swagger standard version 3.0.
type Components struct {
	ExtensionProps
	Schemas         Schemas         `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Parameters      ParametersMap   `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Headers         Headers         `json:"headers,omitempty" yaml:"headers,omitempty"`
	RequestBodies   RequestBodies   `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
	Responses       Responses       `json:"responses,omitempty" yaml:"responses,omitempty"`
	SecuritySchemes SecuritySchemes `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Examples        Examples        `json:"examples,omitempty" yaml:"examples,omitempty"`
	Links           Links           `json:"links,omitempty" yaml:"links,omitempty"`
	Callbacks       Callbacks       `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
}

func NewComponents() Components {
	return Components{}
}

func (components *Components) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(components)
}

func (components *Components) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, components)
}

// Validate goes through the receiver value and its descendants and errors on any non compliance to the OpenAPIv3 specification.
func (components *Components) Validate(ctx context.Context) (err error) {
	for k, v := range components.Schemas {
		if err = ValidateIdentifier(k); err != nil {
			return
		}
		if err = v.Validate(ctx); err != nil {
			return
		}
	}

	for k, v := range components.Parameters {
		if err = ValidateIdentifier(k); err != nil {
			return
		}
		if err = v.Validate(ctx); err != nil {
			return
		}
	}

	for k, v := range components.RequestBodies {
		if err = ValidateIdentifier(k); err != nil {
			return
		}
		if err = v.Validate(ctx); err != nil {
			return
		}
	}

	for k, v := range components.Responses {
		if err = ValidateIdentifier(k); err != nil {
			return
		}
		if err = v.Validate(ctx); err != nil {
			return
		}
	}

	for k, v := range components.Headers {
		if err = ValidateIdentifier(k); err != nil {
			return
		}
		if err = v.Validate(ctx); err != nil {
			return
		}
	}

	for k, v := range components.SecuritySchemes {
		if err = ValidateIdentifier(k); err != nil {
			return
		}
		if err = v.Validate(ctx); err != nil {
			return
		}
	}

	return
}
