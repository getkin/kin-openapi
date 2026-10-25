package openapi2

import (
	"encoding/json"
	"maps"

	"github.com/getkin/kin-openapi/openapi3"
)

type Schemas map[string]*SchemaRef
type SchemaRefs []*SchemaRef

// Schema is specified by OpenAPI/Swagger 2.0 standard.
// See https://swagger.io/specification/v2/#schema-object
type Schema struct {
	Extensions map[string]any `json:"-" yaml:"-"`

	AllOf        SchemaRefs             `json:"allOf,omitempty,case:strict" yaml:"allOf,omitempty"`
	Not          *SchemaRef             `json:"not,omitempty,case:strict" yaml:"not,omitempty"`
	Type         *openapi3.Types        `json:"type,omitempty,case:strict" yaml:"type,omitempty"`
	Title        string                 `json:"title,omitempty,case:strict" yaml:"title,omitempty"`
	Format       string                 `json:"format,omitempty,case:strict" yaml:"format,omitempty"`
	Description  string                 `json:"description,omitempty,case:strict" yaml:"description,omitempty"`
	Enum         []any                  `json:"enum,omitempty,case:strict" yaml:"enum,omitempty"`
	Default      any                    `json:"default,omitempty,case:strict" yaml:"default,omitempty"`
	Example      any                    `json:"example,omitempty,case:strict" yaml:"example,omitempty"`
	ExternalDocs *openapi3.ExternalDocs `json:"externalDocs,omitempty,case:strict" yaml:"externalDocs,omitempty"`

	// Array-related, here for struct compactness
	UniqueItems bool `json:"uniqueItems,omitempty,case:strict" yaml:"uniqueItems,omitempty"`
	// Number-related, here for struct compactness
	ExclusiveMin bool `json:"exclusiveMinimum,omitempty,case:strict" yaml:"exclusiveMinimum,omitempty"`
	ExclusiveMax bool `json:"exclusiveMaximum,omitempty,case:strict" yaml:"exclusiveMaximum,omitempty"`
	// Properties
	ReadOnly        bool          `json:"readOnly,omitempty,case:strict" yaml:"readOnly,omitempty"`
	WriteOnly       bool          `json:"writeOnly,omitempty,case:strict" yaml:"writeOnly,omitempty"`
	AllowEmptyValue bool          `json:"allowEmptyValue,omitempty,case:strict" yaml:"allowEmptyValue,omitempty"`
	Deprecated      bool          `json:"deprecated,omitempty,case:strict" yaml:"deprecated,omitempty"`
	XML             *openapi3.XML `json:"xml,omitempty,case:strict" yaml:"xml,omitempty"`

	// Number
	Min        *float64 `json:"minimum,omitempty,case:strict" yaml:"minimum,omitempty"`
	Max        *float64 `json:"maximum,omitempty,case:strict" yaml:"maximum,omitempty"`
	MultipleOf *float64 `json:"multipleOf,omitempty,case:strict" yaml:"multipleOf,omitempty"`

	// String
	MinLength uint64  `json:"minLength,omitempty,case:strict" yaml:"minLength,omitempty"`
	MaxLength *uint64 `json:"maxLength,omitempty,case:strict" yaml:"maxLength,omitempty"`
	Pattern   string  `json:"pattern,omitempty,case:strict" yaml:"pattern,omitempty"`

	// Array
	MinItems uint64     `json:"minItems,omitempty,case:strict" yaml:"minItems,omitempty"`
	MaxItems *uint64    `json:"maxItems,omitempty,case:strict" yaml:"maxItems,omitempty"`
	Items    *SchemaRef `json:"items,omitempty,case:strict" yaml:"items,omitempty"`

	// Object
	Required             []string                      `json:"required,omitempty,case:strict" yaml:"required,omitempty"`
	Properties           Schemas                       `json:"properties,omitempty,case:strict" yaml:"properties,omitempty"`
	MinProps             uint64                        `json:"minProperties,omitempty,case:strict" yaml:"minProperties,omitempty"`
	MaxProps             *uint64                       `json:"maxProperties,omitempty,case:strict" yaml:"maxProperties,omitempty"`
	AdditionalProperties openapi3.AdditionalProperties `json:"additionalProperties,omitempty,case:strict" yaml:"additionalProperties,omitempty"`
	Discriminator        string                        `json:"discriminator,omitempty,case:strict" yaml:"discriminator,omitempty"`
}

// MarshalJSON returns the JSON encoding of Schema.
func (schema Schema) MarshalJSON() ([]byte, error) {
	m, err := schema.MarshalYAML()
	if err != nil {
		return nil, err
	}

	return json.Marshal(m)
}

// MarshalYAML returns the YAML encoding of Schema.
func (schema Schema) MarshalYAML() (any, error) {
	m := make(map[string]any, 36+len(schema.Extensions))
	maps.Copy(m, schema.Extensions)

	if x := schema.AllOf; len(x) != 0 {
		m["allOf"] = x
	}
	if x := schema.Not; x != nil {
		m["not"] = x
	}
	if x := schema.Type; x != nil {
		m["type"] = x
	}
	if x := schema.Title; len(x) != 0 {
		m["title"] = x
	}
	if x := schema.Format; len(x) != 0 {
		m["format"] = x
	}
	if x := schema.Description; len(x) != 0 {
		m["description"] = x
	}
	if x := schema.Enum; len(x) != 0 {
		m["enum"] = x
	}
	if x := schema.Default; x != nil {
		m["default"] = x
	}
	if x := schema.Example; x != nil {
		m["example"] = x
	}
	if x := schema.ExternalDocs; x != nil {
		m["externalDocs"] = x
	}

	// Array-related
	if x := schema.UniqueItems; x {
		m["uniqueItems"] = x
	}
	// Number-related
	if x := schema.ExclusiveMin; x {
		m["exclusiveMinimum"] = x
	}
	if x := schema.ExclusiveMax; x {
		m["exclusiveMaximum"] = x
	}
	if x := schema.ReadOnly; x {
		m["readOnly"] = x
	}
	if x := schema.WriteOnly; x {
		m["writeOnly"] = x
	}
	if x := schema.AllowEmptyValue; x {
		m["allowEmptyValue"] = x
	}
	if x := schema.Deprecated; x {
		m["deprecated"] = x
	}
	if x := schema.XML; x != nil {
		m["xml"] = x
	}

	// Number
	if x := schema.Min; x != nil {
		m["minimum"] = x
	}
	if x := schema.Max; x != nil {
		m["maximum"] = x
	}
	if x := schema.MultipleOf; x != nil {
		m["multipleOf"] = x
	}

	// String
	if x := schema.MinLength; x != 0 {
		m["minLength"] = x
	}
	if x := schema.MaxLength; x != nil {
		m["maxLength"] = x
	}
	if x := schema.Pattern; x != "" {
		m["pattern"] = x
	}

	// Array
	if x := schema.MinItems; x != 0 {
		m["minItems"] = x
	}
	if x := schema.MaxItems; x != nil {
		m["maxItems"] = x
	}
	if x := schema.Items; x != nil {
		m["items"] = x
	}

	// Object
	if x := schema.Required; len(x) != 0 {
		m["required"] = x
	}
	if x := schema.Properties; len(x) != 0 {
		m["properties"] = x
	}
	if x := schema.MinProps; x != 0 {
		m["minProperties"] = x
	}
	if x := schema.MaxProps; x != nil {
		m["maxProperties"] = x
	}
	if x := schema.AdditionalProperties; x.Has != nil || x.Schema != nil {
		m["additionalProperties"] = &x
	}
	if x := schema.Discriminator; x != "" {
		m["discriminator"] = x
	}

	return m, nil
}

// UnmarshalJSON sets Schema to a copy of data.
func (schema *Schema) UnmarshalJSON(data []byte) error {
	type SchemaBis Schema
	var x SchemaBis
	if err := unmarshalJSON(data, &x); err != nil {
		return unmarshalError(err)
	}
	_ = unmarshalJSON(data, &x.Extensions)

	delete(x.Extensions, "oneOf")
	delete(x.Extensions, "anyOf")
	delete(x.Extensions, "allOf")
	delete(x.Extensions, "not")
	delete(x.Extensions, "type")
	delete(x.Extensions, "title")
	delete(x.Extensions, "format")
	delete(x.Extensions, "description")
	delete(x.Extensions, "enum")
	delete(x.Extensions, "default")
	delete(x.Extensions, "example")
	delete(x.Extensions, "externalDocs")

	// Array-related
	delete(x.Extensions, "uniqueItems")
	// Number-related
	delete(x.Extensions, "exclusiveMinimum")
	delete(x.Extensions, "exclusiveMaximum")
	// Properties
	delete(x.Extensions, "nullable")
	delete(x.Extensions, "readOnly")
	delete(x.Extensions, "writeOnly")
	delete(x.Extensions, "allowEmptyValue")
	delete(x.Extensions, "deprecated")
	delete(x.Extensions, "xml")

	// Number
	delete(x.Extensions, "minimum")
	delete(x.Extensions, "maximum")
	delete(x.Extensions, "multipleOf")

	// String
	delete(x.Extensions, "minLength")
	delete(x.Extensions, "maxLength")
	delete(x.Extensions, "pattern")

	// Array
	delete(x.Extensions, "minItems")
	delete(x.Extensions, "maxItems")
	delete(x.Extensions, "items")

	// Object
	delete(x.Extensions, "required")
	delete(x.Extensions, "properties")
	delete(x.Extensions, "minProperties")
	delete(x.Extensions, "maxProperties")
	delete(x.Extensions, "additionalProperties")
	delete(x.Extensions, "discriminator")

	if len(x.Extensions) == 0 {
		x.Extensions = nil
	}

	*schema = Schema(x)
	return nil
}
