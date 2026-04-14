package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestOAS31_RefSiblingKeyword verifies that sibling keywords alongside $ref are honoured
// when loading an OpenAPI 3.1 document.
//
// In OpenAPI 3.0 / JSON Schema draft-07, $ref replaces its entire object so any sibling
// keywords (e.g. deprecated, description) are silently ignored.
// In OpenAPI 3.1 / JSON Schema 2020-12, $ref and sibling keywords are both applied, so
// a property like:
//
//	status:
//	  deprecated: true
//	  $ref: "#/components/schemas/PingStatus"
//
// should result in a SchemaRef whose Value has Deprecated==true.
func TestOAS31_RefSiblingKeyword(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromFile("testdata/schema31-ref-siblings.yml")
	require.NoError(t, err)

	pingResp := doc.Components.Schemas["PingResponse"].Value
	require.NotNil(t, pingResp)

	statusRef := pingResp.Properties["status"]
	require.NotNil(t, statusRef)

	// The $ref should still be resolved.
	require.NotNil(t, statusRef.Value, "$ref to PingStatus should be resolved")
	require.Equal(t, "string", statusRef.Value.Type.Slice()[0], "$ref target type should be string")

	// The sibling deprecated:true must survive — not be discarded because $ref is present.
	require.True(t, statusRef.Value.Deprecated, "deprecated:true sibling to $ref must be honoured in OAS 3.1")
}

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
	require.NotNil(t, unItems.UnevaluatedItems.Schema)
	require.Equal(t, "#/components/schemas/StringType", unItems.UnevaluatedItems.Schema.Ref)
	require.NotNil(t, unItems.UnevaluatedItems.Schema.Value, "unevaluatedItems $ref should be resolved")

	// unevaluatedProperties ref should be resolved
	unProps := schemas["ObjectWithUnevaluatedProperties"].Value
	require.NotNil(t, unProps)
	require.NotNil(t, unProps.UnevaluatedProperties.Schema)
	require.Equal(t, "#/components/schemas/StringType", unProps.UnevaluatedProperties.Schema.Ref)
	require.NotNil(t, unProps.UnevaluatedProperties.Schema.Value, "unevaluatedProperties $ref should be resolved")

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
