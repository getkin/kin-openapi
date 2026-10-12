package openapi3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"sync"

	"github.com/go-openapi/jsonpointer"
)

const (
	TypeArray   = "array"
	TypeBoolean = "boolean"
	TypeInteger = "integer"
	TypeNumber  = "number"
	TypeObject  = "object"
	TypeString  = "string"
	TypeNull    = "null"
)

var (
	errSchema = errors.New("input does not match the schema")

	// ErrOneOfConflict is the SchemaError Origin when data matches more than one oneOf schema
	ErrOneOfConflict = errors.New("input matches more than one oneOf schemas")

	// ErrSchemaInputNaN may be returned when validating a number
	ErrSchemaInputNaN = errors.New("floating point NaN is not allowed")
	// ErrSchemaInputInf may be returned when validating a number
	ErrSchemaInputInf = errors.New("floating point Inf is not allowed")

	compiledPatterns sync.Map
)

// NewSchemaRef simply builds a SchemaRef
func NewSchemaRef(ref string, value *Schema) *SchemaRef {
	return &SchemaRef{
		Ref:   ref,
		Value: value,
	}
}

type SchemaRefs []*SchemaRef

var _ jsonpointer.JSONPointable = (*SchemaRefs)(nil)

// JSONLookup implements https://pkg.go.dev/github.com/go-openapi/jsonpointer#JSONPointable
func (s SchemaRefs) JSONLookup(token string) (any, error) {
	i, err := strconv.ParseUint(token, 10, 64)
	if err != nil {
		return nil, err
	}

	if i >= uint64(len(s)) {
		return nil, fmt.Errorf("index out of range: %d", i)
	}

	ref := s[i]

	if ref == nil || ref.Ref != "" {
		return &Ref{Ref: ref.Ref}, nil
	}
	return ref.Value, nil
}

// Schema is specified by OpenAPI/Swagger 3.0 standard.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#schema-object
// and https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.0.md#schema-object
type Schema struct {
	Extensions map[string]any `json:"-" yaml:"-"`
	Origin     *Origin        `json:"-" yaml:"-"`

	OneOf        SchemaRefs    `json:"oneOf,omitempty,case:strict" yaml:"oneOf,omitempty"`
	AnyOf        SchemaRefs    `json:"anyOf,omitempty,case:strict" yaml:"anyOf,omitempty"`
	AllOf        SchemaRefs    `json:"allOf,omitempty,case:strict" yaml:"allOf,omitempty"`
	Not          *SchemaRef    `json:"not,omitempty,case:strict" yaml:"not,omitempty"`
	Type         *Types        `json:"type,omitempty,case:strict" yaml:"type,omitempty"`
	Title        string        `json:"title,omitempty,case:strict" yaml:"title,omitempty"`
	Format       string        `json:"format,omitempty,case:strict" yaml:"format,omitempty"`
	Description  string        `json:"description,omitempty,case:strict" yaml:"description,omitempty"`
	Enum         []any         `json:"enum,omitempty,case:strict" yaml:"enum,omitempty"`
	Default      any           `json:"default,omitempty,case:strict" yaml:"default,omitempty"`
	Example      any           `json:"example,omitempty,case:strict" yaml:"example,omitempty"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty,case:strict" yaml:"externalDocs,omitempty"`

	// Array-related, here for struct compactness
	UniqueItems bool `json:"uniqueItems,omitempty,case:strict" yaml:"uniqueItems,omitempty"`
	// Number-related, here for struct compactness
	ExclusiveMin ExclusiveBound `json:"exclusiveMinimum,omitempty,case:strict" yaml:"exclusiveMinimum,omitempty"` // Number for v3.1+ otherwise boolean
	ExclusiveMax ExclusiveBound `json:"exclusiveMaximum,omitempty,case:strict" yaml:"exclusiveMaximum,omitempty"` // Number for v3.1+ otherwise boolean
	// Properties
	Nullable        bool `json:"nullable,omitempty,case:strict" yaml:"nullable,omitempty"`
	ReadOnly        bool `json:"readOnly,omitempty,case:strict" yaml:"readOnly,omitempty"`
	WriteOnly       bool `json:"writeOnly,omitempty,case:strict" yaml:"writeOnly,omitempty"`
	AllowEmptyValue bool `json:"allowEmptyValue,omitempty,case:strict" yaml:"allowEmptyValue,omitempty"`
	Deprecated      bool `json:"deprecated,omitempty,case:strict" yaml:"deprecated,omitempty"`
	XML             *XML `json:"xml,omitempty,case:strict" yaml:"xml,omitempty"`

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
	Required             []string             `json:"required,omitempty,case:strict" yaml:"required,omitempty"`
	Properties           Schemas              `json:"properties,omitempty,case:strict" yaml:"properties,omitempty"`
	MinProps             uint64               `json:"minProperties,omitempty,case:strict" yaml:"minProperties,omitempty"`
	MaxProps             *uint64              `json:"maxProperties,omitempty,case:strict" yaml:"maxProperties,omitempty"`
	AdditionalProperties AdditionalProperties `json:"additionalProperties,omitempty,case:strict" yaml:"additionalProperties,omitempty"`
	Discriminator        *Discriminator       `json:"discriminator,omitempty,case:strict" yaml:"discriminator,omitempty"`

	Const                 any        `json:"const,omitempty,case:strict" yaml:"const,omitempty"`                                 // OpenAPI >=3.1
	Examples              []any      `json:"examples,omitempty,case:strict" yaml:"examples,omitempty"`                           // OpenAPI >=3.1
	PrefixItems           SchemaRefs `json:"prefixItems,omitempty,case:strict" yaml:"prefixItems,omitempty"`                     // OpenAPI >=3.1
	Contains              *SchemaRef `json:"contains,omitempty,case:strict" yaml:"contains,omitempty"`                           // OpenAPI >=3.1
	MinContains           *uint64    `json:"minContains,omitempty,case:strict" yaml:"minContains,omitempty"`                     // OpenAPI >=3.1
	MaxContains           *uint64    `json:"maxContains,omitempty,case:strict" yaml:"maxContains,omitempty"`                     // OpenAPI >=3.1
	PatternProperties     Schemas    `json:"patternProperties,omitempty,case:strict" yaml:"patternProperties,omitempty"`         // OpenAPI >=3.1
	DependentSchemas      Schemas    `json:"dependentSchemas,omitempty,case:strict" yaml:"dependentSchemas,omitempty"`           // OpenAPI >=3.1
	PropertyNames         *SchemaRef `json:"propertyNames,omitempty,case:strict" yaml:"propertyNames,omitempty"`                 // OpenAPI >=3.1
	UnevaluatedItems      BoolSchema `json:"unevaluatedItems,omitempty,case:strict" yaml:"unevaluatedItems,omitempty"`           // OpenAPI >=3.1
	UnevaluatedProperties BoolSchema `json:"unevaluatedProperties,omitempty,case:strict" yaml:"unevaluatedProperties,omitempty"` // OpenAPI >=3.1

	If   *SchemaRef `json:"if,omitempty,case:strict" yaml:"if,omitempty"`     // OpenAPI >=3.1
	Then *SchemaRef `json:"then,omitempty,case:strict" yaml:"then,omitempty"` // OpenAPI >=3.1
	Else *SchemaRef `json:"else,omitempty,case:strict" yaml:"else,omitempty"` // OpenAPI >=3.1

	DependentRequired map[string][]string `json:"dependentRequired,omitempty,case:strict" yaml:"dependentRequired,omitempty"` // OpenAPI >=3.1

	// Always is set when the schema was written as a JSON Schema boolean
	// schema rather than an object, and says what that schema always does:
	// `true` accepts every instance, `false` accepts none. It is nil for an
	// ordinary object schema, and Validate rejects one that carries any other
	// keyword.
	//
	// Neither direction goes through a struct tag, hence the "-" on both:
	// UnmarshalJSON takes the bare boolean before the field decode, and
	// MarshalYAML emits it again ahead of the field map. OpenAPI >=3.1.
	Always *bool `json:"-" yaml:"-"`

	Defs          Schemas `json:"$defs,omitempty,case:strict" yaml:"$defs,omitempty"`       // OpenAPI >=3.1
	SchemaDialect string  `json:"$schema,omitempty,case:strict" yaml:"$schema,omitempty"`   // OpenAPI >=3.1
	Comment       string  `json:"$comment,omitempty,case:strict" yaml:"$comment,omitempty"` // OpenAPI >=3.1

	SchemaID      string `json:"$id,omitempty,case:strict" yaml:"$id,omitempty"`                       // OpenAPI >=3.1
	Anchor        string `json:"$anchor,omitempty,case:strict" yaml:"$anchor,omitempty"`               // OpenAPI >=3.1
	DynamicRef    string `json:"$dynamicRef,omitempty,case:strict" yaml:"$dynamicRef,omitempty"`       // OpenAPI >=3.1
	DynamicAnchor string `json:"$dynamicAnchor,omitempty,case:strict" yaml:"$dynamicAnchor,omitempty"` // OpenAPI >=3.1

	ContentMediaType string     `json:"contentMediaType,omitempty,case:strict" yaml:"contentMediaType,omitempty"` // OpenAPI >=3.1
	ContentEncoding  string     `json:"contentEncoding,omitempty,case:strict" yaml:"contentEncoding,omitempty"`   // OpenAPI >=3.1
	ContentSchema    *SchemaRef `json:"contentSchema,omitempty,case:strict" yaml:"contentSchema,omitempty"`       // OpenAPI >=3.1
}

// Types represents the type(s) of a schema.
//
// In OpenAPI 3.0, this is typically a single type (e.g., "string").
// In OpenAPI 3.1, it can be an array of types (e.g., ["string", "null"]).
//
// Serialization behavior:
//   - Single type: serializes as a string (e.g., "string")
//   - Multiple types: serializes as an array (e.g., ["string", "null"])
//   - Accepts both string and array formats when unmarshaling
//
// Example OpenAPI 3.0 (single type):
//
//	schema := &Schema{Type: &Types{"string"}}
//	// JSON: {"type": "string"}
//
// Example OpenAPI 3.1 (type array):
//
//	schema := &Schema{Type: &Types{"string", "null"}}
//	// JSON: {"type": ["string", "null"]}
type Types []string

// Is returns true if the schema has exactly one type and it matches the given type.
// This is useful for OpenAPI 3.0 style single-type checks.
//
// Example:
//
//	types := &Types{"string"}
//	types.Is("string")  // true
//	types.Is("number")  // false
//
//	types = &Types{"string", "null"}
//	types.Is("string")  // false (multiple types)
func (types *Types) Is(typ string) bool {
	return types != nil && len(*types) == 1 && (*types)[0] == typ
}

// Slice returns the types as a string slice.
// Returns nil if types is nil.
//
// Example:
//
//	types := &Types{"string", "null"}
//	slice := types.Slice()  // []string{"string", "null"}
func (types *Types) Slice() []string {
	if types == nil {
		return nil
	}
	return *types
}

// Includes returns true if the given type is included in the type array.
// Returns false if types is nil.
//
// Example:
//
//	types := &Types{"string", "null"}
//	types.Includes("string")  // true
//	types.Includes("null")    // true
//	types.Includes("number")  // false
func (pTypes *Types) Includes(typ string) bool {
	if pTypes == nil {
		return false
	}
	return slices.Contains(*pTypes, typ)
}

// Permits returns true if the given type is permitted.
// Returns true if types is nil (any type allowed), otherwise checks if the type is included.
//
// Example:
//
//	var nilTypes *Types
//	nilTypes.Permits("anything")  // true (nil permits everything)
//
//	types := &Types{"string"}
//	types.Permits("string")  // true
//	types.Permits("number")  // false
func (types *Types) Permits(typ string) bool {
	if types == nil {
		return true
	}
	return types.Includes(typ)
}

// IncludesNull returns true if the type array includes "null".
// This is useful for OpenAPI 3.1 where null is a first-class type.
//
// Example:
//
//	types := &Types{"string", "null"}
//	types.IncludesNull()  // true
//
//	types = &Types{"string"}
//	types.IncludesNull()  // false
func (types *Types) IncludesNull() bool {
	return types.Includes(TypeNull)
}

// IsMultiple returns true if multiple types are specified.
// This is an OpenAPI 3.1 feature that enables type arrays.
//
// Example:
//
//	types := &Types{"string"}
//	types.IsMultiple()  // false
//
//	types = &Types{"string", "null"}
//	types.IsMultiple()  // true
func (types *Types) IsMultiple() bool {
	return types != nil && len(*types) > 1
}

// IsSingle returns true if exactly one type is specified.
//
// Example:
//
//	types := &Types{"string"}
//	types.IsSingle()  // true
//
//	types = &Types{"string", "null"}
//	types.IsSingle()  // false
func (types *Types) IsSingle() bool {
	return types != nil && len(*types) == 1
}

// IsEmpty returns true if no types are specified (nil or empty array).
// When a schema has no type specified, it permits any type.
//
// Example:
//
//	var nilTypes *Types
//	nilTypes.IsEmpty()  // true
//
//	types := &Types{}
//	types.IsEmpty()  // true
//
//	types = &Types{"string"}
//	types.IsEmpty()  // false
func (types *Types) IsEmpty() bool {
	return types == nil || len(*types) == 0
}

// Clone returns a copy of the types that shares no memory with the receiver.
// Returns nil if types is nil.
//
// Example:
//
//	types := &Types{"string"}
//	clone := types.Clone()
//	clone.With("null")  // types is still {"string"}
func (types *Types) Clone() Types {
	if types == nil {
		return nil
	}
	return slices.Clone(*types)
}

// With adds the given types to the receiver, skipping those already present.
// Duplicates are removed, keeping the first occurrence.
// types must not be nil.
//
// Example:
//
//	types := &Types{"string"}
//	types.With("null", "string")  // types is now {"string", "null"}
func (types *Types) With(typs ...string) {
	*types = uniqueTypes(slices.Concat(*types, typs))
}

// Without removes the given types from the receiver.
// Duplicates are removed, keeping the first occurrence.
// Does nothing if types is nil.
//
// Example:
//
//	types := &Types{"string", "null"}
//	types.Without("null")  // types is now {"string"}
func (types *Types) Without(typs ...string) {
	if types == nil {
		return
	}
	// Cloning keeps DeleteFunc from modifying a backing array other slices may share.
	*types = slices.DeleteFunc(uniqueTypes(slices.Clone(*types)), func(typ string) bool {
		return slices.Contains(typs, typ)
	})
}

// uniqueTypes removes duplicates from ts in place, keeping the first occurrence.
func uniqueTypes(ts Types) Types {
	unique := ts[:0]
	for _, typ := range ts {
		if !slices.Contains(unique, typ) {
			unique = append(unique, typ)
		}
	}
	return unique
}

func (pTypes *Types) MarshalJSON() ([]byte, error) {
	x, err := pTypes.MarshalYAML()
	if err != nil {
		return nil, err
	}
	return json.Marshal(x)
}

func (pTypes *Types) MarshalYAML() (any, error) {
	if pTypes == nil {
		return nil, nil
	}
	types := *pTypes
	switch len(types) {
	case 0:
		return nil, nil
	case 1:
		return types[0], nil
	default:
		return []string(types), nil
	}
}

func (types *Types) UnmarshalJSON(data []byte) error {
	var strings []string
	if err := unmarshalJSON(data, &strings); err != nil {
		var s string
		if err := unmarshalJSON(data, &s); err != nil {
			return unmarshalError(err)
		}
		strings = []string{s}
	}
	*types = strings
	return nil
}

// UnmarshalYAML accepts a single type as well as a sequence of types, mirroring
// UnmarshalJSON. Swagger 2 and OpenAPI 3.0 documents always write `type` as a
// scalar, so decoding such a document into openapi2.T or openapi3.T straight
// from YAML would otherwise fail with "cannot unmarshal !!str into Types".
func (types *Types) UnmarshalYAML(unmarshal func(any) error) error {
	var single string
	if err := unmarshal(&single); err == nil {
		*types = Types{single}
		return nil
	}
	var multiple []string
	if err := unmarshal(&multiple); err != nil {
		return err
	}
	*types = Types(multiple)
	return nil
}

// BoolSchema represents a JSON Schema keyword that can be either a boolean or a schema object.
// Used for additionalProperties, unevaluatedProperties, and unevaluatedItems.
type BoolSchema struct {
	Has    *bool
	Schema *SchemaRef
}

// AdditionalProperties is a type alias for BoolSchema, kept for backward compatibility.
type AdditionalProperties = BoolSchema

// MarshalYAML returns the YAML encoding of BoolSchema.
func (bs BoolSchema) MarshalYAML() (any, error) {
	if x := bs.Has; x != nil {
		if *x {
			return true, nil
		}
		return false, nil
	}
	if x := bs.Schema; x != nil {
		return x.MarshalYAML()
	}
	return nil, nil
}

// MarshalJSON returns the JSON encoding of BoolSchema.
func (bs BoolSchema) MarshalJSON() ([]byte, error) {
	x, err := bs.MarshalYAML()
	if err != nil {
		return nil, err
	}
	return json.Marshal(x)
}

// UnmarshalJSON sets BoolSchema to a copy of data.
func (bs *BoolSchema) UnmarshalJSON(data []byte) error {
	var x any
	if err := unmarshalJSON(data, &x); err != nil {
		return unmarshalError(err)
	}
	switch y := x.(type) {
	case nil:
	case bool:
		bs.Has = &y
	case map[string]any:
		if len(y) == 0 {
			bs.Schema = &SchemaRef{Value: &Schema{}}
		} else {
			buf := new(bytes.Buffer)
			_ = json.NewEncoder(buf).Encode(y)
			if err := json.NewDecoder(buf).Decode(&bs.Schema); err != nil {
				return err
			}
		}
	default:
		return errors.New("cannot unmarshal: value must be either a schema object or a boolean")
	}
	return nil
}

// ExclusiveBound represents exclusiveMinimum/exclusiveMaximum which changed type between OpenAPI versions.
// In OpenAPI 3.0 (JSON Schema draft-04): boolean that modifies minimum/maximum
// In OpenAPI 3.1 (JSON Schema 2020-12): number representing the actual exclusive bound
type ExclusiveBound struct {
	Bool  *bool    // For OpenAPI 3.0 style (modifier for min/max)
	Value *float64 // For OpenAPI 3.1 style (actual bound value)
}

// IsSet returns true if either Bool or Value is set.
func (eb ExclusiveBound) IsSet() bool {
	return eb.Bool != nil || eb.Value != nil
}

// IsTrue returns true if the bound is set as a boolean true (OpenAPI 3.0 style).
func (eb ExclusiveBound) IsTrue() bool {
	return eb.Bool != nil && *eb.Bool
}

// MarshalYAML returns the YAML encoding of ExclusiveBound.
func (eb ExclusiveBound) MarshalYAML() (any, error) {
	if eb.Value != nil {
		return *eb.Value, nil
	}
	if eb.Bool != nil {
		return *eb.Bool, nil
	}
	return nil, nil
}

// MarshalJSON returns the JSON encoding of ExclusiveBound.
func (eb ExclusiveBound) MarshalJSON() ([]byte, error) {
	x, err := eb.MarshalYAML()
	if err != nil {
		return nil, err
	}
	if x == nil {
		return nil, nil
	}
	return json.Marshal(x)
}

// UnmarshalJSON sets ExclusiveBound to a copy of data.
func (eb *ExclusiveBound) UnmarshalJSON(data []byte) error {
	var x any
	if err := unmarshalJSON(data, &x); err != nil {
		return unmarshalError(err)
	}
	switch y := x.(type) {
	case nil:
		// nothing to do
	case bool:
		eb.Bool = &y
	case float64:
		eb.Value = &y
	default:
		return errors.New("cannot unmarshal exclusiveMinimum/exclusiveMaximum: value must be either a number or a boolean")
	}
	return nil
}

var _ jsonpointer.JSONPointable = (*Schema)(nil)

func NewSchema() *Schema {
	return &Schema{}
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
	if x := schema.Always; x != nil {
		return *x, nil
	}

	m := make(map[string]any, 61+len(schema.Extensions))
	maps.Copy(m, schema.Extensions)

	if x := schema.OneOf; len(x) != 0 {
		m["oneOf"] = x
	}
	if x := schema.AnyOf; len(x) != 0 {
		m["anyOf"] = x
	}
	if x := schema.AllOf; len(x) != 0 {
		m["allOf"] = x
	}
	if x := schema.Not; x != nil {
		m["not"] = x
	}
	if x := schema.Type; !x.IsEmpty() {
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
	if x := schema.ExclusiveMin; x.IsSet() {
		m["exclusiveMinimum"] = x
	}
	if x := schema.ExclusiveMax; x.IsSet() {
		m["exclusiveMaximum"] = x
	}
	// Properties
	if x := schema.Nullable; x {
		m["nullable"] = x
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
	if x := schema.Discriminator; x != nil {
		m["discriminator"] = x
	}

	// OpenAPI 3.1 / JSON Schema 2020-12 fields
	if x := schema.Const; x != nil {
		m["const"] = x
	}
	if x := schema.Examples; len(x) != 0 {
		m["examples"] = x
	}
	if x := schema.PrefixItems; len(x) != 0 {
		m["prefixItems"] = x
	}
	if x := schema.Contains; x != nil {
		m["contains"] = x
	}
	if x := schema.MinContains; x != nil {
		m["minContains"] = x
	}
	if x := schema.MaxContains; x != nil {
		m["maxContains"] = x
	}
	if x := schema.PatternProperties; len(x) != 0 {
		m["patternProperties"] = x
	}
	if x := schema.DependentSchemas; len(x) != 0 {
		m["dependentSchemas"] = x
	}
	if x := schema.PropertyNames; x != nil {
		m["propertyNames"] = x
	}
	if x := schema.UnevaluatedItems; x.Has != nil || x.Schema != nil {
		m["unevaluatedItems"] = &x
	}
	if x := schema.UnevaluatedProperties; x.Has != nil || x.Schema != nil {
		m["unevaluatedProperties"] = &x
	}
	if x := schema.If; x != nil {
		m["if"] = x
	}
	if x := schema.Then; x != nil {
		m["then"] = x
	}
	if x := schema.Else; x != nil {
		m["else"] = x
	}
	if x := schema.DependentRequired; len(x) != 0 {
		m["dependentRequired"] = x
	}
	if x := schema.Defs; len(x) != 0 {
		m["$defs"] = x
	}
	if x := schema.SchemaDialect; x != "" {
		m["$schema"] = x
	}
	if x := schema.Comment; x != "" {
		m["$comment"] = x
	}
	if x := schema.SchemaID; x != "" {
		m["$id"] = x
	}
	if x := schema.Anchor; x != "" {
		m["$anchor"] = x
	}
	if x := schema.DynamicRef; x != "" {
		m["$dynamicRef"] = x
	}
	if x := schema.DynamicAnchor; x != "" {
		m["$dynamicAnchor"] = x
	}
	if x := schema.ContentMediaType; x != "" {
		m["contentMediaType"] = x
	}
	if x := schema.ContentEncoding; x != "" {
		m["contentEncoding"] = x
	}
	if x := schema.ContentSchema; x != nil {
		m["contentSchema"] = x
	}

	return m, nil
}

// unmarshalBooleanSchema reports whether data is a JSON Schema boolean schema,
// and its value if so. Only a bare `true` or `false` qualifies; an object, a
// string or anything else is an ordinary schema for the caller to decode.
func unmarshalBooleanSchema(data []byte) (bool, bool) {
	var boolean bool
	if err := unmarshalJSON(data, &boolean); err != nil {
		return false, false
	}
	return boolean, true
}

// UnmarshalJSON sets Schema to a copy of data.
func (schema *Schema) UnmarshalJSON(data []byte) error {
	// JSON Schema 2020-12 allows a boolean wherever a schema is expected, so
	// every SchemaRef position can hold one. SchemaRef.UnmarshalJSON funnels
	// into here, which is why this one case covers them all.
	if boolean, ok := unmarshalBooleanSchema(data); ok {
		origin := schema.Origin
		*schema = Schema{Always: &boolean, Origin: origin}
		return nil
	}

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

	// OpenAPI 3.1 / JSON Schema 2020-12 fields
	delete(x.Extensions, "const")
	delete(x.Extensions, "examples")
	delete(x.Extensions, "prefixItems")
	delete(x.Extensions, "contains")
	delete(x.Extensions, "minContains")
	delete(x.Extensions, "maxContains")
	delete(x.Extensions, "patternProperties")
	delete(x.Extensions, "dependentSchemas")
	delete(x.Extensions, "propertyNames")
	delete(x.Extensions, "unevaluatedItems")
	delete(x.Extensions, "unevaluatedProperties")
	delete(x.Extensions, "if")
	delete(x.Extensions, "then")
	delete(x.Extensions, "else")
	delete(x.Extensions, "dependentRequired")
	delete(x.Extensions, "$defs")
	delete(x.Extensions, "$schema")
	delete(x.Extensions, "$comment")
	delete(x.Extensions, "$id")
	delete(x.Extensions, "$anchor")
	delete(x.Extensions, "$dynamicRef")
	delete(x.Extensions, "$dynamicAnchor")
	delete(x.Extensions, "contentMediaType")
	delete(x.Extensions, "contentEncoding")
	delete(x.Extensions, "contentSchema")

	if len(x.Extensions) == 0 {
		x.Extensions = nil
	}

	*schema = Schema(x)
	return nil
}

// JSONLookup implements https://pkg.go.dev/github.com/go-openapi/jsonpointer#JSONPointable
func (schema Schema) JSONLookup(token string) (any, error) {
	switch token {
	case "additionalProperties":
		if addProps := schema.AdditionalProperties.Has; addProps != nil {
			return *addProps, nil
		}
		if addProps := schema.AdditionalProperties.Schema; addProps != nil {
			if addProps.Ref != "" {
				return &Ref{Ref: addProps.Ref}, nil
			}
			return addProps.Value, nil
		}
	case "not":
		if schema.Not != nil {
			if schema.Not.Ref != "" {
				return &Ref{Ref: schema.Not.Ref}, nil
			}
			return schema.Not.Value, nil
		}
	case "items":
		if schema.Items != nil {
			if schema.Items.Ref != "" {
				return &Ref{Ref: schema.Items.Ref}, nil
			}
			return schema.Items.Value, nil
		}
	case "oneOf":
		return schema.OneOf, nil
	case "anyOf":
		return schema.AnyOf, nil
	case "allOf":
		return schema.AllOf, nil
	case "type":
		return schema.Type, nil
	case "title":
		return schema.Title, nil
	case "format":
		return schema.Format, nil
	case "description":
		return schema.Description, nil
	case "enum":
		return schema.Enum, nil
	case "default":
		return schema.Default, nil
	case "example":
		return schema.Example, nil
	case "externalDocs":
		return schema.ExternalDocs, nil
	case "uniqueItems":
		return schema.UniqueItems, nil
	case "exclusiveMin":
		return schema.ExclusiveMin, nil
	case "exclusiveMax":
		return schema.ExclusiveMax, nil
	case "nullable":
		return schema.Nullable, nil
	case "readOnly":
		return schema.ReadOnly, nil
	case "writeOnly":
		return schema.WriteOnly, nil
	case "allowEmptyValue":
		return schema.AllowEmptyValue, nil
	case "xml":
		return schema.XML, nil
	case "deprecated":
		return schema.Deprecated, nil
	case "min":
		return schema.Min, nil
	case "max":
		return schema.Max, nil
	case "multipleOf":
		return schema.MultipleOf, nil
	case "minLength":
		return schema.MinLength, nil
	case "maxLength":
		return schema.MaxLength, nil
	case "pattern":
		return schema.Pattern, nil
	case "minItems":
		return schema.MinItems, nil
	case "maxItems":
		return schema.MaxItems, nil
	case "required":
		return schema.Required, nil
	case "properties":
		return schema.Properties, nil
	case "minProps":
		return schema.MinProps, nil
	case "maxProps":
		return schema.MaxProps, nil
	case "discriminator":
		return schema.Discriminator, nil

	// OpenAPI 3.1 / JSON Schema 2020-12 fields
	case "const":
		return schema.Const, nil
	case "examples":
		return schema.Examples, nil
	case "prefixItems":
		return schema.PrefixItems, nil
	case "contains":
		if schema.Contains != nil {
			if schema.Contains.Ref != "" {
				return &Ref{Ref: schema.Contains.Ref}, nil
			}
			return schema.Contains.Value, nil
		}
	case "minContains":
		return schema.MinContains, nil
	case "maxContains":
		return schema.MaxContains, nil
	case "patternProperties":
		return schema.PatternProperties, nil
	case "dependentSchemas":
		return schema.DependentSchemas, nil
	case "propertyNames":
		if schema.PropertyNames != nil {
			if schema.PropertyNames.Ref != "" {
				return &Ref{Ref: schema.PropertyNames.Ref}, nil
			}
			return schema.PropertyNames.Value, nil
		}
	case "unevaluatedItems":
		if ui := schema.UnevaluatedItems.Has; ui != nil {
			return *ui, nil
		}
		if ui := schema.UnevaluatedItems.Schema; ui != nil {
			if ui.Ref != "" {
				return &Ref{Ref: ui.Ref}, nil
			}
			return ui.Value, nil
		}
	case "unevaluatedProperties":
		if up := schema.UnevaluatedProperties.Has; up != nil {
			return *up, nil
		}
		if up := schema.UnevaluatedProperties.Schema; up != nil {
			if up.Ref != "" {
				return &Ref{Ref: up.Ref}, nil
			}
			return up.Value, nil
		}
	case "if":
		if schema.If != nil {
			if schema.If.Ref != "" {
				return &Ref{Ref: schema.If.Ref}, nil
			}
			return schema.If.Value, nil
		}
	case "then":
		if schema.Then != nil {
			if schema.Then.Ref != "" {
				return &Ref{Ref: schema.Then.Ref}, nil
			}
			return schema.Then.Value, nil
		}
	case "else":
		if schema.Else != nil {
			if schema.Else.Ref != "" {
				return &Ref{Ref: schema.Else.Ref}, nil
			}
			return schema.Else.Value, nil
		}
	case "dependentRequired":
		return schema.DependentRequired, nil
	case "$defs":
		return schema.Defs, nil
	case "$schema":
		return schema.SchemaDialect, nil
	case "$comment":
		return schema.Comment, nil
	case "$id":
		return schema.SchemaID, nil
	case "$anchor":
		return schema.Anchor, nil
	case "$dynamicRef":
		return schema.DynamicRef, nil
	case "$dynamicAnchor":
		return schema.DynamicAnchor, nil
	case "contentMediaType":
		return schema.ContentMediaType, nil
	case "contentEncoding":
		return schema.ContentEncoding, nil
	case "contentSchema":
		if schema.ContentSchema != nil {
			if schema.ContentSchema.Ref != "" {
				return &Ref{Ref: schema.ContentSchema.Ref}, nil
			}
			return schema.ContentSchema.Value, nil
		}
	}

	v, _, err := jsonpointer.GetForToken(schema.Extensions, token)
	return v, err
}

func (schema *Schema) NewRef() *SchemaRef {
	return &SchemaRef{
		Value: schema,
	}
}

func NewOneOfSchema(schemas ...*Schema) *Schema {
	refs := make([]*SchemaRef, 0, len(schemas))
	for _, schema := range schemas {
		refs = append(refs, &SchemaRef{Value: schema})
	}
	return &Schema{
		OneOf: refs,
	}
}

func NewAnyOfSchema(schemas ...*Schema) *Schema {
	refs := make([]*SchemaRef, 0, len(schemas))
	for _, schema := range schemas {
		refs = append(refs, &SchemaRef{Value: schema})
	}
	return &Schema{
		AnyOf: refs,
	}
}

func NewAllOfSchema(schemas ...*Schema) *Schema {
	refs := make([]*SchemaRef, 0, len(schemas))
	for _, schema := range schemas {
		refs = append(refs, &SchemaRef{Value: schema})
	}
	return &Schema{
		AllOf: refs,
	}
}

func NewBoolSchema() *Schema {
	return &Schema{
		Type: &Types{TypeBoolean},
	}
}

func NewFloat64Schema() *Schema {
	return &Schema{
		Type: &Types{TypeNumber},
	}
}

func NewIntegerSchema() *Schema {
	return &Schema{
		Type: &Types{TypeInteger},
	}
}

func NewInt32Schema() *Schema {
	return &Schema{
		Type:   &Types{TypeInteger},
		Format: "int32",
	}
}

func NewInt64Schema() *Schema {
	return &Schema{
		Type:   &Types{TypeInteger},
		Format: "int64",
	}
}

func NewStringSchema() *Schema {
	return &Schema{
		Type: &Types{TypeString},
	}
}

func NewDateTimeSchema() *Schema {
	return &Schema{
		Type:   &Types{TypeString},
		Format: "date-time",
	}
}

func NewUUIDSchema() *Schema {
	return &Schema{
		Type:   &Types{TypeString},
		Format: "uuid",
	}
}

func NewBytesSchema() *Schema {
	return &Schema{
		Type:   &Types{TypeString},
		Format: "byte",
	}
}

func NewArraySchema() *Schema {
	return &Schema{
		Type: &Types{TypeArray},
	}
}

func NewObjectSchema() *Schema {
	return &Schema{
		Type:       &Types{TypeObject},
		Properties: make(Schemas),
	}
}

// WithNullable set Nullable (OpenAPI 3.0)
// For 3.1+ append TypeNull to Types.
func (schema *Schema) WithNullable() *Schema {
	schema.Nullable = true
	return schema
}

// WithTypes adds the given types to the schema's Type, initializing it if nil.
// See Types.With.
func (schema *Schema) WithTypes(typs ...string) *Schema {
	if schema.Type == nil {
		schema.Type = &Types{}
	}
	schema.Type.With(typs...)
	return schema
}

func (schema *Schema) WithMin(value float64) *Schema {
	schema.Min = &value
	return schema
}

func (schema *Schema) WithMax(value float64) *Schema {
	schema.Max = &value
	return schema
}

// WithExclusiveMin sets exclusiveMinimum as a boolean (OpenAPI 3.0 style).
func (schema *Schema) WithExclusiveMin(value bool) *Schema {
	schema.ExclusiveMin = ExclusiveBound{Bool: &value}
	return schema
}

// WithExclusiveMax sets exclusiveMaximum as a boolean (OpenAPI 3.0 style).
func (schema *Schema) WithExclusiveMax(value bool) *Schema {
	schema.ExclusiveMax = ExclusiveBound{Bool: &value}
	return schema
}

// WithExclusiveMinValue sets exclusiveMinimum as a number (OpenAPI 3.1 style).
func (schema *Schema) WithExclusiveMinValue(value float64) *Schema {
	schema.ExclusiveMin = ExclusiveBound{Value: &value}
	return schema
}

// WithExclusiveMaxValue sets exclusiveMaximum as a number (OpenAPI 3.1 style).
func (schema *Schema) WithExclusiveMaxValue(value float64) *Schema {
	schema.ExclusiveMax = ExclusiveBound{Value: &value}
	return schema
}

func (schema *Schema) WithEnum(values ...any) *Schema {
	schema.Enum = values
	return schema
}

func (schema *Schema) WithDefault(defaultValue any) *Schema {
	schema.Default = defaultValue
	return schema
}

func (schema *Schema) WithFormat(value string) *Schema {
	schema.Format = value
	return schema
}

func (schema *Schema) WithLength(i int64) *Schema {
	n := uint64(i)
	schema.MinLength = n
	schema.MaxLength = &n
	return schema
}

func (schema *Schema) WithMinLength(i int64) *Schema {
	n := uint64(i)
	schema.MinLength = n
	return schema
}

func (schema *Schema) WithMaxLength(i int64) *Schema {
	n := uint64(i)
	schema.MaxLength = &n
	return schema
}

func (schema *Schema) WithLengthDecodedBase64(i int64) *Schema {
	n := uint64(i)
	v := (n*8 + 5) / 6
	schema.MinLength = v
	schema.MaxLength = &v
	return schema
}

func (schema *Schema) WithMinLengthDecodedBase64(i int64) *Schema {
	n := uint64(i)
	schema.MinLength = (n*8 + 5) / 6
	return schema
}

func (schema *Schema) WithMaxLengthDecodedBase64(i int64) *Schema {
	n := uint64(i)
	schema.MinLength = (n*8 + 5) / 6
	return schema
}

func (schema *Schema) WithPattern(pattern string) *Schema {
	schema.Pattern = pattern
	return schema
}

func (schema *Schema) WithItems(value *Schema) *Schema {
	schema.Items = &SchemaRef{
		Value: value,
	}
	return schema
}

func (schema *Schema) WithMinItems(i int64) *Schema {
	n := uint64(i)
	schema.MinItems = n
	return schema
}

func (schema *Schema) WithMaxItems(i int64) *Schema {
	n := uint64(i)
	schema.MaxItems = &n
	return schema
}

func (schema *Schema) WithUniqueItems(unique bool) *Schema {
	schema.UniqueItems = unique
	return schema
}

func (schema *Schema) WithProperty(name string, propertySchema *Schema) *Schema {
	return schema.WithPropertyRef(name, &SchemaRef{
		Value: propertySchema,
	})
}

func (schema *Schema) WithPropertyRef(name string, ref *SchemaRef) *Schema {
	if schema.Properties == nil {
		schema.Properties = make(Schemas)
	}
	schema.Properties[name] = ref
	return schema
}

func (schema *Schema) WithProperties(properties map[string]*Schema) *Schema {
	result := make(Schemas, len(properties))
	for k, v := range properties {
		result[k] = &SchemaRef{
			Value: v,
		}
	}
	schema.Properties = result
	return schema
}

func (schema *Schema) WithRequired(required []string) *Schema {
	schema.Required = required
	return schema
}

func (schema *Schema) WithMinProperties(i int64) *Schema {
	n := uint64(i)
	schema.MinProps = n
	return schema
}

func (schema *Schema) WithMaxProperties(i int64) *Schema {
	n := uint64(i)
	schema.MaxProps = &n
	return schema
}

func (schema *Schema) WithAnyAdditionalProperties() *Schema {
	schema.AdditionalProperties = AdditionalProperties{Has: Ptr(true)}
	return schema
}

func (schema *Schema) WithoutAdditionalProperties() *Schema {
	schema.AdditionalProperties = AdditionalProperties{Has: Ptr(false)}
	return schema
}

func (schema *Schema) WithAdditionalProperties(v *Schema) *Schema {
	schema.AdditionalProperties = AdditionalProperties{}
	if v != nil {
		schema.AdditionalProperties.Schema = &SchemaRef{Value: v}
	}
	return schema
}

// PermitsNull returns true when a schema is Nullable (only in OpenAPI 3.0)
// or when "null" is among the allowed types.
// In OpenAPI 3.1+, nullable is an unknown keyword.
//
// Usage: schema.PermitsNull(EnableJSONSchema2020()) for 3.1+
func (schema *Schema) PermitsNull(opts ...SchemaValidationOption) bool {
	settings := newSchemaValidationSettings(opts...)
	return schema.permitsNull(settings)
}

func (schema *Schema) permitsNull(settings *schemaValidationSettings) bool {
	if settings.useJSONSchema2020 {
		return schema.Type.IncludesNull()
	}
	return schema.Nullable || schema.Type.IncludesNull()
}

// IsEmpty tells whether schema is equivalent to the empty schema `{}`.
func (schema *Schema) IsEmpty() bool {
	if schema.Type != nil || schema.Format != "" || len(schema.Enum) != 0 ||
		schema.UniqueItems || schema.ExclusiveMin.IsSet() || schema.ExclusiveMax.IsSet() ||
		schema.Nullable || schema.ReadOnly || schema.WriteOnly || schema.AllowEmptyValue ||
		schema.Min != nil || schema.Max != nil || schema.MultipleOf != nil ||
		schema.MinLength != 0 || schema.MaxLength != nil || schema.Pattern != "" ||
		schema.MinItems != 0 || schema.MaxItems != nil ||
		len(schema.Required) != 0 ||
		schema.MinProps != 0 || schema.MaxProps != nil ||
		schema.Const != nil {
		return false
	}
	if n := schema.Not; n != nil && n.Value != nil && !n.Value.IsEmpty() {
		return false
	}
	if ap := schema.AdditionalProperties.Schema; ap != nil && ap.Value != nil && !ap.Value.IsEmpty() {
		return false
	}
	if apa := schema.AdditionalProperties.Has; apa != nil && !*apa {
		return false
	}
	if items := schema.Items; items != nil && items.Value != nil && !items.Value.IsEmpty() {
		return false
	}
	for _, s := range schema.PrefixItems {
		if ss := s.Value; ss != nil && !ss.IsEmpty() {
			return false
		}
	}
	if c := schema.Contains; c != nil && c.Value != nil && !c.Value.IsEmpty() {
		return false
	}
	if schema.MinContains != nil || schema.MaxContains != nil {
		return false
	}
	for _, s := range schema.Properties {
		if ss := s.Value; ss != nil && !ss.IsEmpty() {
			return false
		}
	}
	for _, s := range schema.PatternProperties {
		if ss := s.Value; ss != nil && !ss.IsEmpty() {
			return false
		}
	}
	for _, s := range schema.DependentSchemas {
		if ss := s.Value; ss != nil && !ss.IsEmpty() {
			return false
		}
	}
	if pn := schema.PropertyNames; pn != nil && pn.Value != nil && !pn.Value.IsEmpty() {
		return false
	}
	if ui := schema.UnevaluatedItems.Schema; ui != nil && ui.Value != nil && !ui.Value.IsEmpty() {
		return false
	}
	if uih := schema.UnevaluatedItems.Has; uih != nil && !*uih {
		return false
	}
	if up := schema.UnevaluatedProperties.Schema; up != nil && up.Value != nil && !up.Value.IsEmpty() {
		return false
	}
	if uph := schema.UnevaluatedProperties.Has; uph != nil && !*uph {
		return false
	}
	if len(schema.Examples) != 0 {
		return false
	}
	for _, s := range schema.OneOf {
		if ss := s.Value; ss != nil && !ss.IsEmpty() {
			return false
		}
	}
	for _, s := range schema.AnyOf {
		if ss := s.Value; ss != nil && !ss.IsEmpty() {
			return false
		}
	}
	for _, s := range schema.AllOf {
		if ss := s.Value; ss != nil && !ss.IsEmpty() {
			return false
		}
	}
	if f := schema.If; f != nil && f.Value != nil && !f.Value.IsEmpty() {
		return false
	}
	if t := schema.Then; t != nil && t.Value != nil && !t.Value.IsEmpty() {
		return false
	}
	if e := schema.Else; e != nil && e.Value != nil && !e.Value.IsEmpty() {
		return false
	}
	if len(schema.DependentRequired) != 0 {
		return false
	}
	if len(schema.Defs) != 0 {
		return false
	}
	if schema.SchemaDialect != "" || schema.Comment != "" {
		return false
	}
	if schema.SchemaID != "" || schema.Anchor != "" || schema.DynamicRef != "" || schema.DynamicAnchor != "" {
		return false
	}
	if schema.ContentMediaType != "" || schema.ContentEncoding != "" {
		return false
	}
	if cs := schema.ContentSchema; cs != nil && cs.Value != nil && !cs.Value.IsEmpty() {
		return false
	}
	// Last, so a schema carrying both a boolean and other keywords is judged by
	// the keywords. Neither boolean is empty: `false` rejects every instance,
	// and `true` is accepted by visitJSON explicitly rather than by skipping
	// validation, which is what leaves `not: true` free to reject.
	if schema.Always != nil {
		return false
	}

	return true
}

// Validate returns an error if Schema does not comply with the OpenAPI spec.
func (schema *Schema) Validate(ctx context.Context, opts ...ValidationOption) error {
	// Apply document-level validation options to the context
	ctx = WithValidationOptions(ctx, opts...)

	// Perform schema validation with the options in context
	_, err := schema.validate(ctx, make(map[*Schema]struct{}))
	return err
}

// returns the updated set of visited schemas and an error if Schema does not comply with the OpenAPI spec.
// Cycles and memoization is achieved with `visited`.
func (schema *Schema) validate(ctx context.Context, visited map[*Schema]struct{}) (map[*Schema]struct{}, error) {
	if _, ok := visited[schema]; ok {
		return visited, nil
	}

	validationOpts := getValidationOptions(ctx)

	visited[schema] = struct{}{}

	// A boolean schema carries no other keyword. One that does would marshal
	// back as the bare boolean and silently drop the rest, so reject it rather
	// than let a hand-built value lose content on the next serialization.
	if schema.Always != nil {
		// A boolean schema is JSON Schema 2020-12, so 3.0 has no such thing:
		// there a schema MUST be a Schema Object.
		if !validationOpts.isOpenAPI31OrLater {
			return visited, newBooleanSchemaFor31Plus(schema.Origin)
		}
		rest := *schema
		rest.Always = nil
		if !rest.IsEmpty() || len(schema.Extensions) != 0 {
			return visited, newSchemaBooleanFieldsExclusive(schema.Origin)
		}
		return visited, nil
	}

	// OpenAPI 3.0 forbids both. 3.1 takes readOnly and writeOnly from JSON
	// Schema 2020-12, which does not.
	if schema.ReadOnly && schema.WriteOnly && !validationOpts.isOpenAPI31OrLater {
		return visited, newSchemaReadOnlyWriteOnlyExclusive(schema.Origin)
	}

	// The elements of `required` MUST be unique (JSON Schema 2020-12 §6.5.3
	// for OAS 3.1, draft-04 for OAS 3.0).
	if len(schema.Required) > 1 {
		seen := make(map[string]struct{}, len(schema.Required))
		for _, name := range schema.Required {
			if _, dup := seen[name]; dup {
				return visited, &DuplicateRequiredFieldError{Field: name, Origin: schema.Origin}
			}
			seen[name] = struct{}{}
		}
	}

	// Reject fields that only exist in OAS 3.1 / JSON Schema 2020-12 when the
	// document is OAS 3.0. Fields explicitly allowed via AllowExtraSiblingFields
	// are skipped (opt-in escape hatch for 3.0 docs that reference external
	// JSON Schema documents). Once 3.0 moves to its own schema validator this
	// block becomes a no-op.
	if !validationOpts.isOpenAPI31OrLater {
		allowed := validationOpts.extraSiblingFieldsAllowed
		reject := func(field string) error {
			if _, ok := allowed[field]; ok {
				return nil
			}
			return errFieldFor31Plus(field, schema.Origin)
		}
		if schema.Type.IsMultiple() { // In 3.0 but only as a single string
			return visited, newTypeArrayFor31Plus(schema.Origin)
		}
		if schema.ExclusiveMin.Value != nil { // In 3.0 but only as bool
			return visited, newExclusiveBoundNumberFor31Plus("exclusiveMinimum", schema.Origin)
		}
		if schema.ExclusiveMax.Value != nil { // In 3.0 but only as bool
			return visited, newExclusiveBoundNumberFor31Plus("exclusiveMaximum", schema.Origin)
		}
		if schema.Const != nil {
			if err := reject("const"); err != nil {
				return visited, err
			}
		}
		if len(schema.Examples) != 0 {
			if err := reject("examples"); err != nil {
				return visited, err
			}
		}
		if len(schema.PrefixItems) != 0 {
			if err := reject("prefixItems"); err != nil {
				return visited, err
			}
		}
		if schema.Contains != nil {
			if err := reject("contains"); err != nil {
				return visited, err
			}
		}
		if schema.MinContains != nil {
			if err := reject("minContains"); err != nil {
				return visited, err
			}
		}
		if schema.MaxContains != nil {
			if err := reject("maxContains"); err != nil {
				return visited, err
			}
		}
		if len(schema.PatternProperties) != 0 {
			if err := reject("patternProperties"); err != nil {
				return visited, err
			}
		}
		if len(schema.DependentSchemas) != 0 {
			if err := reject("dependentSchemas"); err != nil {
				return visited, err
			}
		}
		if schema.PropertyNames != nil {
			if err := reject("propertyNames"); err != nil {
				return visited, err
			}
		}
		if schema.UnevaluatedItems.Has != nil || schema.UnevaluatedItems.Schema != nil {
			if err := reject("unevaluatedItems"); err != nil {
				return visited, err
			}
		}
		if schema.UnevaluatedProperties.Has != nil || schema.UnevaluatedProperties.Schema != nil {
			if err := reject("unevaluatedProperties"); err != nil {
				return visited, err
			}
		}
		if schema.If != nil {
			if err := reject("if"); err != nil {
				return visited, err
			}
		}
		if schema.Then != nil {
			if err := reject("then"); err != nil {
				return visited, err
			}
		}
		if schema.Else != nil {
			if err := reject("else"); err != nil {
				return visited, err
			}
		}
		if len(schema.DependentRequired) != 0 {
			if err := reject("dependentRequired"); err != nil {
				return visited, err
			}
		}
		if len(schema.Defs) != 0 {
			if err := reject("$defs"); err != nil {
				return visited, err
			}
		}
		if schema.SchemaDialect != "" {
			if err := reject("$schema"); err != nil {
				return visited, err
			}
		}
		if schema.Comment != "" {
			if err := reject("$comment"); err != nil {
				return visited, err
			}
		}
		if schema.SchemaID != "" {
			if err := reject("$id"); err != nil {
				return visited, err
			}
		}
		if schema.Anchor != "" {
			if err := reject("$anchor"); err != nil {
				return visited, err
			}
		}
		if schema.DynamicRef != "" {
			if err := reject("$dynamicRef"); err != nil {
				return visited, err
			}
		}
		if schema.DynamicAnchor != "" {
			if err := reject("$dynamicAnchor"); err != nil {
				return visited, err
			}
		}
		if schema.ContentMediaType != "" {
			if err := reject("contentMediaType"); err != nil {
				return visited, err
			}
		}
		if schema.ContentEncoding != "" {
			if err := reject("contentEncoding"); err != nil {
				return visited, err
			}
		}
		if schema.ContentSchema != nil {
			if err := reject("contentSchema"); err != nil {
				return visited, err
			}
		}
	}

	// Reject the OAS 3.0 forms that 3.1 / JSON Schema 2020-12 removed. (nullable
	// is not rejected: since 3.1 it is an unknown keyword, which JSON Schema ignores.)
	if validationOpts.isOpenAPI31OrLater {
		if schema.ExclusiveMin.Bool != nil {
			return visited, newExclusiveBoundBooleanBefore31("exclusiveMinimum", schema.Origin)
		}
		if schema.ExclusiveMax.Bool != nil {
			return visited, newExclusiveBoundBooleanBefore31("exclusiveMaximum", schema.Origin)
		}
	}

	for _, item := range schema.OneOf {
		v := item.Value
		if v == nil {
			return visited, newUnresolvedRef(item.Ref, item.Origin)
		}

		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, &SchemaCombinatorElementValidationError{Combinator: "oneOf", Cause: err}
		}
	}

	for _, item := range schema.AnyOf {
		v := item.Value
		if v == nil {
			return visited, newUnresolvedRef(item.Ref, item.Origin)
		}

		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, &SchemaCombinatorElementValidationError{Combinator: "anyOf", Cause: err}
		}
	}

	for _, item := range schema.AllOf {
		v := item.Value
		if v == nil {
			return visited, newUnresolvedRef(item.Ref, item.Origin)
		}

		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, &SchemaCombinatorElementValidationError{Combinator: "allOf", Cause: err}
		}
	}

	if ref := schema.Not; ref != nil {
		v := ref.Value
		if v == nil {
			return visited, newUnresolvedRef(ref.Ref, ref.Origin)
		}

		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, err
		}
	}

	if ref := schema.If; ref != nil {
		v := ref.Value
		if v == nil {
			return visited, newUnresolvedRef(ref.Ref, ref.Origin)
		}
		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, err
		}
	}
	if ref := schema.Then; ref != nil {
		v := ref.Value
		if v == nil {
			return visited, newUnresolvedRef(ref.Ref, ref.Origin)
		}
		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, err
		}
	}
	if ref := schema.Else; ref != nil {
		v := ref.Value
		if v == nil {
			return visited, newUnresolvedRef(ref.Ref, ref.Origin)
		}
		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, err
		}
	}

	for _, schemaType := range schema.Type.Slice() {
		switch schemaType {
		case TypeBoolean:
		case TypeNumber:
			if format := schema.Format; len(format) > 0 {
				switch format {
				case "float", "double":
				default:
					if _, ok := SchemaNumberFormats[format]; !ok && validationOpts.schemaFormatValidationEnabled {
						return visited, unsupportedFormat(format)
					}
				}
			}
		case TypeInteger:
			if format := schema.Format; len(format) > 0 {
				switch format {
				case "int32", "int64":
				default:
					if _, ok := SchemaIntegerFormats[format]; !ok && validationOpts.schemaFormatValidationEnabled {
						return visited, unsupportedFormat(format)
					}
				}
			}
		case TypeString:
			if format := schema.Format; len(format) > 0 {
				switch format {
				// Supported by OpenAPIv3.0.3:
				// https://spec.openapis.org/oas/v3.0.3
				case "byte", "binary", "date", "date-time", "password":
				// In JSON Draft-07 (not validated yet though):
				// https://json-schema.org/draft-07/json-schema-release-notes.html#formats
				case "iri", "iri-reference", "uri-template", "idn-email", "idn-hostname":
				case "json-pointer", "relative-json-pointer", "regex", "time":
				// In JSON Draft 2019-09 (not validated yet though):
				// https://json-schema.org/draft/2019-09/release-notes.html#format-vocabulary
				case "duration", "uuid":
				// Defined in some other specification
				case "email", "hostname", "ipv4", "ipv6", "uri", "uri-reference":
				default:
					if _, ok := SchemaStringFormats[format]; !ok && validationOpts.schemaFormatValidationEnabled {
						return visited, unsupportedFormat(format)
					}
				}
			}
			if !validationOpts.schemaPatternValidationDisabled && schema.Pattern != "" {
				// Only successfully compiled patterns are cached.
				// The cache does not record which compiler was used, so a custom one always compiles.
				if _, ok := compiledPatterns.Load(schema.Pattern); !ok || validationOpts.regexCompilerFunc != nil {
					if _, err := schema.compilePattern(nil, validationOpts.regexCompilerFunc); err != nil {
						return visited, err
					}
				}
			}
		case TypeArray:
			if schema.Items == nil && !validationOpts.jsonSchema2020ValidationEnabled && len(schema.PrefixItems) == 0 {
				return visited, newSchemaItemsRequired(schema.Origin)
			}
		case TypeObject:
		case TypeNull:
			if !validationOpts.jsonSchema2020ValidationEnabled {
				return visited, newSchemaTypeError(schemaType, schema.Origin)
			}
		default:
			return visited, newSchemaTypeError(schemaType, schema.Origin)
		}
	}

	if ref := schema.Items; ref != nil {
		if err := ref.validateExtras(ctx); err != nil {
			return visited, err
		}
		v := ref.Value
		if v == nil {
			return visited, newUnresolvedRef(ref.Ref, ref.Origin)
		}

		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, err
		}
	}

	for _, name := range componentNames(schema.Properties) {
		ref := schema.Properties[name]
		if err := ref.validateExtras(ctx); err != nil {
			return visited, err
		}
		v := ref.Value
		if v == nil {
			return visited, newUnresolvedRef(ref.Ref, ref.Origin)
		}

		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, err
		}
	}

	if schema.AdditionalProperties.Has != nil && schema.AdditionalProperties.Schema != nil {
		return visited, newSchemaAdditionalPropertiesBothForms(schema.Origin)
	}
	if ref := schema.AdditionalProperties.Schema; ref != nil {
		if err := ref.validateExtras(ctx); err != nil {
			return visited, err
		}
		v := ref.Value
		if v == nil {
			return visited, newUnresolvedRef(ref.Ref, ref.Origin)
		}

		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, err
		}
	}

	// OpenAPI 3.1 / JSON Schema 2020-12 sub-schemas
	for _, ref := range schema.PrefixItems {
		if err := ref.validateExtras(ctx); err != nil {
			return visited, err
		}
		v := ref.Value
		if v == nil {
			return visited, newUnresolvedRef(ref.Ref, ref.Origin)
		}

		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, err
		}
	}
	if ref := schema.Contains; ref != nil {
		if err := ref.validateExtras(ctx); err != nil {
			return visited, err
		}
		v := ref.Value
		if v == nil {
			return visited, newUnresolvedRef(ref.Ref, ref.Origin)
		}

		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, err
		}
	}
	for _, name := range componentNames(schema.PatternProperties) {
		ref := schema.PatternProperties[name]
		if err := ref.validateExtras(ctx); err != nil {
			return visited, err
		}
		v := ref.Value
		if v == nil {
			return visited, newUnresolvedRef(ref.Ref, ref.Origin)
		}

		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, err
		}
	}
	for _, name := range componentNames(schema.DependentSchemas) {
		ref := schema.DependentSchemas[name]
		if err := ref.validateExtras(ctx); err != nil {
			return visited, err
		}
		v := ref.Value
		if v == nil {
			return visited, newUnresolvedRef(ref.Ref, ref.Origin)
		}

		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, err
		}
	}
	for _, name := range componentNames(schema.Defs) {
		ref := schema.Defs[name]
		if err := ref.validateExtras(ctx); err != nil {
			return visited, err
		}
		v := ref.Value
		if v == nil {
			return visited, newUnresolvedRef(ref.Ref, ref.Origin)
		}

		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, err
		}
	}
	if ref := schema.PropertyNames; ref != nil {
		if err := ref.validateExtras(ctx); err != nil {
			return visited, err
		}
		v := ref.Value
		if v == nil {
			return visited, newUnresolvedRef(ref.Ref, ref.Origin)
		}

		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, err
		}
	}
	if schema.UnevaluatedItems.Has != nil && schema.UnevaluatedItems.Schema != nil {
		return visited, newSchemaUnevaluatedItemsBothForms(schema.Origin)
	}
	if ref := schema.UnevaluatedItems.Schema; ref != nil {
		if err := ref.validateExtras(ctx); err != nil {
			return visited, err
		}
		v := ref.Value
		if v == nil {
			return visited, newUnresolvedRef(ref.Ref, ref.Origin)
		}

		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, err
		}
	}
	if schema.UnevaluatedProperties.Has != nil && schema.UnevaluatedProperties.Schema != nil {
		return visited, newSchemaUnevaluatedPropertiesBothForms(schema.Origin)
	}
	if ref := schema.UnevaluatedProperties.Schema; ref != nil {
		if err := ref.validateExtras(ctx); err != nil {
			return visited, err
		}
		v := ref.Value
		if v == nil {
			return visited, newUnresolvedRef(ref.Ref, ref.Origin)
		}

		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, err
		}
	}
	if ref := schema.ContentSchema; ref != nil {
		if err := ref.validateExtras(ctx); err != nil {
			return visited, err
		}
		v := ref.Value
		if v == nil {
			return visited, newUnresolvedRef(ref.Ref, ref.Origin)
		}

		var err error
		if visited, err = v.validate(ctx, visited); err != nil {
			return visited, err
		}
	}

	if v := schema.ExternalDocs; v != nil {
		if err := v.Validate(ctx); err != nil {
			return visited, &SectionValidationError{Section: "external docs", Cause: err}
		}
	}

	if v := schema.Default; v != nil && !validationOpts.schemaDefaultsValidationDisabled {
		if err := validateExampleValue(ctx, v, schema); err != nil {
			return visited, newSchemaValueError("default", err, schema.Origin)
		}
	}

	if x := schema.Example; x != nil && !validationOpts.examplesValidationDisabled {
		if err := validateExampleValue(ctx, x, schema); err != nil {
			return visited, newSchemaValueError("example", err, schema.Origin)
		}
	}

	return visited, validateExtensions(ctx, schema.Extensions, schema.Origin)
}

// VisitJSON applies a Schema to the given data, considering opts.
// To validate data against an OpenAPIv3.1+ schema, be sure to pass the EnableJSONSchema2020() option.
// The schema must have passed Validate (for the document's OpenAPI version) first:
// VisitJSON's behaviour on an invalid schema is undefined.
func (schema *Schema) VisitJSON(value any, opts ...SchemaValidationOption) error {
	settings := newSchemaValidationSettings(opts...)

	if settings.useJSONSchema2020 {
		if settings.schemaErrorDetailsDisabled {
			return ErrSchemaErrorDetailsWithJSONSchema2020
		}
		return schema.useJSONSchema2020(settings, value)
	}
	return schema.visitJSON(settings, value)
}
