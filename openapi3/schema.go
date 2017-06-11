package openapi3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jban332/kinapi/jsoninfo"
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
	jsoninfo.RefProps
	jsoninfo.ExtensionProps

	OneOf        []*Schema     `json:"oneOf,omitempty"`
	AnyOf        []*Schema     `json:"anyOf,omitempty"`
	AllOf        []*Schema     `json:"allOf,omitempty"`
	Not          *Schema       `json:"not,omitempty"`
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
	MinItems int64   `json:"minItems,omitempty"`
	MaxItems *int64  `json:"maxItems,omitempty"`
	Items    *Schema `json:"items,omitempty"`

	// Object
	Required                    []string           `json:"required,omitempty"`
	Properties                  map[string]*Schema `json:"properties,omitempty"`
	AdditionalProperties        *Schema            `json:"-" multijson:"additionalProperties,omitempty"`
	AdditionalPropertiesAllowed bool               `json:"-" multijson:"additionalProperties,omitempty"`
	Discriminator               string             `json:"discriminator,omitempty"`

	// A propriatery extension we thought is useful for many
	AdditionalKeys *Schema `json:"x-additionalKeys,omitempty"`
}

func NewSchema() *Schema {
	return &Schema{}
}

func (value *Schema) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStructFields(value)
}

func (value *Schema) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStructFields(data, value)
}

func NewOneOfSchema(schemas ...*Schema) *Schema {
	return &Schema{
		OneOf: schemas,
	}
}

func NewAllOfSchema(schemas ...*Schema) *Schema {
	return &Schema{
		AllOf: schemas,
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
		Properties: make(map[string]*Schema),
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
	schema.Items = value
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
	properties := schema.Properties
	if properties == nil {
		properties = make(map[string]*Schema)
		schema.Properties = properties
	}
	properties[name] = propertySchema
	return schema
}

func (schema *Schema) WithProperties(properties map[string]*Schema) *Schema {
	schema.Properties = properties
	return schema
}

func (schema *Schema) WithAnyAdditionalProperties() *Schema {
	schema.AdditionalProperties = NewObjectSchema()
	return schema
}

func (schema *Schema) WithAdditionalProperties(additionalProperties *Schema) *Schema {
	schema.AdditionalProperties = additionalProperties
	return schema
}

// AddToSchemaSet puts this schema and all referred schemas to a map
func (schema *Schema) AddToSchemaSet(schemas map[*Schema]struct{}) {
	if _, exists := schemas[schema]; exists {
		return
	}
	schemas[schema] = struct{}{}
	if v := schema.Items; v != nil {
		v.AddToSchemaSet(schemas)
	}
	if properties := schema.Properties; properties != nil {
		for _, v := range properties {
			v.AddToSchemaSet(schemas)
		}
	}
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
	for _, v := range schema.OneOf {
		if err := v.validate(stack, c); err == nil {
			return err
		}
	}
	for _, v := range schema.AnyOf {
		if err := v.validate(stack, c); err != nil {
			return err
		}
	}
	for _, v := range schema.AllOf {
		if err := v.validate(stack, c); err != nil {
			return err
		}
	}
	if v := schema.Not; v != nil {
		if err := v.validate(stack, c); err != nil {
			return err
		}
	}
	schemaType := schema.Type
	switch schemaType {
	case "":
	case "boolean":
	case "number":
		if format := schema.Format; len(format) > 0 {
			switch format {
			case "int32", "int64", "float", "double":
			default:
				return fmt.Errorf("Unsupported 'format' value '%v", format)
			}
		}
	case "string":
	case "array":
		if schema.Items == nil {
			return fmt.Errorf("When schema type is 'array', schema 'items' must be non-null")
		}
	case "object":
	default:
		return fmt.Errorf("Unsupported 'type' value '%v", schemaType)
	}
	if v := schema.Items; v != nil {
		if err := v.validate(stack, c); err != nil {
			return err
		}
	}
	if v := schema.Properties; v != nil {
		for _, x := range v {
			if err := x.validate(stack, c); err != nil {
				return err
			}
		}
	}
	if v := schema.AdditionalProperties; v != nil {
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
	if v := schema.Not; v != nil {
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
			err := item.visitJSON(value, true)
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
			err := item.visitJSON(value, true)
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
			err := item.visitJSON(value, true)
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
	switch schema.Type {
	case "":
		err := schema.validateTypeListAllows("null", fast)
		if err != nil {
			return err
		}
	case "null":
		return nil
	default:
		return &SchemaError{
			Value:       nil,
			Schema:      schema,
			SchemaField: "type",
		}
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
	schemaType := schema.Type
	switch schemaType {
	case "":
		if err := schema.validateTypeListAllows("boolean", fast); err != nil {
			return err
		}
	case "boolean":
	default:
		if fast {
			return errSchema
		}
		return &SchemaError{
			Value:       value,
			Schema:      schema,
			SchemaField: "type",
		}
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
	schemaType := schema.Type
	switch schemaType {
	case "":
		if err := schema.validateTypeListAllows("number", fast); err != nil {
			return err
		}
	case "number":
	default:
		if fast {
			return errSchema
		}
		return &SchemaError{
			Value:       value,
			Schema:      schema,
			SchemaField: "type",
		}
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
	schemaType := schema.Type
	switch schemaType {
	case "":
		if err := schema.validateTypeListAllows("string", fast); err != nil {
			if fast {
				return errSchema
			}
			return err
		}
	case "string":
	default:
		if fast {
			return errSchema
		}
		return &SchemaError{
			Value:       value,
			Schema:      schema,
			SchemaField: "type",
		}
	}
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
			// No pattern, but does have a schema
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
	schemaType := schema.Type
	switch schemaType {
	case "":
		if err := schema.validateTypeListAllows("array", fast); err != nil {
			return err
		}
	case "array":
	default:
		if fast {
			return errSchema
		}
		return &SchemaError{
			Value:       value,
			Schema:      schema,
			SchemaField: "type",
		}
	}
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
	if itemSchema := schema.Items; itemSchema != nil {
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
	schemaType := schema.Type
	switch schemaType {
	case "":
		if err := schema.validateTypeListAllows("object", fast); err != nil {
			return err
		}
	case "object":
	default:
		if fast {
			return errSchema
		}
		return &SchemaError{
			Value:       value,
			Schema:      schema,
			SchemaField: "type",
		}
	}
	properties := schema.Properties
	additionalKeys := schema.AdditionalKeys
	additionalProperties := schema.AdditionalProperties
	for k, v := range value {
		if properties != nil {
			p := properties[k]
			if p != nil {
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
			if additionalKeys != nil {
				err := additionalKeys.VisitJSONString(k)
				if err != nil {
					if fast {
						return errSchema
					}
					return &SchemaError{
						Schema:      schema,
						SchemaField: "x-additionalKeys",
						Reason:      fmt.Sprintf("Invalid property name '%s'", k),
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
