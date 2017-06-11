package openapi3

import (
	"context"
	"fmt"
	"github.com/jban332/kinapi/jsoninfo"
)

// Parameters is specified by OpenAPI/Swagger 3.0 standard.
type Parameters []*Parameter

func NewParameters() Parameters {
	return make([]*Parameter, 0, 4)
}

func (all Parameters) GetByInAndName(in string, name string) *Parameter {
	for _, item := range all {
		if item.Name == name && item.In == in {
			return item
		}
	}
	return nil
}

func (all Parameters) Validate(c context.Context) error {
	m := make(map[string]struct{})
	for _, item := range all {
		in := item.In
		name := item.Name
		key := in + ":" + name
		if _, exists := m[key]; exists {
			return fmt.Errorf("More than one '%s' parameter has name '%s'", in, name)
		}
		m[key] = struct{}{}
		err := item.Validate(c)
		if err != nil {
			return err
		}
	}
	return nil
}

// Parameter is specified by OpenAPI/Swagger 3.0 standard.
type Parameter struct {
	jsoninfo.RefProps
	jsoninfo.ExtensionProps
	Name            string        `json:"name,omitempty"`
	In              string        `json:"in,omitempty"`
	Description     string        `json:"description,omitempty"`
	Deprecated      bool          `json:"deprecated,omitempty"`
	Required        bool          `json:"required,omitempty"`
	Style           string        `json:"style,omitempty"`
	AllowEmptyValue bool          `json:"allowEmptyValue,omitempty"`
	AllowReserved   bool          `json:"allowReserved,omitempty"`
	Schema          *Schema       `json:"schema,omitempty"`
	Example         interface{}   `json:"example,omitempty"`
	Examples        []interface{} `json:"examples,omitempty"`
}

const (
	ParameterInPath   = "path"
	ParameterInQuery  = "query"
	ParameterInHeader = "header"
	ParameterInCookie = "cookie"
)

func NewPathParameter(name string) *Parameter {
	return &Parameter{
		Name:     name,
		In:       ParameterInPath,
		Required: true,
	}
}

func NewQueryParameter(name string) *Parameter {
	return &Parameter{
		Name: name,
		In:   ParameterInQuery,
	}
}

func NewHeaderParameter(name string) *Parameter {
	return &Parameter{
		Name: name,
		In:   ParameterInHeader,
	}
}

func NewCookieParameter(name string) *Parameter {
	return &Parameter{
		Name: name,
		In:   ParameterInCookie,
	}
}

func (parameter *Parameter) WithDescription(value string) *Parameter {
	parameter.Description = value
	return parameter
}

func (parameter *Parameter) WithRequired(value bool) *Parameter {
	parameter.Required = value
	return parameter
}

func (parameter *Parameter) WithSchema(value *Schema) *Parameter {
	parameter.Schema = value
	return parameter
}

func (value *Parameter) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStructFields(value)
}

func (value *Parameter) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStructFields(data, value)
}

func (parameter *Parameter) Validate(c context.Context) error {
	if parameter.Name == "" {
		return fmt.Errorf("Parameter name can't be blank")
	}
	in := parameter.In
	switch in {
	case
		ParameterInPath,
		ParameterInQuery,
		ParameterInHeader,
		ParameterInCookie:
	default:
		return fmt.Errorf("Parameter can't have 'in' value '%s'", parameter.In)
	}
	if schema := parameter.Schema; schema != nil {
		err := schema.Validate(c)
		if err != nil {
			return fmt.Errorf("Parameter '%v' schema is invalid: %v", parameter.Name, err)
		}
	}
	return nil
}
