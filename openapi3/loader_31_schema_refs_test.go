package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveSchemaRefsIn31Fields(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromFile("testdata/schema31refs.yml")
	require.NoError(t, err)

	schemas := doc.Components.Schemas

	// prefixItems refs should be resolved
	tupleArray := schemas["TupleArray"].Value
	require.NotNil(t, tupleArray)
	require.Len(t, tupleArray.PrefixItems, 2)
	require.Equal(t, "#/components/schemas/StringType", tupleArray.PrefixItems[0].Ref)
	require.NotNil(t, tupleArray.PrefixItems[0].Value, "prefixItems[0] $ref should be resolved")
	require.Equal(t, "string", tupleArray.PrefixItems[0].Value.Type.Slice()[0])
	require.Equal(t, "#/components/schemas/IntegerType", tupleArray.PrefixItems[1].Ref)
	require.NotNil(t, tupleArray.PrefixItems[1].Value, "prefixItems[1] $ref should be resolved")
	require.Equal(t, "integer", tupleArray.PrefixItems[1].Value.Type.Slice()[0])

	// contains ref should be resolved
	arrayContains := schemas["ArrayWithContains"].Value
	require.NotNil(t, arrayContains)
	require.Equal(t, "#/components/schemas/StringType", arrayContains.Contains.Ref)
	require.NotNil(t, arrayContains.Contains.Value, "contains $ref should be resolved")
	require.Equal(t, "string", arrayContains.Contains.Value.Type.Slice()[0])

	// patternProperties refs should be resolved
	patternProps := schemas["ObjectWithPatternProperties"].Value
	require.NotNil(t, patternProps)
	pp := patternProps.PatternProperties["^x-"]
	require.NotNil(t, pp)
	require.Equal(t, "#/components/schemas/StringType", pp.Ref)
	require.NotNil(t, pp.Value, "patternProperties $ref should be resolved")

	// dependentSchemas refs should be resolved
	depSchemas := schemas["ObjectWithDependentSchemas"].Value
	require.NotNil(t, depSchemas)
	ds := depSchemas.DependentSchemas["name"]
	require.NotNil(t, ds)
	require.Equal(t, "#/components/schemas/NonNegative", ds.Ref)
	require.NotNil(t, ds.Value, "dependentSchemas $ref should be resolved")

	// propertyNames ref should be resolved
	propNames := schemas["ObjectWithPropertyNames"].Value
	require.NotNil(t, propNames)
	require.Equal(t, "#/components/schemas/NamePattern", propNames.PropertyNames.Ref)
	require.NotNil(t, propNames.PropertyNames.Value, "propertyNames $ref should be resolved")

	// unevaluatedItems ref should be resolved
	unItems := schemas["ArrayWithUnevaluatedItems"].Value
	require.NotNil(t, unItems)
	require.Equal(t, "#/components/schemas/StringType", unItems.UnevaluatedItems.Ref)
	require.NotNil(t, unItems.UnevaluatedItems.Value, "unevaluatedItems $ref should be resolved")

	// unevaluatedProperties ref should be resolved
	unProps := schemas["ObjectWithUnevaluatedProperties"].Value
	require.NotNil(t, unProps)
	require.Equal(t, "#/components/schemas/StringType", unProps.UnevaluatedProperties.Ref)
	require.NotNil(t, unProps.UnevaluatedProperties.Value, "unevaluatedProperties $ref should be resolved")

	// if/then/else refs should be resolved
	ifThenElse := schemas["ObjectWithIfThenElse"].Value
	require.NotNil(t, ifThenElse)
	require.Equal(t, "#/components/schemas/StringType", ifThenElse.If.Ref)
	require.NotNil(t, ifThenElse.If.Value, "if $ref should be resolved")
	require.Equal(t, "string", ifThenElse.If.Value.Type.Slice()[0])
	require.Equal(t, "#/components/schemas/IntegerType", ifThenElse.Then.Ref)
	require.NotNil(t, ifThenElse.Then.Value, "then $ref should be resolved")
	require.Equal(t, "integer", ifThenElse.Then.Value.Type.Slice()[0])
	require.Equal(t, "#/components/schemas/NonNegative", ifThenElse.Else.Ref)
	require.NotNil(t, ifThenElse.Else.Value, "else $ref should be resolved")

	// contentSchema ref should be resolved
	contentSchema := schemas["StringWithContentSchema"].Value
	require.NotNil(t, contentSchema)
	require.Equal(t, "#/components/schemas/NonNegative", contentSchema.ContentSchema.Ref)
	require.NotNil(t, contentSchema.ContentSchema.Value, "contentSchema $ref should be resolved")
}
