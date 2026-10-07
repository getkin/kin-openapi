package openapi3

import "testing"

func TestSchemaUsesJSONSchema2020Features(t *testing.T) {
	falseValue := false
	ref := func() *SchemaRef { return &SchemaRef{Value: &Schema{}} }

	// Keep this table in sync with the keywords guarding the legacy fallback.
	tests := []struct {
		name   string
		schema *Schema
	}{
		{"const", &Schema{Const: false}},
		{"prefixItems", &Schema{PrefixItems: SchemaRefs{ref()}}},
		{"contains", &Schema{Contains: ref()}},
		{"patternProperties", &Schema{PatternProperties: Schemas{"^x": ref()}}},
		{"dependentSchemas", &Schema{DependentSchemas: Schemas{"x": ref()}}},
		{"propertyNames", &Schema{PropertyNames: ref()}},
		{"unevaluatedItems false", &Schema{UnevaluatedItems: BoolSchema{Has: &falseValue}}},
		{"unevaluatedItems schema", &Schema{UnevaluatedItems: BoolSchema{Schema: ref()}}},
		{"unevaluatedProperties false", &Schema{UnevaluatedProperties: BoolSchema{Has: &falseValue}}},
		{"unevaluatedProperties schema", &Schema{UnevaluatedProperties: BoolSchema{Schema: ref()}}},
		{"if", &Schema{If: ref()}},
		{"then", &Schema{Then: ref()}},
		{"else", &Schema{Else: ref()}},
		{"dependentRequired", &Schema{DependentRequired: map[string][]string{"x": {"y"}}}},
		{"defs", &Schema{Defs: Schemas{"x": ref()}}},
		{"content schema", &Schema{ContentSchema: ref()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !schemaUsesJSONSchema2020Features(tt.schema) {
				t.Fatal("3.1 field was not detected")
			}
		})
	}
}

func TestSchemaUsesJSONSchema2020FeaturesNestedAndLegacy(t *testing.T) {
	falseValue := false
	zeroCount := uint64(0)
	zeroNumber := float64(0)
	legacy := &Schema{
		Nullable:             true,
		ExclusiveMin:         ExclusiveBound{Bool: &falseValue},
		AdditionalProperties: BoolSchema{Has: &falseValue},
		Default:              map[string]any{"if": "ordinary data"},
		Examples:             []any{map[string]any{"contains": "ordinary data"}},
		MinContains:          &zeroCount,
		ExclusiveMax:         ExclusiveBound{Value: &zeroNumber},
	}
	if schemaUsesJSONSchema2020Features(legacy) {
		t.Fatal("other fields and data values must not change legacy fallback behavior")
	}

	root := &Schema{Properties: Schemas{
		"nested": {Ref: "#/components/schemas/Nested", Value: &Schema{
			UnevaluatedProperties: BoolSchema{Has: &falseValue},
		}},
	}}
	if !schemaUsesJSONSchema2020Features(root) {
		t.Fatal("a 3.1 field on a resolved reference must be detected")
	}

	cyclic := &Schema{}
	cyclic.Properties = Schemas{"self": {Value: cyclic}}
	if schemaUsesJSONSchema2020Features(cyclic) {
		t.Fatal("a cycle without 3.1 fields must not be detected")
	}
}
