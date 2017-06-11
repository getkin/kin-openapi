package openapi3

import (
	"context"
	"errors"
	"github.com/jban332/kinapi/jsoninfo"
)

// Components is specified by OpenAPI/Swagger standard version 3.0.
type Components struct {
	jsoninfo.ExtensionProps
	Schemas         map[string]*Schema         `json:"schemas,omitempty,noref"`
	Parameters      map[string]*Parameter      `json:"parameters,omitempty,noref"`
	Headers         map[string]*Parameter      `json:"headers,omitempty"`
	RequestBodies   map[string]*RequestBody    `json:"requestBodies,omitempty,noref"`
	Responses       map[string]*Response       `json:"responses,omitempty,noref"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty"`
	Examples        map[string]interface{}     `json:"examples,omitempty,noref"`
	Tags            Tags                       `json:"tags,omitempty"`
	Links           map[string]*Link           `json:"links,omitempty,noref"`
	Callbacks       map[string]*Callback       `json:"callbacks,omitempty,noref"`
}

func NewComponents() Components {
	return Components{}
}

func (value *Components) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStructFields(value)
}

func (value *Components) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStructFields(data, value)
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
			if v.Name != "" {
				return errors.New("Header component can't contain 'name'")
			}
			if v.In != "" {
				return errors.New("Header component can't contain 'in'")
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
