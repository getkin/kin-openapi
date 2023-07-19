package openapi3

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
)

// Merge replaces objects under AllOf with a flattened equivalent
func Merge(schema Schema) *Schema {
	if !isListOfObjects(&schema) {
		return &schema
	}
	if schema.AllOf != nil {
		mergedAllOf := mergeAllOf(schema.AllOf)
		schema = copy(mergedAllOf, schema) // temporary.
	}
	schema.AllOf = nil

	//todo: merge result of mergedAllOf with all other fields of base Schema.
	//todo: implement merge functions for OneOf, AnyOf, Items
	if schema.Properties != nil {
		schema.Properties = mergeProperties(schema.Properties)
	}
	return &schema
}

func mergeAllOf(allOf SchemaRefs) Schema {

	schemas := make([]Schema, 0) // naming
	for _, schema := range allOf {
		schemas = append(schemas, *Merge(*schema.Value))
	}

	schema := mergeFields(schemas)
	return *schema
}

func mergeProperties(schemas Schemas) Schemas {
	res := make(Schemas)
	for name, schemaRef := range schemas {
		schemaRef.Value = Merge(*schemaRef.Value)
		res[name] = schemaRef
	}
	return res
}

func mergeFields(schemas []Schema) *Schema {
	result := NewSchema()
	titles := getStringValues(schemas, "title")
	if len(titles) > 0 {
		result.Title = titleResolver(titles)
	}

	required := getStringValues(schemas, "required")
	if len(required) > 0 {
		result.Required = requiredResolver(required)
	}

	description := getStringValues(schemas, "description")
	if len(description) > 0 {
		result.Description = resolveDescriptions(description)
	}

	formats := getStringValues(schemas, "format")
	if len(formats) > 0 {
		res, err := resolveFormat(formats)
		if err != nil {
			log.Fatal(err.Error())
		}
		result.Format = res
	}

	types := getStringValues(schemas, "type")
	if len(types) > 0 {
		res, err := resolveType(types)
		if err != nil {
			log.Fatal(err.Error())
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

	minimum := getFloat64Values(schemas, "min")
	if len(minimum) > 0 {
		result.Min = Float64Ptr(resolveMinimum(minimum))
	}

	maximum := getFloat64Values(schemas, "max")
	if len(maximum) > 0 {
		result.Max = Float64Ptr(resolveMaximum(maximum))
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
		result.Properties = resolveProperties(properties)
	}

	enum := getEnum(schemas, "enum")
	if len(enum) > 0 {
		res, err := resolveEnum(enum)
		if err != nil {
			log.Fatal(err.Error())
		}
		result.Enum = res
	}
	return result
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

func resolveProperties(schemas []Schemas) Schemas {
	allRefs := map[string][]Schema{} //naming
	for _, schema := range schemas { //naming
		for name, schemaRef := range schema {
			allRefs[name] = append(allRefs[name], *schemaRef.Value)
		}
	}
	result := make(Schemas)
	for name, schemas := range allRefs {
		ref := SchemaRef{
			Value: mergeFields(schemas),
		}
		result[name] = &ref
	}
	return result
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

func resolveEnum(values []interface{}) ([]interface{}, error) {
	if areAllUnique(values) {
		return values, nil
	} else {
		return nil, errors.New("could not resovle Enum conflict - all Enum values must be unique")
	}
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

func resolveMaximum(values []float64) float64 {
	min := math.Inf(1)
	for _, value := range values {
		if value < min {
			min = value
		}
	}
	return min
}

func resolveMinimum(values []float64) float64 {
	max := math.Inf(-1)
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}

func resolveDescriptions(values []string) string {
	return values[0]
}

func resolveType(values []string) (string, error) {
	if allStringsEqual(values) {
		return values[0], nil
	}
	return values[0], errors.New("could not resovle Type conflict - all Type values must be identical")
}

func resolveFormat(values []string) (string, error) {
	if allStringsEqual(values) {
		return values[0], nil
	}
	return values[0], errors.New("could not resovle Format conflict - all Format values must be identical")
}

func titleResolver(values []string) string {
	return values[0]
}

func requiredResolver(values []string) []string {
	return values
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

func allStringsEqual(values []string) bool {
	first := values[0]
	for _, value := range values {
		if first != value {
			return false
		}
	}
	return true
}

func areAllUnique(values []interface{}) bool {
	occurrenceMap := make(map[interface{}]bool)
	for _, item := range values {
		if occurrenceMap[item] {
			return false
		}
		occurrenceMap[item] = true
	}
	return true
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
