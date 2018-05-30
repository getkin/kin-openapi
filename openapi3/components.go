package openapi3

import (
	"context"
	"fmt"
	"github.com/ronniedada/kin-openapi/jsoninfo"
	"regexp"
)

// Components is specified by OpenAPI/Swagger standard version 3.0.
type Components struct {
	ExtensionProps
	Schemas         map[string]*SchemaRef         `json:"schemas,omitempty"`
	Parameters      map[string]*ParameterRef      `json:"parameters,omitempty"`
	Headers         map[string]*HeaderRef         `json:"headers,omitempty"`
	RequestBodies   map[string]*RequestBodyRef    `json:"requestBodies,omitempty"`
	Responses       map[string]*ResponseRef       `json:"responses,omitempty"`
	SecuritySchemes map[string]*SecuritySchemeRef `json:"securitySchemes,omitempty"`
	Examples        map[string]*ExampleRef        `json:"examples,omitempty"`
	Tags            Tags                          `json:"tags,omitempty"`
	Links           map[string]*LinkRef           `json:"links,omitempty"`
	Callbacks       map[string]*CallbackRef       `json:"callbacks,omitempty"`
}

func NewComponents() Components {
	return Components{}
}

func (value *Components) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(value)
}

func (value *Components) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, value)
}

func (components *Components) Validate(c context.Context) error {
	if m := components.Schemas; m != nil {
		for k, v := range m {
			if err := ValidateIdentifier(k); err != nil {
				return err
			}
			if err := v.Validate(c); err != nil {
				return err
			}
		}
	}
	if m := components.Parameters; m != nil {
		for k, v := range m {
			if err := ValidateIdentifier(k); err != nil {
				return err
			}
			if err := v.Validate(c); err != nil {
				return err
			}
		}
	}
	if m := components.RequestBodies; m != nil {
		for k, v := range m {
			if err := ValidateIdentifier(k); err != nil {
				return err
			}
			if err := v.Validate(c); err != nil {
				return err
			}
		}
	}
	if m := components.Responses; m != nil {
		for k, v := range m {
			if err := ValidateIdentifier(k); err != nil {
				return err
			}
			if err := v.Validate(c); err != nil {
				return err
			}
		}
	}
	if m := components.Headers; m != nil {
		for k, v := range m {
			if err := ValidateIdentifier(k); err != nil {
				return err
			}
			if err := v.Validate(c); err != nil {
				return err
			}
		}
	}
	if m := components.SecuritySchemes; m != nil {
		for k, v := range m {
			if err := ValidateIdentifier(k); err != nil {
				return err
			}
			if err := v.Validate(c); err != nil {
				return err
			}
		}
	}
	return nil
}

const identifierPattern = `^[a-zA-Z0-9.\-_]+$`

var identifierRegExp = regexp.MustCompile(identifierPattern)

func ValidateIdentifier(value string) error {
	if identifierRegExp.MatchString(value) {
		return nil
	}
	return fmt.Errorf("Identifier '%s' is not supported by OpenAPI version 3 standard (regexp: '%s')", value, identifierPattern)
}
