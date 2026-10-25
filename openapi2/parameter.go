package openapi2

import (
	"encoding/json"
	"maps"

	"github.com/getkin/kin-openapi/openapi3"
)

type Parameters []*Parameter

type Parameter struct {
	Extensions map[string]any `json:"-" yaml:"-"`

	Ref string `json:"$ref,omitempty,case:strict" yaml:"$ref,omitempty"`

	In               string          `json:"in,omitempty,case:strict" yaml:"in,omitempty"`
	Name             string          `json:"name,omitempty,case:strict" yaml:"name,omitempty"`
	Description      string          `json:"description,omitempty,case:strict" yaml:"description,omitempty"`
	CollectionFormat string          `json:"collectionFormat,omitempty,case:strict" yaml:"collectionFormat,omitempty"`
	Type             *openapi3.Types `json:"type,omitempty,case:strict" yaml:"type,omitempty"`
	Format           string          `json:"format,omitempty,case:strict" yaml:"format,omitempty"`
	Pattern          string          `json:"pattern,omitempty,case:strict" yaml:"pattern,omitempty"`
	AllowEmptyValue  bool            `json:"allowEmptyValue,omitempty,case:strict" yaml:"allowEmptyValue,omitempty"`
	Required         bool            `json:"required,omitempty,case:strict" yaml:"required,omitempty"`
	UniqueItems      bool            `json:"uniqueItems,omitempty,case:strict" yaml:"uniqueItems,omitempty"`
	ExclusiveMin     bool            `json:"exclusiveMinimum,omitempty,case:strict" yaml:"exclusiveMinimum,omitempty"`
	ExclusiveMax     bool            `json:"exclusiveMaximum,omitempty,case:strict" yaml:"exclusiveMaximum,omitempty"`
	Schema           *SchemaRef      `json:"schema,omitempty,case:strict" yaml:"schema,omitempty"`
	Items            *SchemaRef      `json:"items,omitempty,case:strict" yaml:"items,omitempty"`
	Enum             []any           `json:"enum,omitempty,case:strict" yaml:"enum,omitempty"`
	MultipleOf       *float64        `json:"multipleOf,omitempty,case:strict" yaml:"multipleOf,omitempty"`
	Minimum          *float64        `json:"minimum,omitempty,case:strict" yaml:"minimum,omitempty"`
	Maximum          *float64        `json:"maximum,omitempty,case:strict" yaml:"maximum,omitempty"`
	MaxLength        *uint64         `json:"maxLength,omitempty,case:strict" yaml:"maxLength,omitempty"`
	MaxItems         *uint64         `json:"maxItems,omitempty,case:strict" yaml:"maxItems,omitempty"`
	MinLength        uint64          `json:"minLength,omitempty,case:strict" yaml:"minLength,omitempty"`
	MinItems         uint64          `json:"minItems,omitempty,case:strict" yaml:"minItems,omitempty"`
	Default          any             `json:"default,omitempty,case:strict" yaml:"default,omitempty"`
}

// MarshalJSON returns the JSON encoding of Parameter.
func (parameter Parameter) MarshalJSON() ([]byte, error) {
	if ref := parameter.Ref; ref != "" {
		return json.Marshal(openapi3.Ref{Ref: ref})
	}

	m := make(map[string]any, 24+len(parameter.Extensions))
	maps.Copy(m, parameter.Extensions)

	if x := parameter.In; x != "" {
		m["in"] = x
	}
	if x := parameter.Name; x != "" {
		m["name"] = x
	}
	if x := parameter.Description; x != "" {
		m["description"] = x
	}
	if x := parameter.CollectionFormat; x != "" {
		m["collectionFormat"] = x
	}
	if x := parameter.Type; x != nil {
		m["type"] = x
	}
	if x := parameter.Format; x != "" {
		m["format"] = x
	}
	if x := parameter.Pattern; x != "" {
		m["pattern"] = x
	}
	if x := parameter.AllowEmptyValue; x {
		m["allowEmptyValue"] = x
	}
	if x := parameter.Required; x {
		m["required"] = x
	}
	if x := parameter.UniqueItems; x {
		m["uniqueItems"] = x
	}
	if x := parameter.ExclusiveMin; x {
		m["exclusiveMinimum"] = x
	}
	if x := parameter.ExclusiveMax; x {
		m["exclusiveMaximum"] = x
	}
	if x := parameter.Schema; x != nil {
		m["schema"] = x
	}
	if x := parameter.Items; x != nil {
		m["items"] = x
	}
	if x := parameter.Enum; x != nil {
		m["enum"] = x
	}
	if x := parameter.MultipleOf; x != nil {
		m["multipleOf"] = x
	}
	if x := parameter.Minimum; x != nil {
		m["minimum"] = x
	}
	if x := parameter.Maximum; x != nil {
		m["maximum"] = x
	}
	if x := parameter.MaxLength; x != nil {
		m["maxLength"] = x
	}
	if x := parameter.MaxItems; x != nil {
		m["maxItems"] = x
	}
	if x := parameter.MinLength; x != 0 {
		m["minLength"] = x
	}
	if x := parameter.MinItems; x != 0 {
		m["minItems"] = x
	}
	if x := parameter.Default; x != nil {
		m["default"] = x
	}

	return json.Marshal(m)
}

// UnmarshalJSON sets Parameter to a copy of data.
func (parameter *Parameter) UnmarshalJSON(data []byte) error {
	type ParameterBis Parameter
	var x ParameterBis
	if err := unmarshalJSON(data, &x); err != nil {
		return unmarshalError(err)
	}
	_ = unmarshalJSON(data, &x.Extensions)
	delete(x.Extensions, "$ref")

	delete(x.Extensions, "in")
	delete(x.Extensions, "name")
	delete(x.Extensions, "description")
	delete(x.Extensions, "collectionFormat")
	delete(x.Extensions, "type")
	delete(x.Extensions, "format")
	delete(x.Extensions, "pattern")
	delete(x.Extensions, "allowEmptyValue")
	delete(x.Extensions, "required")
	delete(x.Extensions, "uniqueItems")
	delete(x.Extensions, "exclusiveMinimum")
	delete(x.Extensions, "exclusiveMaximum")
	delete(x.Extensions, "schema")
	delete(x.Extensions, "items")
	delete(x.Extensions, "enum")
	delete(x.Extensions, "multipleOf")
	delete(x.Extensions, "minimum")
	delete(x.Extensions, "maximum")
	delete(x.Extensions, "maxLength")
	delete(x.Extensions, "maxItems")
	delete(x.Extensions, "minLength")
	delete(x.Extensions, "minItems")
	delete(x.Extensions, "default")

	if len(x.Extensions) == 0 {
		x.Extensions = nil
	}

	*parameter = Parameter(x)
	return nil
}
