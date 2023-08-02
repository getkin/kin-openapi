package openapi3

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
)

const (
	FormatErrorMessage = "Unable to resolve Format conflict: all Format values must be identical."
	TypeErrorMessage   = "Unable to resolve Format conflict: all Type values must be identical."
)

// Merge replaces objects under AllOf with a flattened equivalent
func Merge(schema Schema) (*Schema, error) {
	if !isListOfObjects(&schema) {
		return &schema, nil
	}

	if schema.AllOf != nil {
		mergedAllOf, err := mergeAllOf(schema.AllOf)
		if err != nil {
			return &Schema{}, err
		}
		schema = copy(mergedAllOf, schema) // temporary.
	}
	schema.AllOf = nil

	//todo: merge result of mergedAllOf with all other fields of base Schema.
	//todo: implement merge functions for OneOf, AnyOf, Items
	if schema.Properties != nil {
		properties, err := mergeProperties(schema.Properties)
		if err != nil {
			return &Schema{}, err
		}
		schema.Properties = properties

	}

	return &schema, nil
}

func mergeAllOf(allOf SchemaRefs) (Schema, error) {

	schemas := make([]Schema, 0) // naming
	for _, schema := range allOf {
		merged, err := Merge(*schema.Value)
		if err != nil {
			return Schema{}, err
		}
		schemas = append(schemas, *merged)
	}

	schema, err := mergeFields(schemas)
	if err != nil {
		return *schema, err
	}
	return *schema, nil
}

func mergeProperties(schemas Schemas) (Schemas, error) {
	res := make(Schemas)
	for name, schemaRef := range schemas {
		merged, err := Merge(*schemaRef.Value)
		if err != nil {
			return res, err
		}
		schemaRef.Value = merged
		res[name] = schemaRef
	}
	return res, nil
}

func mergeFields(schemas []Schema) (*Schema, error) {
	result := NewSchema()
	titles := getStringValues(schemas, "title")
	if len(titles) > 0 {
		result.Title = titleResolver(titles)
	}

	required := getStringValues(schemas, "required")
	if len(required) > 0 {
		result.Required = resolveRequired(required)
	}

	description := getStringValues(schemas, "description")
	if len(description) > 0 {
		result.Description = resolveDescriptions(description)
	}

	formats := getStringValues(schemas, "format")
	if len(formats) > 0 {
		res, err := resolveFormat(formats)
		if err != nil {
			return result, err
		}
		result.Format = res
	}

	types := getStringValues(schemas, "type")
	if len(types) > 0 {
		res, err := resolveType(types)
		if err != nil {
			return result, err
		}
		result.Type = res
	}

	minLength := getUint64Values(schemas, "minLength")
	if len(minLength) > 0 {
		result.MinLength = resolveMinLength(minLength)
	}

	maxLength := getUint64Values(schemas, "maxLength")
	if len(maxLength) > 0 {
		result.MaxLength = Uint64Ptr(resolveMaxLength(maxLength))
	}

	minimum, isExcludedMin := resolveMinimumRange(schemas)
	if minimum != nil {
		result.Min = minimum
		result.ExclusiveMin = isExcludedMin
	}

	maximum, isExcludedMax := resolveMaximumRange(schemas)
	if maximum != nil {
		result.Max = maximum
		result.ExclusiveMax = isExcludedMax
	}

	minItems := getUint64Values(schemas, "minItems")
	if len(minItems) > 0 {
		result.MinItems = resolveMinItems(minItems)
	}

	maxItems := getUint64Values(schemas, "maxItems")
	if len(maxItems) > 0 {
		result.MaxItems = Uint64Ptr(resolveMaxItems(maxItems))
	}

	patterns := getStringValues(schemas, "pattern")
	if len(patterns) > 0 {
		result.Pattern = resolvePattern(patterns)
	}

	properties := getProperties(schemas)
	if len(properties) > 0 {
		res, err := resolveProperties(properties)
		if err != nil {
			return result, err
		}
		result.Properties = res
	}

	enum := getEnum(schemas, "enum")
	if len(enum) > 0 {
		result.Enum = resolveEnum(enum)
	}

	multipleOf := getFloat64Values(schemas, "multipleOf")
	if len(multipleOf) > 0 {
		result.MultipleOf = Float64Ptr(resolveMultipleOf(multipleOf))
	}

	items := getItems(schemas)
	if len(items) > 0 {
		res, err := resolveItems(items)
		if err != nil {
			return result, err
		}
		ref := SchemaRef{
			Value: res,
		}
		result.Items = &ref
	}

	// uniqueItems := getBoolValues(schemas, "uniqueItems")
	// if len(uniqueItems) > 0 {
	// 	result.UniqueItems = resolveUniqueItems(uniqueItems)
	// }

	return result, nil
}

/* Items */
func resolveItems(items []Schema) (*Schema, error) {
	s, err := mergeFields(items)
	if err != nil {
		return s, err
	}
	return s, nil
}

func getItems(schemas []Schema) []Schema {
	items := []Schema{}
	for _, s := range schemas {
		if s.Items != nil {
			items = append(items, *(s.Items.Value))
		}
	}
	return items
}

func resolveUniqueItems(values []bool) bool {
	for _, v := range values {
		if v == true {
			return true
		}
	}
	return false
}

/* MultipleOf */
func gcd(a, b uint64) uint64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func lcm(a, b uint64) uint64 {
	return a * b / gcd(a, b)
}

func containsNonInteger(arr []float64) bool {
	for _, num := range arr {
		if num != math.Trunc(num) {
			return true
		}
	}
	return false
}

func resolveMultipleOf(values []float64) float64 {
	factor := 1.0
	for containsNonInteger(values) {
		factor *= 10.0
		for i := range values {
			values[i] *= factor
		}
	}

	uintValues := make([]uint64, len(values))
	for i, val := range values {
		uintValues[i] = uint64(val)
	}

	lcmValue := uintValues[0]
	for _, v := range uintValues {
		lcmValue = lcm(lcmValue, v)
	}

	return float64(lcmValue) / factor
}

/* Properties */
func getProperties(schemas []Schema) []Schemas {
	sr := []Schemas{}
	for _, s := range schemas {
		if s.Properties != nil {
			sr = append(sr, s.Properties)
		}
	}
	return sr
}

func resolveProperties(schemas []Schemas) (Schemas, error) {
	allRefs := map[string][]Schema{} //naming
	for _, schema := range schemas { //naming
		for name, schemaRef := range schema {
			allRefs[name] = append(allRefs[name], *schemaRef.Value)
		}
	}
	result := make(Schemas)
	for name, schemas := range allRefs {
		merged, err := mergeFields(schemas)
		if err != nil {
			return Schemas{}, err
		}
		ref := SchemaRef{
			Value: merged,
		}
		result[name] = &ref
	}
	return result, nil
}

func getEnum(schemas []Schema, field string) []interface{} {
	enums := make([]interface{}, 0)
	for _, schema := range schemas {
		if schema.Enum != nil {
			enums = append(enums, schema.Enum...)
		}
	}
	return enums
}

func resolveEnum(values []interface{}) []interface{} {
	return uniqueValues(values)
}

func resolvePattern(values []string) string {
	var pattern strings.Builder
	for _, p := range values {
		pattern.WriteString(fmt.Sprintf("(?=%s)", p))
	}
	return pattern.String()
}

func resolveMinLength(values []uint64) uint64 {
	return findMaxValue(values)
}

func resolveMaxLength(values []uint64) uint64 {
	return findMinValue(values)
}

func resolveMinItems(values []uint64) uint64 {
	return findMaxValue(values)
}

func resolveMaxItems(values []uint64) uint64 {
	return findMinValue(values)
}

func findMaxValue(values []uint64) uint64 {
	max := uint64(0)
	for _, num := range values {
		if num > max {
			max = num
		}
	}
	return max
}

func findMinValue(values []uint64) uint64 {
	min := uint64(math.MaxUint64)
	for _, num := range values {
		if num < min {
			min = num
		}
	}
	return min
}

func resolveMaximumRange(schemas []Schema) (*float64, bool) {
	min := math.Inf(1)
	isExcluded := false
	var value *float64
	for _, s := range schemas {
		if s.Max != nil {
			if *s.Max < min {
				min = *s.Max
				value = s.Max
				isExcluded = s.ExclusiveMax
			}
		}
	}
	return value, isExcluded
}

func resolveMinimumRange(schemas []Schema) (*float64, bool) {
	max := math.Inf(-1)
	isExcluded := false
	var value *float64
	for _, s := range schemas {
		if s.Min != nil {
			if *s.Min > max {
				max = *s.Min
				value = s.Min
				isExcluded = s.ExclusiveMin
			}
		}
	}
	return value, isExcluded
}

func resolveDescriptions(values []string) string {
	return values[0]
}

func resolveType(values []string) (string, error) {
	if allStringsEqual(values) {
		return values[0], nil
	}
	return values[0], errors.New(TypeErrorMessage)
}

func resolveFormat(values []string) (string, error) {
	if allStringsEqual(values) {
		return values[0], nil
	}
	return values[0], errors.New(FormatErrorMessage)
}

func titleResolver(values []string) string {
	return values[0]
}

func resolveRequired(values []string) []string {
	uniqueMap := make(map[string]bool)
	var uniqueValues []string
	for _, str := range values {
		if _, found := uniqueMap[str]; !found {
			uniqueMap[str] = true
			uniqueValues = append(uniqueValues, str)
		}
	}
	return uniqueValues
}

func isListOfObjects(schema *Schema) bool {
	if schema == nil || schema.AllOf == nil {
		return false
	}

	for _, subSchema := range schema.AllOf {
		if subSchema.Value.Type != "object" {
			return false
		}
	}

	return true
}

func getStringValues(schemas []Schema, field string) []string {
	values := []string{}
	for _, schema := range schemas {
		value, err := schema.JSONLookup(field)
		if err != nil {
			log.Fatal(err.Error())
		}
		switch v := value.(type) {
		case string:
			if len(v) > 0 {
				values = append(values, v)
			}
		case []string:
			values = append(values, v...)
		}
	}
	return values
}

func getUint64Values(schemas []Schema, field string) []uint64 {
	values := []uint64{}
	for _, schema := range schemas {
		value, err := schema.JSONLookup(field)
		if err != nil {
			log.Fatal(err.Error())
		}
		if v, ok := value.(*uint64); ok {
			if v != nil {
				values = append(values, *v)
			}
		}
		if v, ok := value.(uint64); ok {
			values = append(values, v)
		}
	}
	return values
}

func getFloat64Values(schemas []Schema, field string) []float64 {
	values := []float64{}
	for _, schema := range schemas {
		value, err := schema.JSONLookup(field)
		if err != nil {
			log.Fatal(err.Error())
		}

		if v, ok := value.(*float64); ok {
			if v != nil {
				values = append(values, *v)
			}
		}
	}
	return values
}

func getBoolValues(schemas []Schema, field string) []bool {
	values := []bool{}
	for _, schema := range schemas {
		value, err := schema.JSONLookup(field)
		if err != nil {
			log.Fatal(err.Error())
		}
		if v, ok := value.(bool); ok {
			values = append(values, v)
		}
	}
	return values
}

func allStringsEqual(values []string) bool {
	first := values[0]
	for _, value := range values {
		if first != value {
			return false
		}
	}
	return true
}

func uniqueValues(values []interface{}) []interface{} {
	uniqueMap := make(map[interface{}]struct{})
	uniqueValues := make([]interface{}, 0)
	for _, value := range values {
		if _, found := uniqueMap[value]; !found {
			uniqueMap[value] = struct{}{}
			uniqueValues = append(uniqueValues, value)
		}
	}
	return uniqueValues
}

/* temporary */
func copy(source Schema, destination Schema) Schema {
	destination.Extensions = source.Extensions
	destination.OneOf = source.OneOf
	destination.AnyOf = source.AnyOf
	destination.AllOf = source.AllOf
	destination.Not = source.Not
	destination.Type = source.Type
	destination.Title = source.Title
	destination.Format = source.Format
	destination.Description = source.Description
	destination.Enum = source.Enum
	destination.Default = source.Default
	destination.Example = source.Example
	destination.ExternalDocs = source.ExternalDocs
	destination.UniqueItems = source.UniqueItems
	destination.ExclusiveMin = source.ExclusiveMin
	destination.ExclusiveMax = source.ExclusiveMax
	destination.Nullable = source.Nullable
	destination.ReadOnly = source.ReadOnly
	destination.WriteOnly = source.WriteOnly
	destination.AllowEmptyValue = source.AllowEmptyValue
	destination.Deprecated = source.Deprecated
	destination.XML = source.XML
	destination.Min = source.Min
	destination.Max = source.Max
	destination.MultipleOf = source.MultipleOf
	destination.MinLength = source.MinLength
	destination.MaxLength = source.MaxLength
	destination.Pattern = source.Pattern
	destination.compiledPattern = source.compiledPattern
	destination.MinItems = source.MinItems
	destination.MaxItems = source.MaxItems
	destination.Items = source.Items
	destination.Required = source.Required
	destination.Properties = source.Properties
	destination.MinProps = source.MinProps
	destination.MaxProps = source.MaxProps
	destination.AdditionalProperties = source.AdditionalProperties
	destination.Discriminator = source.Discriminator
	return destination
}
