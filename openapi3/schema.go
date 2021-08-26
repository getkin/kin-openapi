package openapi3

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/getkin/kin-openapi/jsoninfo"
	"github.com/go-openapi/jsonpointer"
)

// Float64Ptr is a helper for defining OpenAPI schemas.
func Float64Ptr(value float64) *float64 {
	return &value
}

// BoolPtr is a helper for defining OpenAPI schemas.
func BoolPtr(value bool) *bool {
	return &value
}

// Int64Ptr is a helper for defining OpenAPI schemas.
func Int64Ptr(value int64) *int64 {
	return &value
}

// Uint64Ptr is a helper for defining OpenAPI schemas.
func Uint64Ptr(value uint64) *uint64 {
	return &value
}

type Schemas map[string]*SchemaRef

var _ jsonpointer.JSONPointable = (*Schemas)(nil)

func (s Schemas) JSONLookup(token string) (interface{}, error) {
	ref, ok := s[token]
	if ref == nil || ok == false {
		return nil, fmt.Errorf("object has no field %q", token)
	}

	if ref.Ref != "" {
		return &Ref{Ref: ref.Ref}, nil
	}
	return ref.Value, nil
}

type SchemaRefs []*SchemaRef

var _ jsonpointer.JSONPointable = (*SchemaRefs)(nil)

func (s SchemaRefs) JSONLookup(token string) (interface{}, error) {
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
type Schema struct {
	ExtensionProps

	OneOf        SchemaRefs    `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf        SchemaRefs    `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	AllOf        SchemaRefs    `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	Not          *SchemaRef    `json:"not,omitempty" yaml:"not,omitempty"`
	Type         string        `json:"type,omitempty" yaml:"type,omitempty"`
	Title        string        `json:"title,omitempty" yaml:"title,omitempty"`
	Format       string        `json:"format,omitempty" yaml:"format,omitempty"`
	Description  string        `json:"description,omitempty" yaml:"description,omitempty"`
	Enum         []interface{} `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default      interface{}   `json:"default,omitempty" yaml:"default,omitempty"`
	Example      interface{}   `json:"example,omitempty" yaml:"example,omitempty"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`

	// Array-related, here for struct compactness
	UniqueItems bool `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	// Number-related, here for struct compactness
	ExclusiveMin bool `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	ExclusiveMax bool `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	// Properties
	Nullable        bool        `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	ReadOnly        bool        `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly       bool        `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
	AllowEmptyValue bool        `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	XML             interface{} `json:"xml,omitempty" yaml:"xml,omitempty"`
	Deprecated      bool        `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`

	// Number
	Min        *float64 `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Max        *float64 `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	MultipleOf *float64 `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`

	// String
	MinLength       uint64  `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength       *uint64 `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	Pattern         string  `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	compiledPattern *regexp.Regexp

	// Array
	MinItems uint64     `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	MaxItems *uint64    `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	Items    *SchemaRef `json:"items,omitempty" yaml:"items,omitempty"`

	// Object
	Required                    []string       `json:"required,omitempty" yaml:"required,omitempty"`
	Properties                  Schemas        `json:"properties,omitempty" yaml:"properties,omitempty"`
	MinProps                    uint64         `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	MaxProps                    *uint64        `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	AdditionalPropertiesAllowed *bool          `multijson:"additionalProperties,omitempty" json:"-" yaml:"-"` // In this order...
	AdditionalProperties        *SchemaRef     `multijson:"additionalProperties,omitempty" json:"-" yaml:"-"` // ...for multijson
	Discriminator               *Discriminator `json:"discriminator,omitempty" yaml:"discriminator,omitempty"`
}

var _ jsonpointer.JSONPointable = (*Schema)(nil)

func NewSchema() *Schema {
	return &Schema{}
}

func (schema *Schema) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(schema)
}

func (schema *Schema) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, schema)
}

func (schema Schema) JSONLookup(token string) (interface{}, error) {
	switch token {
	case "additionalProperties":
		if schema.AdditionalProperties != nil {
			if schema.AdditionalProperties.Ref != "" {
				return &Ref{Ref: schema.AdditionalProperties.Ref}, nil
			}
			return schema.AdditionalProperties.Value, nil
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
	case "additionalPropertiesAllowed":
		return schema.AdditionalPropertiesAllowed, nil
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
	}

	v, _, err := jsonpointer.GetForToken(schema.ExtensionProps, token)
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
		Type: "boolean",
	}
}

func NewFloat64Schema() *Schema {
	return &Schema{
		Type: "number",
	}
}

func NewIntegerSchema() *Schema {
	return &Schema{
		Type: "integer",
	}
}

func NewInt32Schema() *Schema {
	return &Schema{
		Type:   "integer",
		Format: "int32",
	}
}

func NewInt64Schema() *Schema {
	return &Schema{
		Type:   "integer",
		Format: "int64",
	}
}

func NewStringSchema() *Schema {
	return &Schema{
		Type: "string",
	}
}

func NewDateTimeSchema() *Schema {
	return &Schema{
		Type:   "string",
		Format: "date-time",
	}
}

func NewUUIDSchema() *Schema {
	return &Schema{
		Type:   "string",
		Format: "uuid",
	}
}

func NewBytesSchema() *Schema {
	return &Schema{
		Type:   "string",
		Format: "byte",
	}
}

func NewArraySchema() *Schema {
	return &Schema{
		Type: "array",
	}
}

func NewObjectSchema() *Schema {
	return &Schema{
		Type:       "object",
		Properties: make(map[string]*SchemaRef),
	}
}

func (schema *Schema) WithNullable() *Schema {
	schema.Nullable = true
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
func (schema *Schema) WithExclusiveMin(value bool) *Schema {
	schema.ExclusiveMin = value
	return schema
}

func (schema *Schema) WithExclusiveMax(value bool) *Schema {
	schema.ExclusiveMax = value
	return schema
}

func (schema *Schema) WithEnum(values ...interface{}) *Schema {
	schema.Enum = values
	return schema
}

func (schema *Schema) WithDefault(defaultValue interface{}) *Schema {
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
	schema.compiledPattern = nil
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
	properties := schema.Properties
	if properties == nil {
		properties = make(map[string]*SchemaRef)
		schema.Properties = properties
	}
	properties[name] = ref
	return schema
}

func (schema *Schema) WithProperties(properties map[string]*Schema) *Schema {
	result := make(map[string]*SchemaRef, len(properties))
	for k, v := range properties {
		result[k] = &SchemaRef{
			Value: v,
		}
	}
	schema.Properties = result
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
	schema.AdditionalProperties = nil
	t := true
	schema.AdditionalPropertiesAllowed = &t
	return schema
}

func (schema *Schema) WithAdditionalProperties(v *Schema) *Schema {
	if v == nil {
		schema.AdditionalProperties = nil
	} else {
		schema.AdditionalProperties = &SchemaRef{
			Value: v,
		}
	}
	return schema
}

func (schema *Schema) IsEmpty() bool {
	if schema.Type != "" || schema.Format != "" || len(schema.Enum) != 0 ||
		schema.UniqueItems || schema.ExclusiveMin || schema.ExclusiveMax ||
		schema.Nullable || schema.ReadOnly || schema.WriteOnly || schema.AllowEmptyValue ||
		schema.Min != nil || schema.Max != nil || schema.MultipleOf != nil ||
		schema.MinLength != 0 || schema.MaxLength != nil || schema.Pattern != "" ||
		schema.MinItems != 0 || schema.MaxItems != nil ||
		len(schema.Required) != 0 ||
		schema.MinProps != 0 || schema.MaxProps != nil {
		return false
	}
	if n := schema.Not; n != nil && !n.Value.IsEmpty() {
		return false
	}
	if ap := schema.AdditionalProperties; ap != nil && !ap.Value.IsEmpty() {
		return false
	}
	if apa := schema.AdditionalPropertiesAllowed; apa != nil && !*apa {
		return false
	}
	if items := schema.Items; items != nil && !items.Value.IsEmpty() {
		return false
	}
	for _, s := range schema.Properties {
		if !s.Value.IsEmpty() {
			return false
		}
	}
	for _, s := range schema.OneOf {
		if !s.Value.IsEmpty() {
			return false
		}
	}
	for _, s := range schema.AnyOf {
		if !s.Value.IsEmpty() {
			return false
		}
	}
	for _, s := range schema.AllOf {
		if !s.Value.IsEmpty() {
			return false
		}
	}
	return true
}

// func (schema *Schema) expectedType(settings *schemaValidationSettings, typ string) error {
// 	if settings.failfast {
// 		return errSchema
// 	}
// 	return &SchemaError{
// 		Value:       typ,
// 		Schema:      schema,
// 		SchemaField: "type",
// 		Reason:      "Field must be set to " + schema.Type + " or not be present",
// 	}
// }

// func (schema *Schema) compilePattern() (err error) {
// 	if schema.compiledPattern, err = regexp.Compile(schema.Pattern); err != nil {
// 		return &SchemaError{
// 			Schema:      schema,
// 			SchemaField: "pattern",
// 			Reason:      fmt.Sprintf("cannot compile pattern %q: %v", schema.Pattern, err),
// 		}
// 	}
// 	return nil
// }

// type SchemaError struct {
// 	Value       interface{}
// 	reversePath []string
// 	Schema      *Schema
// 	SchemaField string
// 	Reason      string
// 	Origin      error
// }

// func markSchemaErrorKey(err error, key string) error {
// 	if v, ok := err.(*SchemaError); ok {
// 		v.reversePath = append(v.reversePath, key)
// 		return v
// 	}
// 	if v, ok := err.(MultiError); ok {
// 		for _, e := range v {
// 			_ = markSchemaErrorKey(e, key)
// 		}
// 		return v
// 	}
// 	return err
// }

// func markSchemaErrorIndex(err error, index int) error {
// 	if v, ok := err.(*SchemaError); ok {
// 		v.reversePath = append(v.reversePath, strconv.FormatInt(int64(index), 10))
// 		return v
// 	}
// 	if v, ok := err.(MultiError); ok {
// 		for _, e := range v {
// 			_ = markSchemaErrorIndex(e, index)
// 		}
// 		return v
// 	}
// 	return err
// }

// func (err *SchemaError) JSONPointer() []string {
// 	reversePath := err.reversePath
// 	path := append([]string(nil), reversePath...)
// 	for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
// 		path[left], path[right] = path[right], path[left]
// 	}
// 	return path
// }

// func (err *SchemaError) Error() string {
// 	if err.Origin != nil {
// 		return err.Origin.Error()
// 	}

// 	buf := bytes.NewBuffer(make([]byte, 0, 256))
// 	if len(err.reversePath) > 0 {
// 		buf.WriteString(`Error at "`)
// 		reversePath := err.reversePath
// 		for i := len(reversePath) - 1; i >= 0; i-- {
// 			buf.WriteByte('/')
// 			buf.WriteString(reversePath[i])
// 		}
// 		buf.WriteString(`": `)
// 	}
// 	reason := err.Reason
// 	if reason == "" {
// 		buf.WriteString(`Doesn't match schema "`)
// 		buf.WriteString(err.SchemaField)
// 		buf.WriteString(`"`)
// 	} else {
// 		buf.WriteString(reason)
// 	}
// 	if !SchemaErrorDetailsDisabled {
// 		buf.WriteString("\nSchema:\n  ")
// 		encoder := json.NewEncoder(buf)
// 		encoder.SetIndent("  ", "  ")
// 		if err := encoder.Encode(err.Schema); err != nil {
// 			panic(err)
// 		}
// 		buf.WriteString("\nValue:\n  ")
// 		if err := encoder.Encode(err.Value); err != nil {
// 			panic(err)
// 		}
// 	}
// 	return buf.String()
// }

// func isSliceOfUniqueItems(xs []interface{}) bool {
// 	s := len(xs)
// 	m := make(map[string]struct{}, s)
// 	for _, x := range xs {
// 		// The input slice is coverted from a JSON string, there shall
// 		// have no error when covert it back.
// 		key, _ := json.Marshal(&x)
// 		m[string(key)] = struct{}{}
// 	}
// 	return s == len(m)
// }

// // SliceUniqueItemsChecker is an function used to check if an given slice
// // have unique items.
// type SliceUniqueItemsChecker func(items []interface{}) bool

// // By default using predefined func isSliceOfUniqueItems which make use of
// // json.Marshal to generate a key for map used to check if a given slice
// // have unique items.
// var sliceUniqueItemsChecker SliceUniqueItemsChecker = isSliceOfUniqueItems

// // RegisterArrayUniqueItemsChecker is used to register a customized function
// // used to check if JSON array have unique items.
// func RegisterArrayUniqueItemsChecker(fn SliceUniqueItemsChecker) {
// 	sliceUniqueItemsChecker = fn
// }

// func unsupportedFormat(format string) error {
// 	return fmt.Errorf("unsupported 'format' value %q", format)
// }
