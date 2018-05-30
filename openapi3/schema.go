package openapi3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ronniedada/kin-openapi/jsoninfo"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"
)

// Float64Ptr is a helper for defining OpenAPI schemas.
func Float64Ptr(value float64) *float64 {
	return &value
}

// Int64Ptr is a helper for defining OpenAPI schemas.
func Int64Ptr(value int64) *int64 {
	return &value
}

// Schema is specified by OpenAPI/Swagger 3.0 standard.
type Schema struct {
	ExtensionProps

	OneOf        []*SchemaRef  `json:"oneOf,omitempty"`
	AnyOf        []*SchemaRef  `json:"anyOf,omitempty"`
	AllOf        []*SchemaRef  `json:"allOf,omitempty"`
	Not          *SchemaRef    `json:"not,omitempty"`
	Type         string        `json:"-" multijson:"type,omitempty"`
	Types        []string      `json:"-" multijson:"type,omitempty"`
	Format       string        `json:"format,omitempty"`
	Description  string        `json:"description,omitempty"`
	Enum         []interface{} `json:"enum,omitempty"`
	Default      interface{}   `json:"default,omitempty"`
	Example      interface{}   `json:"example,omitempty"`
	Examples     []interface{} `json:"examples,omitempty"`
	ExternalDocs interface{}   `json:"externalDocs,omitempty"`

	// Properties
	Nullable  bool        `json:"nullable,omitempty"`
	ReadOnly  bool        `json:"readOnly,omitempty"`
	WriteOnly bool        `json:"writeOnly,omitempty"`
	XML       interface{} `json:"xml,omitempty"`

	// Number
	ExclusiveMin *float64 `json:"exclusiveMin,omitempty"`
	ExclusiveMax *float64 `json:"exclusiveMax,omitempty"`
	Min          *float64 `json:"min,omitempty"`
	Max          *float64 `json:"max,omitempty"`
	Multiple     int64    `json:"multiple,omitempty"`

	// String
	MinLength       int64  `json:"minLength,omitempty"`
	MaxLength       *int64 `json:"maxLength,omitempty"`
	Pattern         string `json:"pattern,omitempty"`
	compiledPattern *compiledPattern

	// Array
	MinItems int64      `json:"minItems,omitempty"`
	MaxItems *int64     `json:"maxItems,omitempty"`
	Items    *SchemaRef `json:"items,omitempty"`

	// Object
	Required                    []string              `json:"required,omitempty"`
	Properties                  map[string]*SchemaRef `json:"properties,omitempty"`
	AdditionalProperties        *SchemaRef            `json:"-" multijson:"additionalProperties,omitempty"`
	AdditionalPropertiesAllowed bool                  `json:"-" multijson:"additionalProperties,omitempty"`
	Discriminator               string                `json:"discriminator,omitempty"`

	PatternProperties         string `json:"patternProperties,omitempty"`
	compiledPatternProperties *compiledPattern
}

func NewSchema() *Schema {
	return &Schema{}
}

func (value *Schema) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(value)
}

func (value *Schema) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, value)
}

func (value *Schema) NewRef() *SchemaRef {
	return &SchemaRef{
		Value: value,
	}
}

func NewOneOfSchema(schemas ...*Schema) *Schema {
	refs := make([]*SchemaRef, len(schemas))
	for i, schema := range schemas {
		refs[i].Value = schema
	}
	return &Schema{
		OneOf: refs,
	}
}

func NewAnyOfSchema(schemas ...*Schema) *Schema {
	refs := make([]*SchemaRef, len(schemas))
	for i, schema := range schemas {
		refs[i].Value = schema
	}
	return &Schema{
		AnyOf: refs,
	}
}

func NewAllOfSchema(schemas ...*Schema) *Schema {
	refs := make([]*SchemaRef, len(schemas))
	for i, schema := range schemas {
		refs[i].Value = schema
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

func NewInt32Schema() *Schema {
	return &Schema{
		Type:   "number",
		Format: "int32",
	}
}

func NewInt64Schema() *Schema {
	return &Schema{
		Type:   "number",
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

type compiledPattern struct {
	Regexp    *regexp.Regexp
	ErrReason string
}

func (schema *Schema) WithMin(value float64) *Schema {
	schema.Min = &value
	return schema
}

func (schema *Schema) WithMax(value float64) *Schema {
	schema.Max = &value
	return schema
}
func (schema *Schema) WithExclusiveMin(value float64) *Schema {
	schema.ExclusiveMin = &value
	return schema
}

func (schema *Schema) WithExclusiveMax(value float64) *Schema {
	schema.ExclusiveMax = &value
	return schema
}

func (schema *Schema) WithEnum(values ...interface{}) *Schema {
	schema.Enum = values
	return schema
}

func (schema *Schema) WithFormat(value string) *Schema {
	schema.Format = value
	return schema
}

func (schema *Schema) WithLength(n int64) *Schema {
	schema.MinLength = n
	schema.MaxLength = &n
	return schema
}

func (schema *Schema) WithMinLength(n int64) *Schema {
	schema.MinLength = n
	return schema
}

func (schema *Schema) WithMaxLength(n int64) *Schema {
	schema.MaxLength = &n
	return schema
}

func (schema *Schema) WithLengthDecodedBase64(n int64) *Schema {
	v := (n*8 + 5) / 6
	schema.MinLength = v
	schema.MaxLength = &v
	return schema
}

func (schema *Schema) WithMinLengthDecodedBase64(n int64) *Schema {
	schema.MinLength = (n*8 + 5) / 6
	return schema
}

func (schema *Schema) WithMaxLengthDecodedBase64(n int64) *Schema {
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

func (schema *Schema) WithMinItems(n int64) *Schema {
	schema.MinItems = n
	return schema
}

func (schema *Schema) WithMaxItems(n int64) *Schema {
	schema.MaxItems = &n
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

func (schema *Schema) WithAnyAdditionalProperties() *Schema {
	schema.AdditionalProperties = nil
	schema.AdditionalPropertiesAllowed = true
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

func (schema *Schema) TypesContains(value string) bool {
	if schema.Type == value {
		return true
	}
	for _, item := range schema.Types {
		if item == value {
			return true
		}
	}
	return false
}

func (schema *Schema) Validate(c context.Context) error {
	return schema.validate(make([]*Schema, 2), c)
}

func (schema *Schema) validate(stack []*Schema, c context.Context) error {
	for _, existing := range stack {
		if existing == schema {
			return nil
		}
	}
	stack = append(stack, schema)
	for _, item := range schema.OneOf {
		v := item.Value
		if v == nil {
			return foundUnresolvedRef(item.Ref)
		}
		if err := v.validate(stack, c); err == nil {
			return err
		}
	}
	for _, item := range schema.AnyOf {
		v := item.Value
		if v == nil {
			return foundUnresolvedRef(item.Ref)
		}
		if err := v.validate(stack, c); err != nil {
			return err
		}
	}
	for _, item := range schema.AllOf {
		v := item.Value
		if v == nil {
			return foundUnresolvedRef(item.Ref)
		}
		if err := v.validate(stack, c); err != nil {
			return err
		}
	}
	if ref := schema.Not; ref != nil {
		v := ref.Value
		if v == nil {
			return foundUnresolvedRef(ref.Ref)
		}
		if err := v.validate(stack, c); err != nil {
			return err
		}
	}
	schemaType := schema.Type
	switch schemaType {
	case "":
	case "integer", "long", "float", "double":
	case "string":
	case "byte":
	case "binary":
	case "boolean":
	case "date":
	case "dateTime":
	case "password":
	case "array":
		if schema.Items == nil {
			return fmt.Errorf("When schema type is 'array', schema 'items' must be non-null")
		}
	case "object":
	default:
		return fmt.Errorf("Unsupported 'type' value '%v", schemaType)
	}
	if ref := schema.Items; ref != nil {
		v := ref.Value
		if v == nil {
			return foundUnresolvedRef(ref.Ref)
		}
		if err := v.validate(stack, c); err != nil {
			return err
		}
	}
	if m := schema.Properties; m != nil {
		for _, ref := range m {
			v := ref.Value
			if v == nil {
				return foundUnresolvedRef(ref.Ref)
			}
			if err := v.validate(stack, c); err != nil {
				return err
			}
		}
	}
	if ref := schema.AdditionalProperties; ref != nil {
		v := ref.Value
		if v == nil {
			return foundUnresolvedRef(ref.Ref)
		}
		if err := v.validate(stack, c); err != nil {
			return err
		}
	}
	return nil
}

func (schema *Schema) IsMatching(value interface{}) bool {
	return schema.visitJSON(value, true) == nil
}

func (schema *Schema) IsMatchingJSONBoolean(value bool) bool {
	return schema.visitJSONBoolean(value, true) == nil
}

func (schema *Schema) IsMatchingJSONNumber(value float64) bool {
	return schema.visitJSONNumber(value, true) == nil
}

func (schema *Schema) IsMatchingJSONString(value string) bool {
	return schema.visitJSONString(value, true) == nil
}

func (schema *Schema) IsMatchingJSONArray(value []interface{}) bool {
	return schema.visitJSONArray(value, true) == nil
}

func (schema *Schema) IsMatchingJSONObject(value map[string]interface{}) bool {
	return schema.visitJSONObject(value, true) == nil
}

var (
	errSchema = errors.New("Input does not match the schema")
)

func (schema *Schema) VisitJSON(value interface{}) error {
	return schema.visitJSON(value, false)
}

func (schema *Schema) visitJSON(value interface{}, fast bool) error {
	switch value := value.(type) {
	case nil:
		return schema.visitJSONNull(fast)
	case bool:
		return schema.visitJSONBoolean(value, fast)
	case float64:
		return schema.visitJSONNumber(value, fast)
	case string:
		return schema.visitJSONString(value, fast)
	case []interface{}:
		return schema.visitJSONArray(value, fast)
	case map[string]interface{}:
		return schema.visitJSONObject(value, fast)
	default:
		return &SchemaError{
			Schema:      schema,
			SchemaField: "type",
			Reason:      fmt.Sprintf("Not a JSON value: %T", value),
		}
	}
}

func (schema *Schema) visitSetOperations(value interface{}, fast bool) error {
	if ref := schema.Not; ref != nil {
		v := ref.Value
		if v == nil {
			return foundUnresolvedRef(ref.Ref)
		}
		if err := v.visitJSON(value, true); err == nil {
			if fast {
				return errSchema
			}
			return &SchemaError{
				Value:       value,
				Schema:      schema,
				SchemaField: "not",
			}
		}
	}
	if v := schema.OneOf; len(v) > 0 {
		ok := 0
		for _, item := range v {
			v := item.Value
			if v == nil {
				return foundUnresolvedRef(item.Ref)
			}
			err := v.visitJSON(value, true)
			if err == nil {
				ok++
			}
		}
		if ok == 0 || ok > 1 {
			if fast {
				return errSchema
			}
			return &SchemaError{
				Value:       value,
				Schema:      schema,
				SchemaField: "oneOf",
			}
		}
	}
	if v := schema.AnyOf; len(v) > 0 {
		ok := false
		for _, item := range v {
			v := item.Value
			if v == nil {
				return foundUnresolvedRef(item.Ref)
			}
			err := v.visitJSON(value, true)
			if err == nil {
				ok = true
				break
			}
		}
		if !ok {
			if fast {
				return errSchema
			}
			return &SchemaError{
				Value:       value,
				Schema:      schema,
				SchemaField: "anyOf",
			}
		}
	}
	if v := schema.AllOf; len(v) > 0 {
		for _, item := range v {
			v := item.Value
			if v == nil {
				return foundUnresolvedRef(item.Ref)
			}
			err := v.visitJSON(value, true)
			if err != nil {
				if fast {
					return errSchema
				}
				return &SchemaError{
					Value:       value,
					Schema:      schema,
					SchemaField: "allOf",
				}
			}
		}
	}
	return nil
}

func (schema *Schema) visitJSONNull(fast bool) error {
	err := schema.visitSetOperations(nil, fast)
	if err != nil {
		return err
	}
	err = schema.validateTypeListAllows("null", fast)
	if err != nil {
		return err
	}
	if enum := schema.Enum; enum != nil {
		found := false
	loop:
		for _, validValue := range enum {
			switch validValue.(type) {
			case nil:
				found = true
				break loop
			}
		}
		if !found {
			if fast {
				return errSchema
			}
			return &SchemaError{
				Value:       nil,
				Schema:      schema,
				SchemaField: "enum",
			}
		}
	}
	return nil
}

func (schema *Schema) VisitJSONBoolean(value bool) error {
	return schema.visitJSONBoolean(value, false)
}

func (schema *Schema) visitJSONBoolean(value bool, fast bool) error {
	err := schema.visitSetOperations(value, fast)
	if err != nil {
		return err
	}
	err = schema.validateTypeListAllows("boolean", fast)
	if err != nil {
		return err
	}
	if enum := schema.Enum; enum != nil {
		for _, validValue := range enum {
			switch validValue := validValue.(type) {
			case bool:
				if value == validValue {
					return nil
				}
			}
		}
		if fast {
			return errSchema
		}
		return &SchemaError{
			Value:       value,
			Schema:      schema,
			SchemaField: "enum",
		}
	}
	return nil
}

var (
	ErrSchemaInputNaN = errors.New("NaN is not allowed")
	ErrSchemaInputInf = errors.New("Inf is not allowed")
)

func (schema *Schema) VisitJSONNumber(value float64) error {
	return schema.visitJSONNumber(value, false)
}

func (schema *Schema) visitJSONNumber(value float64, fast bool) error {
	err := schema.visitSetOperations(value, fast)
	if err != nil {
		return err
	}
	if math.IsNaN(value) {
		return ErrSchemaInputNaN
	}
	if math.IsInf(value, 0) {
		return ErrSchemaInputInf
	}
	if err := schema.validateTypeListAllows("number", fast); err != nil {
		return err
	}
	if v := schema.Enum; v != nil {
		for _, item := range v {
			switch item := item.(type) {
			case float64:
				if value == item {
					return nil
				}
			}
		}
		if fast {
			return errSchema
		}
		return &SchemaError{
			Value:       value,
			Schema:      schema,
			SchemaField: "enum",
			Reason:      "JSON number is not one of the allowed values",
		}
	}
	if v := schema.ExclusiveMin; v != nil && !(*v < value) {
		if fast {
			return errSchema
		}
		return &SchemaError{
			Value:       value,
			Schema:      schema,
			SchemaField: "exclusiveMin",
			Reason:      fmt.Sprintf("Number must be more than %g", *v),
		}
	}
	if v := schema.ExclusiveMax; v != nil && !(*v > value) {
		if fast {
			return errSchema
		}
		return &SchemaError{
			Value:       value,
			Schema:      schema,
			SchemaField: "exclusiveMax",
			Reason:      fmt.Sprintf("Number must be less than %g", *v),
		}
	}
	if v := schema.Min; v != nil && !(*v <= value) {
		if fast {
			return errSchema
		}
		return &SchemaError{
			Value:       value,
			Schema:      schema,
			SchemaField: "min",
			Reason:      fmt.Sprintf("Number must be at least %g", *v),
		}
	}
	if v := schema.Max; v != nil && !(*v >= value) {
		if fast {
			return errSchema
		}
		return &SchemaError{
			Value:       value,
			Schema:      schema,
			SchemaField: "max",
			Reason:      fmt.Sprintf("Number must be most %g", *v),
		}
	}
	if v := schema.Multiple; v != 0 && float64(int64(value)/v*v) != value {
		if fast {
			return errSchema
		}
		return &SchemaError{
			Value:       value,
			Schema:      schema,
			SchemaField: "multiple",
		}
	}
	return nil
}

func (schema *Schema) VisitJSONString(value string) error {
	return schema.visitJSONString(value, false)
}

func (schema *Schema) visitJSONString(value string, fast bool) error {
	err := schema.visitSetOperations(value, fast)
	if err != nil {
		return err
	}
	if err := schema.validateTypeListAllows("string", fast); err != nil {
		return err
	}

	// "enum"
	if enum := schema.Enum; enum != nil {
		for _, validValue := range enum {
			switch validValue := validValue.(type) {
			case string:
				if value == validValue {
					return nil
				}
			}
		}
		if fast {
			return errSchema
		}
		return &SchemaError{
			Value:       value,
			Schema:      schema,
			SchemaField: "enum",
		}
	}

	// "minLength" and "maxLength"
	minLength := schema.MinLength
	maxLength := schema.MaxLength
	if minLength > 0 || maxLength != nil {
		// JON schema string lengths are UTF-16, not UTF-8!
		length := int64(0)
		for _, r := range value {
			if utf16.IsSurrogate(r) {
				length += 2
			} else {
				length++
			}
		}
		if minLength > 0 && length < minLength {
			if fast {
				return errSchema
			}
			return &SchemaError{
				Value:       value,
				Schema:      schema,
				SchemaField: "minLength",
				Reason:      fmt.Sprintf("Minimum string length is %d", minLength),
			}
		}
		if maxLength != nil && length > *maxLength {
			if fast {
				return errSchema
			}
			return &SchemaError{
				Value:       value,
				Schema:      schema,
				SchemaField: "maxLength",
				Reason:      fmt.Sprintf("Maximum string length is %d", *maxLength),
			}
		}
	}

	// "format" and "pattern"
	cp := schema.compiledPattern
	if cp == nil {
		pattern := schema.Pattern
		if v := schema.Pattern; len(v) > 0 {
			// Pattern
			re, err := regexp.Compile(v)
			if err != nil {
				return fmt.Errorf("Error while compiling regular expression '%s': %v", pattern, err)
			}
			cp = &compiledPattern{
				Regexp:    re,
				ErrReason: "JSON string doesn't match the regular expression '" + v + "'",
			}
			schema.compiledPattern = cp
		} else if v := schema.Format; len(v) > 0 {
			// No pattern, but does have a format
			re := SchemaStringFormats[v]
			if re != nil {
				cp = &compiledPattern{
					Regexp:    re,
					ErrReason: "JSON string doesn't match the format '" + v + " (regular expression `" + re.String() + "`)'",
				}
				schema.compiledPattern = cp
			}
		}
	}
	if cp != nil {
		if cp.Regexp.MatchString(value) == false {
			field := "format"
			if schema.Pattern != "" {
				field = "pattern"
			}
			return &SchemaError{
				Value:       value,
				Schema:      schema,
				SchemaField: field,
				Reason:      cp.ErrReason,
			}
		}
	}
	return nil
}

func (schema *Schema) VisitJSONArray(value []interface{}) error {
	return schema.visitJSONArray(value, false)
}

func (schema *Schema) visitJSONArray(value []interface{}, fast bool) error {
	err := schema.visitSetOperations(value, fast)
	if err != nil {
		return err
	}
	err = schema.validateTypeListAllows("array", fast)
	if err != nil {
		return err
	}

	// "minItems""
	if v := schema.MinItems; v != 0 && int64(len(value)) < v {
		if fast {
			return errSchema
		}
		return &SchemaError{
			Value:       value,
			Schema:      schema,
			SchemaField: "minItems",
			Reason:      fmt.Sprintf("Minimum number of items is %d", v),
		}
	}

	// "maxItems"
	if v := schema.MaxItems; v != nil && int64(len(value)) > *v {
		if fast {
			return errSchema
		}
		return &SchemaError{
			Value:       value,
			Schema:      schema,
			SchemaField: "maxItems",
			Reason:      fmt.Sprintf("Maximum number of items is %d", *v),
		}
	}

	// "items"
	if itemSchemaRef := schema.Items; itemSchemaRef != nil {
		itemSchema := itemSchemaRef.Value
		if itemSchema == nil {
			return foundUnresolvedRef(itemSchemaRef.Ref)
		}
		for i, item := range value {
			err := itemSchema.VisitJSON(item)
			if err != nil {
				return markSchemaErrorIndex(err, i)
			}
		}
	}
	return nil
}

func (schema *Schema) VisitJSONObject(value map[string]interface{}) error {
	return schema.visitJSONObject(value, false)
}

func (schema *Schema) visitJSONObject(value map[string]interface{}, fast bool) error {
	err := schema.visitSetOperations(value, fast)
	if err != nil {
		return err
	}
	err = schema.validateTypeListAllows("object", fast)
	if err != nil {
		return err
	}

	// "properties"
	properties := schema.Properties

	// "patternProperties"
	var cp *compiledPattern
	patternProperties := schema.PatternProperties
	if len(patternProperties) > 0 {
		cp = schema.compiledPatternProperties
		if cp == nil {
			re, err := regexp.Compile(patternProperties)
			if err != nil {
				return fmt.Errorf("Error while compiling regular expression '%s': %v", patternProperties, err)
			}
			cp = &compiledPattern{
				Regexp:    re,
				ErrReason: "JSON property doesn't match the regular expression '" + patternProperties + "'",
			}
			schema.compiledPatternProperties = cp
		}
	}

	// "additionalProperties"
	var additionalProperties *Schema
	if ref := schema.AdditionalProperties; ref != nil {
		additionalProperties = ref.Value
	}
	for k, v := range value {
		if properties != nil {
			propertyRef := properties[k]
			if propertyRef != nil {
				p := propertyRef.Value
				if p == nil {
					return foundUnresolvedRef(propertyRef.Ref)
				}
				err := p.VisitJSON(v)
				if err != nil {
					if fast {
						return errSchema
					}
					return markSchemaErrorKey(err, k)
				}
				continue
			}
		}
		if additionalProperties != nil || schema.AdditionalPropertiesAllowed {
			if cp != nil {
				if cp.Regexp.MatchString(k) == false {
					return &SchemaError{
						Schema:      schema,
						SchemaField: "patternProperties",
						Reason:      cp.ErrReason,
					}
				}
			}
			if additionalProperties != nil {
				err := additionalProperties.VisitJSON(v)
				if err != nil {
					if fast {
						return errSchema
					}
					return markSchemaErrorKey(err, k)
				}
			}
			continue
		}
		if fast {
			return errSchema
		}
		return &SchemaError{
			Value:       value,
			Schema:      schema,
			SchemaField: "properties",
			Reason:      fmt.Sprintf("Property '%s' is unsupported", k),
		}
	}
	for _, k := range schema.Required {
		if _, ok := value[k]; !ok {
			if fast {
				return errSchema
			}
			return &SchemaError{
				Value:       value,
				Schema:      schema,
				SchemaField: "required",
				Reason:      fmt.Sprintf("Property '%s' is missing", k),
			}
		}
	}
	return nil
}

func (schema *Schema) validateTypeListAllows(value string, fast bool) error {
	schemaType := schema.Type
	if schemaType != "" {
		if schemaType == value {
			return nil
		}
		if fast {
			return errSchema
		}
		return &SchemaError{
			Value:       value,
			Schema:      schema,
			SchemaField: "type",
			Reason:      fmt.Sprintf("Expected JSON types: %s", schemaType),
		}
	}
	schemaTypes := schema.Types
	if len(schemaTypes) == 0 {
		return nil
	}
	for _, t := range schemaTypes {
		if t == value {
			return nil
		}
	}
	if fast {
		return errSchema
	}
	return &SchemaError{
		Value:       value,
		Schema:      schema,
		SchemaField: "type",
		Reason:      fmt.Sprintf("Expected one of the JSON types: '%s'", strings.Join(schemaTypes, "', '")),
	}
}

func newSchemaError(schema *Schema, value interface{}, args ...interface{}) error {
	return errors.New(fmt.Sprint(args...))
}

type SchemaError struct {
	Value       interface{}
	reversePath []string
	Schema      *Schema
	SchemaField string
	Reason      string
}

func markSchemaErrorKey(err error, key string) error {
	if v, ok := err.(*SchemaError); ok {
		v.reversePath = append(v.reversePath, key)
		return v
	}
	return err
}

func markSchemaErrorIndex(err error, index int) error {
	if v, ok := err.(*SchemaError); ok {
		v.reversePath = append(v.reversePath, strconv.FormatInt(int64(index), 10))
		return v
	}
	return err
}

func (err *SchemaError) JSONPointer() []string {
	reversePath := err.reversePath
	path := make([]string, len(reversePath))
	for i := range path {
		path[i] = reversePath[len(path)-1-i]
	}
	return path
}

func (err *SchemaError) Error() string {
	buf := bytes.NewBuffer(make([]byte, 0, 256))
	if len(err.reversePath) > 0 {
		buf.WriteString(`Error at "`)
		reversePath := err.reversePath
		for i := len(reversePath) - 1; i >= 0; i-- {
			buf.WriteByte('/')
			buf.WriteString(reversePath[i])
		}
		buf.WriteString(`":`)
	}
	reason := err.Reason
	if reason == "" {
		buf.WriteString(`Doesn't match schema "`)
		buf.WriteString(err.SchemaField)
		buf.WriteString(`"`)
	} else {
		buf.WriteString(reason)
	}
	if SchemaErrorDetailsDisabled == false {
		buf.WriteString("\nSchema:\n  ")
		encoder := json.NewEncoder(buf)
		encoder.SetIndent("  ", "  ")
		encoder.Encode(err.Schema)
		buf.WriteString("\nValue:\n  ")
		encoder.Encode(err.Value)
	}
	return buf.String()
}

// SchemaErrorDetailsDisabled disables printing of details about schema errors.
var SchemaErrorDetailsDisabled = false
