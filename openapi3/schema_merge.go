package openapi3

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

const (
	FormatErrorMessage = "unable to resolve Format conflict: all Format values must be identical"
	TypeErrorMessage   = "unable to resolve Type conflict: all Type values must be identical"
)

type SchemaCollection struct {
	Not                  []*SchemaRef
	OneOf                []SchemaRefs
	AnyOf                []SchemaRefs
	Title                []string
	Type                 []string
	Format               []string
	Description          []string
	Enum                 [][]interface{}
	UniqueItems          []bool
	ExclusiveMin         []bool
	ExclusiveMax         []bool
	Min                  []*float64
	Max                  []*float64
	MultipleOf           []*float64
	MinLength            []uint64
	MaxLength            []*uint64
	Pattern              []string
	MinItems             []uint64
	MaxItems             []*uint64
	Items                []*SchemaRef
	Required             [][]string
	Properties           []Schemas
	MinProps             []uint64
	MaxProps             []*uint64
	AdditionalProperties []AdditionalProperties
}

// Merge replaces objects under AllOf with a flattened equivalent
func Merge(schema *Schema) (*Schema, error) {
	if !isListOfObjects(schema) {
		return schema, nil
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

	return schema, nil
}

func mergeProperties(schemas Schemas) (Schemas, error) {
	res := make(Schemas)
	for name, schemaRef := range schemas {
		merged, err := Merge(schemaRef.Value)
		if err != nil {
			return res, err
		}
		schemaRef.Value = merged
		res[name] = schemaRef
	}
	return res, nil
}

func mergeFields(schemas []*Schema) (*Schema, error) {
	result := NewSchema()
	collection := collect(schemas)
	result.Title = collection.Title[0]
	result.Description = collection.Description[0]
	format, err := resolveFormat(collection.Format)
	if err != nil {
		return result, err
	}
	result.Format = format
	stype, err := resolveType(collection.Type)
	if err != nil {
		return result, err
	}
	result.Type = stype
	result = resolveNumberRange(result, &collection)
	result.MinLength = findMaxValue(collection.MinLength)
	result.MaxLength = findMinValue(collection.MaxLength)
	result.MinItems = findMaxValue(collection.MinItems)
	result.MaxItems = findMinValue(collection.MaxItems)
	result.MinProps = findMaxValue(collection.MinProps)
	result.MaxProps = findMinValue(collection.MaxProps)
	result.Pattern = resolvePattern(collection.Pattern)
	result.Enum = resolveEnum(collection.Enum) //todo: handle nil enums? (empty arrays)
	result = resolveMultipleOf(result, &collection)
	result.Required = resolveRequired(collection.Required)
	result, err = resolveItems(result, &collection)
	if err != nil {
		return result, err
	}
	result.UniqueItems = resolveUniqueItems(collection.UniqueItems)
	result, err = resolveProperties(result, &collection)
	if err != nil {
		return result, err
	}

	result, err = resolveOneOf(result, &collection)
	if err != nil {
		return result, err
	}

	result, err = resolveAnyOf(result, &collection)
	if err != nil {
		return result, err
	}

	result = resolveNot(result, &collection)
	return result, nil
}

func resolveNumberRange(schema *Schema, collection *SchemaCollection) *Schema {

	//resolve minimum
	max := math.Inf(-1)
	isExcluded := false
	var value *float64
	for i, s := range collection.Min {
		if s != nil {
			if *s > max {
				max = *s
				value = s
				isExcluded = collection.ExclusiveMin[i]
			}
		}
	}

	schema.Min = value
	schema.ExclusiveMin = isExcluded
	//resolve maximum
	min := math.Inf(1)
	isExcluded = false
	// var value *float64
	for i, s := range collection.Max {
		if s != nil {
			if *s < min {
				min = *s
				value = s
				isExcluded = collection.ExclusiveMax[i]
			}
		}
	}

	schema.Max = value
	schema.ExclusiveMax = isExcluded
	return schema
}

func mergeAllOf(allOf SchemaRefs) (*Schema, error) {

	schemas := []*Schema{}
	for _, schema := range allOf {
		merged, err := Merge(schema.Value)
		if err != nil {
			return &Schema{}, err
		}
		schemas = append(schemas, merged)
	}

	schema, err := mergeFields(schemas)
	if err != nil {
		return schema, err
	}
	return schema, nil
}

func resolveItems(schema *Schema, collection *SchemaCollection) (*Schema, error) {
	items := []*Schema{}
	for _, s := range collection.Items {
		if s != nil {
			items = append(items, s.Value)
		}
	}
	if len(items) == 0 {
		schema.Items = nil
		return schema, nil
	}

	res, err := mergeFields(items)
	if err != nil {
		return schema, err
	}
	ref := SchemaRef{
		Value: res,
	}
	schema.Items = &ref
	return schema, nil
}

func resolveUniqueItems(values []bool) bool {
	for _, v := range values {
		if v {
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

func resolveMultipleOf(schema *Schema, collection *SchemaCollection) *Schema {
	values := []float64{}
	for _, v := range collection.MultipleOf {
		if v == nil {
			continue
		}
		values = append(values, *v)
	}
	if len(values) == 0 {
		schema.MultipleOf = nil
		return schema
	}

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
	schema.MultipleOf = Float64Ptr(float64(lcmValue) / factor)
	return schema
}

func resolveProperties(schema *Schema, collection *SchemaCollection) (*Schema, error) {
	propRefs := append([]Schemas{}, collection.Properties...)
	allRefs := map[string][]*Schema{} //naming
	for _, schema := range propRefs { //naming
		for name, schemaRef := range schema {
			allRefs[name] = append(allRefs[name], schemaRef.Value)
		}
	}
	result := make(Schemas)
	for name, schemas := range allRefs {
		merged, err := mergeFields(schemas)
		if err != nil {
			schema.Properties = nil
			return schema, err
		}
		ref := SchemaRef{
			Value: merged,
		}
		result[name] = &ref
	}
	if len(result) == 0 {
		result = nil
	}

	result, additionalProperties := mergeAdditionalProps(collection, result)
	schema.AdditionalProperties = additionalProperties
	schema.Properties = result
	return schema, nil
}

func mergeAdditionalProps(collection *SchemaCollection, propsMap Schemas) (Schemas, AdditionalProperties) {
	additionalProperties := &AdditionalProperties{
		Has:    nil,
		Schema: nil,
	}
	for i, additionalProps := range collection.AdditionalProperties {
		if additionalProps.Has == nil {
			continue
		}
		if !*additionalProps.Has {
			for prop := range propsMap {
				found := false
				for key := range collection.Properties[i] {
					if prop == key {
						found = true
					}
				}
				if !found {
					delete(propsMap, prop)
				}
			}
			f := false
			additionalProperties.Has = &f
			return propsMap, *additionalProperties
		} else {
			t := true
			additionalProperties.Has = &t
		}
	}
	return propsMap, *additionalProperties
}

func resolveEnum(values [][]interface{}) []interface{} {
	return findIntersectionOfArrays(values)
}

func resolvePattern(values []string) string {
	var pattern strings.Builder
	for _, p := range values {
		if len(p) > 0 {
			pattern.WriteString(fmt.Sprintf("(?=%s)", p))
		}
	}
	return pattern.String()
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

func findMinValue(values []*uint64) *uint64 {
	dvalues := []uint64{}
	for _, v := range values {
		if v != nil {
			dvalues = append(dvalues, *v)
		}
	}
	if len(dvalues) == 0 {
		return nil
	}
	min := uint64(math.MaxUint64)
	for _, num := range dvalues {
		if num < min {
			min = num
		}
	}
	return Uint64Ptr(min)
}

func resolveType(values []string) (string, error) {
	values = filterEmptyStrings(values)
	if len(values) == 0 {
		return "", nil
	}
	if allStringsEqual(values) {
		return values[0], nil
	}
	return values[0], errors.New(TypeErrorMessage)
}

func resolveFormat(values []string) (string, error) {
	values = filterEmptyStrings(values)
	if len(values) == 0 {
		return "", nil
	}
	if allStringsEqual(values) {
		return values[0], nil
	}
	return values[0], errors.New(FormatErrorMessage)
}

func filterEmptyStrings(input []string) []string {
	var result []string

	for _, s := range input {
		if s != "" {
			result = append(result, s)
		}
	}

	return result
}

func isListOfObjects(schema *Schema) bool {
	if schema == nil || schema.AllOf == nil {
		return false
	}

	// for _, subSchema := range schema.AllOf {
	// 	if subSchema.Value.Type != "object" {
	// 		return false
	// 	}
	// }

	return true
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

func getIntersection(arr1, arr2 []interface{}) []interface{} {
	intersectionMap := make(map[interface{}]bool)
	result := []interface{}{}

	// Mark elements in the first array
	for _, val := range arr1 {
		intersectionMap[val] = true
	}

	// Check if elements in the second array exist in the intersection map
	for _, val := range arr2 {
		if intersectionMap[val] {
			result = append(result, val)
		}
	}

	return result
}

func findIntersectionOfArrays(arrays [][]interface{}) []interface{} {
	if len(arrays) == 0 {
		return nil
	}

	intersection := arrays[0]

	for i := 1; i < len(arrays); i++ {
		intersection = getIntersection(intersection, arrays[i])
	}
	if len(intersection) == 0 {
		return nil
	}
	return intersection
}

func flattenArray(arrays [][]string) []string {
	var result []string

	for i := 0; i < len(arrays); i++ {
		for j := 0; j < len(arrays[i]); j++ {
			result = append(result, arrays[i][j])
		}
	}

	return result
}

func resolveRequired(values [][]string) []string {
	flatValues := flattenArray(values)
	uniqueMap := make(map[string]bool)
	var uniqueValues []string
	for _, str := range flatValues {
		if _, found := uniqueMap[str]; !found {
			uniqueMap[str] = true
			uniqueValues = append(uniqueValues, str)
		}
	}
	return uniqueValues
}

/* temporary */
func copy(source *Schema, destination *Schema) *Schema {
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

func collect(schemas []*Schema) SchemaCollection {
	collection := SchemaCollection{}
	for _, s := range schemas {
		collection.Not = append(collection.Not, s.Not)
		collection.AnyOf = append(collection.AnyOf, s.AnyOf)
		collection.OneOf = append(collection.OneOf, s.OneOf)
		collection.Title = append(collection.Title, s.Title)
		collection.Type = append(collection.Type, s.Type)
		collection.Format = append(collection.Format, s.Format)
		collection.Description = append(collection.Description, s.Description)
		collection.Enum = append(collection.Enum, s.Enum)
		collection.UniqueItems = append(collection.UniqueItems, s.UniqueItems)
		collection.ExclusiveMin = append(collection.ExclusiveMin, s.ExclusiveMin)
		collection.ExclusiveMax = append(collection.ExclusiveMax, s.ExclusiveMax)
		collection.Min = append(collection.Min, s.Min)
		collection.Max = append(collection.Max, s.Max)
		collection.MultipleOf = append(collection.MultipleOf, s.MultipleOf)
		collection.MinLength = append(collection.MinLength, s.MinLength)
		collection.MaxLength = append(collection.MaxLength, s.MaxLength)
		collection.Pattern = append(collection.Pattern, s.Pattern)
		collection.MinItems = append(collection.MinItems, s.MinItems)
		collection.MaxItems = append(collection.MaxItems, s.MaxItems)
		collection.Items = append(collection.Items, s.Items)
		collection.Required = append(collection.Required, s.Required)
		collection.Properties = append(collection.Properties, s.Properties)
		collection.MinProps = append(collection.MinProps, s.MinProps)
		collection.MaxProps = append(collection.MaxProps, s.MaxProps)
		collection.AdditionalProperties = append(collection.AdditionalProperties, s.AdditionalProperties)
	}
	return collection
}

// getCombinations calculates the cartesian product of groups of SchemaRefs.
func getCombinations(groups []SchemaRefs) []SchemaRefs {
	if len(groups) == 0 {
		return []SchemaRefs{}
	}
	result := []SchemaRefs{{}}
	for _, group := range groups {
		var newResult []SchemaRefs
		for _, resultItem := range result {
			for _, ref := range group {
				combination := append(SchemaRefs{}, resultItem...)
				combination = append(combination, ref)
				newResult = append(newResult, combination)
			}
		}
		result = newResult
	}
	return result
}

func mergeCombinations(combinations []SchemaRefs) ([]*Schema, error) {
	merged := []*Schema{}
	for _, combination := range combinations {
		schemas := []*Schema{}
		for _, ref := range combination {
			schemas = append(schemas, ref.Value)
		}
		schema, err := mergeFields(schemas)
		merged = append(merged, schema)

		//todo: if error is nil, do not add merge, and continue to iterate.
		if err != nil {
			return merged, err
		}
	}
	return merged, nil
}

func resolveNot(schema *Schema, collection *SchemaCollection) *Schema {
	refs := []*SchemaRef{}
	for _, v := range collection.Not {
		if v != nil {
			refs = append(refs, v)
		}
	}
	if len(refs) == 0 {
		return schema
	}
	schema.Not = &SchemaRef{
		Value: &Schema{
			AnyOf: collection.Not,
		},
	}
	return schema
}

func resolveAnyOf(schema *Schema, collection *SchemaCollection) (*Schema, error) {
	combinations := getCombinations(collection.AnyOf)
	if len(combinations) == 0 {
		return schema, nil
	}
	refs, err := resolveGroups(combinations)
	if err != nil {
		return schema, err
	}
	schema.AnyOf = refs
	return schema, nil
}

func resolveOneOf(schema *Schema, collection *SchemaCollection) (*Schema, error) {
	combinations := getCombinations(collection.OneOf)
	if len(combinations) == 0 {
		return schema, nil
	}
	refs, err := resolveGroups(combinations)
	if err != nil {
		return schema, err
	}
	schema.OneOf = refs
	return schema, nil
}

func resolveGroups(combinations []SchemaRefs) (SchemaRefs, error) {
	mergedCombinations, err := mergeCombinations(combinations)
	if err != nil {
		return nil, err
	}

	var refs SchemaRefs
	for _, merged := range mergedCombinations {
		refs = append(refs, &SchemaRef{
			Value: merged,
		})
	}
	return refs, nil
}
