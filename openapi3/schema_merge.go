package openapi3

import (
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
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
func Merge(schema Schema) (*Schema, error) {
	if !isListOfObjects(&schema) {
		return &schema, nil
	}

	if schema.AllOf != nil {
		mergedAllOf, err := mergeAllOf(schema.AllOf)
		if err != nil {
			return &Schema{}, err
		}
		schema.AllOf = nil
		result, err := mergeFields([]*Schema{&schema, mergedAllOf})
		if err != nil {
			log.Fatal(err.Error())
		}
		return result, nil
	}

	// handle cases where AllOf is nil, but other fields might include AllOf.
	result, err := handleNestedAllOfCases(&schema)
	if err != nil {
		return &Schema{}, err
	}

	return result, nil
}

func handleNestedAllOfCases(schema *Schema) (*Schema, error) {
	if schema.Properties != nil {
		properties, err := mergeProperties(schema.Properties)
		if err != nil {
			return &Schema{}, err
		}
		schema.Properties = properties
	}

	if schema.AnyOf != nil {
		var mergedAnyOf SchemaRefs
		for _, schemaRef := range schema.AnyOf {
			if schemaRef == nil {
				continue
			}
			result, err := Merge(*schemaRef.Value)
			if err != nil {
				return &Schema{}, err
			}
			mergedAnyOf = append(mergedAnyOf, &SchemaRef{
				Value: result,
			})
		}
		schema.AnyOf = mergedAnyOf
	}

	if schema.OneOf != nil {
		var mergedOneOf SchemaRefs
		for _, schemaRef := range schema.OneOf {
			if schemaRef == nil {
				continue
			}
			result, err := Merge(*schemaRef.Value)
			if err != nil {
				return &Schema{}, err
			}
			mergedOneOf = append(mergedOneOf, &SchemaRef{
				Value: result,
			})
		}
		schema.AnyOf = mergedOneOf
	}

	if schema.Not != nil {
		result, err := Merge(*schema.Not.Value)
		if err != nil {
			return &Schema{}, err
		}
		schema.Not = &SchemaRef{
			Value: result,
		}
	}

	return schema, nil
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
	result.Enum = resolveEnum(collection.Enum)
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
	result, err = resolveAdditionalProperties(result, &collection)
	if err != nil {
		return result, err
	}
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
		merged, err := Merge(*schema.Value)
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

func getPropFieldsToMerge(collection *SchemaCollection) []string {
	properties := [][]string{}
	for i, schema := range collection.Properties {
		additionalProps := collection.AdditionalProperties[i].Has
		if additionalProps != nil && !*additionalProps {
			keys := []string{}
			for key := range schema {
				keys = append(keys, key)
			}
			properties = append(properties, keys)
		}
	}
	if len(properties) > 0 {
		return findIntersection(properties...)
	}
	keys := []string{}
	for _, schema := range collection.Properties {
		for key := range schema {
			keys = append(keys, key)
		}
	}
	return keys
}

func resolveProperties(schema *Schema, collection *SchemaCollection) (*Schema, error) {
	keys := getPropFieldsToMerge(collection)
	allRefs := map[string][]*Schema{}              //naming
	for _, schema := range collection.Properties { //naming
		for name, schemaRef := range schema {
			if containsString(keys, name) {
				allRefs[name] = append(allRefs[name], schemaRef.Value)
			}
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

	schema.Properties = result
	return schema, nil
}

func resolveAdditionalProperties(schema *Schema, collection *SchemaCollection) (*Schema, error) {
	additionalProperties := &AdditionalProperties{
		Has:    nil,
		Schema: nil,
	}

	additionalSchemas := []*Schema{}
	for _, ap := range collection.AdditionalProperties {
		if ap.Has != nil && !*ap.Has {
			hasValue := false
			additionalProperties.Has = &hasValue
			schema.AdditionalProperties = *additionalProperties
			return schema, nil
		}
		if ap.Schema != nil && ap.Schema.Value != nil {
			additionalSchemas = append(additionalSchemas, ap.Schema.Value)
		}
	}

	if len(additionalSchemas) > 0 {
		result, err := mergeFields(additionalSchemas)
		if err != nil {
			return schema, err
		}
		additionalProperties.Schema = &SchemaRef{
			Value: result,
		}
	}

	schema.AdditionalProperties = *additionalProperties
	return schema, nil
}

func resolveEnum(values [][]interface{}) []interface{} {
	var nonEmptyEnum [][]interface{}
	for _, enum := range values {
		if len(enum) > 0 {
			nonEmptyEnum = append(nonEmptyEnum, enum)
		}
	}
	return findIntersectionOfArrays(nonEmptyEnum)
}

func resolvePattern(values []string) string {
	var pattern strings.Builder
	for _, p := range values {
		if len(p) > 0 {
			if !isPatternResolved(p) {
				pattern.WriteString(fmt.Sprintf("(?=%s)", p))
			} else {
				pattern.WriteString(p)
			}
		}
	}
	return pattern.String()
}

func isPatternResolved(pattern string) bool {
	match, _ := regexp.MatchString(`^\(\?=.+\)$`, pattern)
	return match
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

func containsString(list []string, search string) bool {
	for _, item := range list {
		if item == search {
			return true
		}
	}
	return false
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
	refs := filterNilSchemaRef(collection.Not)
	if len(refs) == 0 {
		return schema
	}
	schema.Not = &SchemaRef{
		Value: &Schema{
			AnyOf: refs,
		},
	}
	return schema
}

func filterNilSchemaRef(refs []*SchemaRef) []*SchemaRef {
	result := []*SchemaRef{}
	for _, v := range refs {
		if v != nil {
			result = append(result, v)
		}
	}
	return result
}

func filterEmptySchemaRefs(groups []SchemaRefs) []SchemaRefs {
	result := []SchemaRefs{}
	for _, group := range groups {
		if len(group) > 0 {
			result = append(result, group)
		}
	}
	return result
}

func resolveAnyOf(schema *Schema, collection *SchemaCollection) (*Schema, error) {
	groups := filterEmptySchemaRefs(collection.AnyOf)
	if len(groups) == 0 {
		return schema, nil
	}
	combinations := getCombinations(groups)
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
	groups := filterEmptySchemaRefs(collection.OneOf)
	if len(groups) == 0 {
		return schema, nil
	}
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

func findIntersection(arrays ...[]string) []string {
	if len(arrays) == 0 {
		return nil
	}

	// Create a map to store the elements of the first array
	elementsMap := make(map[string]bool)
	for _, element := range arrays[0] {
		elementsMap[element] = true
	}

	// Iterate through the remaining arrays and update the map
	for _, arr := range arrays[1:] {
		tempMap := make(map[string]bool)
		for _, element := range arr {
			if elementsMap[element] {
				tempMap[element] = true
			}
		}
		elementsMap = tempMap
	}

	intersection := []string{}
	for element := range elementsMap {
		intersection = append(intersection, element)
	}

	return intersection
}
