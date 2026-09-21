package openapi3_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
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
func TestSchemaRefSiblingKeyword(t *testing.T) {
	spec := `
openapi: "3.1.0"
info:
  title: Ref Sibling Test
  version: "1.0"
paths:
  /ping:
    get:
      operationId: getPing
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/PingResponse"
components:
  schemas:
    PingStatus:
      type: string
      enum: [ok, error]
    PingResponse:
      type: object
      required: [message, status]
      properties:
        message:
          type: string
        status:
          deprecated: true  # sibling keyword alongside $ref — valid in OAS 3.1, ignored in 3.0
          $ref: "#/components/schemas/PingStatus"
`

	type testcase struct {
		oas             string
		siblings, valid bool
	}

	for _, tc := range []testcase{
		{oas: "3.1", siblings: true, valid: true}, {oas: "3.0"}, {oas: "3.0", valid: true},
	} {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			t.Parallel()
			loader := openapi3.NewLoader()

			doc, err := loader.LoadFromData([]byte(strings.ReplaceAll(spec, "3.1.0", tc.oas)))
			require.NoError(t, err)

			statusRef := doc.Components.Schemas["PingResponse"].Value.Properties["status"]
			require.NotNil(t, statusRef)

			// The $ref should still be resolved. In OAS 3.1 the local sibling
			// schema is the effective value and the target is a conjunct.
			require.Equal(t, statusRef.Ref, "#/components/schemas/PingStatus")
			require.NotNil(t, statusRef.Value, "$ref to PingStatus should be resolved")
			if tc.siblings {
				require.True(t, statusRef.Value.Deprecated)
				require.Len(t, statusRef.Value.AllOf, 1)
				require.Equal(t, "string", statusRef.Value.AllOf[0].Value.Type.Slice()[0])
			} else {
				require.False(t, statusRef.Value.Deprecated)
				require.Equal(t, "string", statusRef.Value.Type.Slice()[0])
			}

			var valopts []openapi3.ValidationOption
			if tc.valid && !tc.siblings { // For this test case let's try the option that allows siblings for 3.0
				valopts = append(valopts, openapi3.AllowExtraSiblingFields("deprecated"))
			}
			err = doc.Validate(loader.Context, valopts...)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err, "Siblings to $ref is not valid OpenAPIv3.0 (by default)")
			}
		})
	}
}

func TestSchemaRefSiblingIsSerializedAndAppliedAsConjunct(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.1.0
info:
  title: ref sibling
  version: "1"
paths: {}
components:
  schemas:
    Base:
      type: string
      maxLength: 5
    Use:
      $ref: '#/components/schemas/Base'
      minLength: 3
      x-use-site: true
`))
	require.NoError(t, err)

	useRef := doc.Components.Schemas["Use"]
	require.Equal(t, uint64(3), useRef.Value.MinLength)
	require.Len(t, useRef.Value.AllOf, 1)
	require.Equal(t, uint64(5), *useRef.Value.AllOf[0].Value.MaxLength)
	require.NotContains(t, doc.Components.Schemas["Base"].Value.Extensions, "x-use-site")
	require.Error(t, useRef.Value.VisitJSON("xx", openapi3.EnableJSONSchema2020()))
	require.Error(t, useRef.Value.VisitJSON("xxxxxx", openapi3.EnableJSONSchema2020()))
	require.NoError(t, useRef.Value.VisitJSON("xxxx", openapi3.EnableJSONSchema2020()))

	data, err := json.Marshal(useRef)
	require.NoError(t, err)
	require.JSONEq(t, `{
  "$ref":"#/components/schemas/Base",
  "minLength":3,
  "x-use-site":true
}`, string(data))
	documentData, err := json.Marshal(doc)
	require.NoError(t, err)
	roundTripLoader := openapi3.NewLoader()
	roundTripped, err := roundTripLoader.LoadFromData(documentData)
	require.NoError(t, err)
	roundTrippedUse := roundTripped.Components.Schemas["Use"].Value
	require.Error(t, roundTrippedUse.VisitJSON("xx", openapi3.EnableJSONSchema2020()))
	require.Error(t, roundTrippedUse.VisitJSON("xxxxxx", openapi3.EnableJSONSchema2020()))
	require.NoError(t, roundTrippedUse.VisitJSON("xxxx", openapi3.EnableJSONSchema2020()))

	useRef.Extensions["x-use-site"] = "changed"
	useRef.Extensions["x-added"] = true
	data, err = json.Marshal(useRef)
	require.NoError(t, err)
	require.JSONEq(t, `{
  "$ref":"#/components/schemas/Base",
  "minLength":3,
  "x-use-site":"changed",
  "x-added":true
}`, string(data))
	delete(useRef.Extensions, "x-use-site")
	data, err = json.Marshal(useRef)
	require.NoError(t, err)
	require.NotContains(t, string(data), "x-use-site")

	value := useRef.Value
	for range 3 {
		require.NoError(t, loader.ResolveRefsIn(doc, nil))
		require.Same(t, value, useRef.Value)
		require.Len(t, useRef.Value.AllOf, 1)
	}
}

func TestSchemaRefSiblingUsesOrdinarySchemaRoundTripRules(t *testing.T) {
	var ref openapi3.SchemaRef
	require.NoError(t, json.Unmarshal([]byte(`{
  "$ref":"#/components/schemas/Base",
  "pattern":"",
  "minLength":0,
  "uniqueItems":false,
  "required":[],
  "properties":{},
  "default":null,
  "maximum":0,
  "x-use-site":true
}`), &ref))

	data, err := json.Marshal(ref)
	require.NoError(t, err)
	// Fields omitted by an ordinary Schema marshaler are also omitted here;
	// pointer-backed zero values and extensions remain present.
	require.JSONEq(t, `{
  "$ref":"#/components/schemas/Base",
  "maximum":0,
  "x-use-site":true
}`, string(data))
}

func TestSchemaRefSiblingPreservesBothDynamicRefs(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.1.0
info:
  title: dynamic ref conjunction
  version: "1"
paths: {}
components:
  schemas:
    Base:
      $dynamicRef: '#base'
    Use:
      $ref: '#/components/schemas/Base'
      $dynamicRef: '#use'
`))
	require.NoError(t, err)

	use := doc.Components.Schemas["Use"].Value
	require.Equal(t, "#use", use.DynamicRef)
	require.Len(t, use.AllOf, 1)
	require.Equal(t, "#base", use.AllOf[0].Value.DynamicRef)
}

func TestSchemaRefSiblingOnSelfReference(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.1.0
info:
  title: recursive ref sibling
  version: "1"
paths: {}
components:
  schemas:
    Node:
      type: object
      required: [name]
      properties:
        name:
          type: string
        parent:
          $ref: '#/components/schemas/Node'
          minProperties: 2
`))
	require.NoError(t, err)

	node := doc.Components.Schemas["Node"].Value
	parent := node.Properties["parent"]
	require.Equal(t, uint64(2), parent.Value.MinProps)
	require.Len(t, parent.Value.AllOf, 1)
	require.Same(t, node, parent.Value.AllOf[0].Value)
	require.Error(t, parent.Value.VisitJSON(map[string]any{"name": "child"}, openapi3.EnableJSONSchema2020()))
	require.NoError(t, parent.Value.VisitJSON(map[string]any{"name": "child", "extra": true}, openapi3.EnableJSONSchema2020()))

	invalidRecursiveValue := map[string]any{
		"name": "child",
		"parent": map[string]any{
			"first":  true,
			"second": true,
		},
	}
	validRecursiveValue := map[string]any{
		"name": "child",
		"parent": map[string]any{
			"name":  "parent",
			"extra": true,
		},
	}
	for _, opts := range [][]openapi3.SchemaValidationOption{
		nil,
		{openapi3.EnableJSONSchema2020()},
	} {
		require.Error(t, node.VisitJSON(invalidRecursiveValue, opts...), "the recursive target must still require name")
		require.NoError(t, node.VisitJSON(validRecursiveValue, opts...))
	}
	cyclicValue := map[string]any{"name": "node"}
	cyclicValue["parent"] = cyclicValue
	require.NoError(t, node.VisitJSON(cyclicValue), "a cycle in both the schema and Go value must terminate")

	data, err := json.Marshal(parent)
	require.NoError(t, err)
	require.JSONEq(t, `{
  "$ref":"#/components/schemas/Node",
  "minProperties":2
}`, string(data))
}

func TestSchemaRefSiblingOnSelfReferenceOutsideComponents(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.1.0
info:
  title: path-local recursive ref sibling
  version: "1"
paths:
  /x:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/paths/~1x/post/requestBody/content/application~1json/schema'
              minProperties: 1
      responses:
        "200":
          description: ok
`))
	require.NoError(t, err)

	schemaRef := doc.Paths.Value("/x").Post.RequestBody.Value.Content["application/json"].Schema
	require.NotNil(t, schemaRef.Value)
	require.Equal(t, uint64(1), schemaRef.Value.MinProps)
	require.Error(t, schemaRef.Value.VisitJSON(map[string]any{}))
	require.NoError(t, schemaRef.Value.VisitJSON(map[string]any{"value": true}))

	data, err := json.Marshal(schemaRef)
	require.NoError(t, err)
	require.JSONEq(t, `{
  "$ref":"#/paths/~1x/post/requestBody/content/application~1json/schema",
  "minProperties":1
}`, string(data))
}

func TestSchemaRefSiblingsOnLongerCycle(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.1.0
info:
  title: ref-only cycle with siblings
  version: "1"
paths: {}
components:
  schemas:
    A:
      $ref: '#/components/schemas/B'
      minLength: 2
    B:
      $ref: '#/components/schemas/A'
      maxLength: 4
`))
	require.NoError(t, err)

	a := doc.Components.Schemas["A"].Value
	b := doc.Components.Schemas["B"].Value
	require.Equal(t, uint64(2), a.MinLength)
	require.Equal(t, uint64(4), *b.MaxLength)
	validationErr := a.VisitJSON("x", openapi3.EnableJSONSchema2020())
	require.Error(t, validationErr)
	require.NotPanics(t, func() { _ = validationErr.Error() }, "cyclic effective schemas must remain serializable in errors")
	require.Error(t, a.VisitJSON("12345", openapi3.EnableJSONSchema2020()))
	require.Error(t, b.VisitJSON("x", openapi3.EnableJSONSchema2020()))
	require.Error(t, b.VisitJSON("12345", openapi3.EnableJSONSchema2020()))
	require.NoError(t, a.VisitJSON("1234", openapi3.EnableJSONSchema2020()))
	require.NoError(t, b.VisitJSON("1234", openapi3.EnableJSONSchema2020()))

	aValue, bValue := a, b
	for range 3 {
		require.NoError(t, loader.ResolveRefsIn(doc, nil))
		require.Same(t, aValue, doc.Components.Schemas["A"].Value)
		require.Same(t, bValue, doc.Components.Schemas["B"].Value)
	}
	for name, expected := range map[string]string{
		"A": `{"$ref":"#/components/schemas/B","minLength":2}`,
		"B": `{"$ref":"#/components/schemas/A","maxLength":4}`,
	} {
		data, err := json.Marshal(doc.Components.Schemas[name])
		require.NoError(t, err)
		require.JSONEq(t, expected, string(data))
	}
}

func TestSchemaRefSiblingAnchorsPartiallyAnnotatedAliasCycle(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.1.0
info:
  title: partially annotated alias cycle
  version: "1"
paths: {}
components:
  schemas:
    A:
      $ref: '#/components/schemas/B'
      minLength: 2
    B:
      $ref: '#/components/schemas/A'
`))
	require.NoError(t, err)

	for _, name := range []string{"A", "B"} {
		value := doc.Components.Schemas[name].Value
		require.NotNil(t, value)
		require.Error(t, value.VisitJSON("x", openapi3.EnableJSONSchema2020()))
		require.NoError(t, value.VisitJSON("xx", openapi3.EnableJSONSchema2020()))
	}
}

func TestSchemaRefSiblingsAreIgnoredOnOpenAPI30AliasCycle(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.0.3
info:
  title: ref-only cycle with ignored siblings
  version: "1"
paths: {}
components:
  schemas:
    A:
      $ref: '#/components/schemas/B'
      minLength: 2
    B:
      $ref: '#/components/schemas/A'
      maxLength: 4
`))
	require.NoError(t, err)

	for _, name := range []string{"A", "B"} {
		ref := doc.Components.Schemas[name]
		require.NotNil(t, ref.Value)
		require.Zero(t, ref.Value.MinLength)
		require.Nil(t, ref.Value.MaxLength)
		require.NoError(t, ref.Value.VisitJSON("x"))
		require.NoError(t, ref.Value.VisitJSON("12345"))
	}

	a, err := json.Marshal(doc.Components.Schemas["A"])
	require.NoError(t, err)
	require.JSONEq(t, `{"$ref":"#/components/schemas/B","minLength":2}`, string(a))
	b, err := json.Marshal(doc.Components.Schemas["B"])
	require.NoError(t, err)
	require.JSONEq(t, `{"$ref":"#/components/schemas/A","maxLength":4}`, string(b))
}

func TestSchemaRefSiblingEffectiveValueRetainsUseSiteOrigin(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.1.0
info:
  title: sibling origin
  version: "1"
paths: {}
components:
  schemas:
    Base:
      type: string
    Use:
      $ref: '#/components/schemas/Base'
      pattern: '(?invalid'
`))
	require.NoError(t, err)

	use := doc.Components.Schemas["Use"]
	require.NotNil(t, use.Value.Origin)
	require.NotNil(t, use.Value.Origin.Key)
	require.Equal(t, 11, use.Value.Origin.Key.Line)
	require.Equal(t, 7, use.Value.Origin.Fields.Get("pattern").Column)

	validationErr := use.Value.VisitJSON("value", openapi3.EnableJSONSchema2020())
	var patternErr *openapi3.SchemaPatternRegexError
	require.True(t, errors.As(validationErr, &patternErr))
	require.Same(t, use.Value.Origin, patternErr.Origin)
}

func TestSchemaRefSiblingStructuralCycleKeepsAliasTarget(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.1.0
info:
  title: structural sibling cycle
  version: "1"
paths: {}
components:
  schemas:
    A:
      $ref: '#/components/schemas/B'
      properties:
        child:
          $ref: '#/components/schemas/A'
    B:
      type: object
      required: [name]
      properties:
        name:
          type: string
`))
	require.NoError(t, err)

	a := doc.Components.Schemas["A"].Value
	require.Error(t, a.VisitJSON(map[string]any{
		"name":  "root",
		"child": map[string]any{"other": true},
	}))
	require.NoError(t, a.VisitJSON(map[string]any{
		"name":  "root",
		"child": map[string]any{"name": "nested"},
	}))
}

func TestSchemaRefSiblingInRawExternalSchemaUsesRootDialect(t *testing.T) {
	dir := t.TempDir()
	rootPath := filepath.Join(dir, "openapi.yaml")
	externalPath := filepath.Join(dir, "schemas.yaml")

	require.NoError(t, os.WriteFile(rootPath, []byte(`
openapi: 3.1.0
info:
  title: raw external schema sibling
  version: "1"
paths: {}
components:
  schemas:
    Use:
      $ref: 'schemas.yaml#/$defs/Alias'
`), 0o600))
	require.NoError(t, os.WriteFile(externalPath, []byte(`
$defs:
  Base:
    type: string
  Alias:
    $ref: '#/$defs/Base'
    minLength: 3
`), 0o600))

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile(rootPath)
	require.NoError(t, err)

	use := doc.Components.Schemas["Use"].Value
	require.Error(t, use.VisitJSON("xx", openapi3.EnableJSONSchema2020()))
	require.NoError(t, use.VisitJSON("xxx", openapi3.EnableJSONSchema2020()))
}

func TestSchemaRefSiblingDoesNotMutateSharedTarget(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.1.0
info:
  title: independent ref siblings
  version: "1"
paths: {}
components:
  schemas:
    Base:
      type: string
      maxLength: 5
    Short:
      $ref: '#/components/schemas/Base'
      maxLength: 3
    Long:
      $ref: '#/components/schemas/Base'
      minLength: 4
`))
	require.NoError(t, err)

	base := doc.Components.Schemas["Base"].Value
	short := doc.Components.Schemas["Short"].Value
	long := doc.Components.Schemas["Long"].Value
	require.Equal(t, uint64(5), *base.MaxLength)
	require.Equal(t, uint64(3), *short.MaxLength)
	require.Equal(t, uint64(4), long.MinLength)
	require.Same(t, base, short.AllOf[0].Value)
	require.Same(t, base, long.AllOf[0].Value)
	require.Error(t, short.VisitJSON("xxxx", openapi3.EnableJSONSchema2020()))
	require.Error(t, long.VisitJSON("xxx", openapi3.EnableJSONSchema2020()))
	require.NoError(t, long.VisitJSON("xxxx", openapi3.EnableJSONSchema2020()))
}

func TestSchemaRefSiblingSerializationIsIndependentOfOASVersion(t *testing.T) {
	const spec = `
openapi: %s
info:
  title: versioned sibling semantics
  version: "1"
paths: {}
components:
  schemas:
    Base:
      type: string
    Use:
      $ref: '#/components/schemas/Base'
      minLength: 3
`

	for _, version := range []string{"3.0.3", "3.1.0"} {
		t.Run(version, func(t *testing.T) {
			loader := openapi3.NewLoader()
			doc, err := loader.LoadFromData([]byte(fmt.Sprintf(spec, version)))
			require.NoError(t, err)
			use := doc.Components.Schemas["Use"]
			data, err := json.Marshal(use)
			require.NoError(t, err)
			require.JSONEq(t, `{
  "$ref":"#/components/schemas/Base",
  "minLength":3
}`, string(data))
			if version == "3.0.3" {
				require.Zero(t, use.Value.MinLength)
			} else {
				require.Equal(t, uint64(3), use.Value.MinLength)
			}
		})
	}
}

func TestMalformedSchemaRefSiblingUsesRootDialect(t *testing.T) {
	const spec = `
openapi: %s
info:
  title: malformed sibling
  version: "1"
paths: {}
components:
  schemas:
    Base:
      type: string
    Use:
      $ref: '#/components/schemas/Base'
      allOf:
        type: integer
`

	loader30 := openapi3.NewLoader()
	doc, err := loader30.LoadFromData([]byte(fmt.Sprintf(spec, "3.0.3")))
	require.NoError(t, err, "OpenAPI 3.0 ignores malformed fields next to $ref")
	require.Equal(t, "string", doc.Components.Schemas["Use"].Value.Type.Slice()[0])

	loader31 := openapi3.NewLoader()
	_, err = loader31.LoadFromData([]byte(fmt.Sprintf(spec, "3.1.0")))
	require.Error(t, err, "OpenAPI 3.1 treats the adjacent fields as a Schema Object")
}

func TestResolveSchemaRefsIn31Fields(t *testing.T) {
	loader := openapi3.NewLoader()
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
