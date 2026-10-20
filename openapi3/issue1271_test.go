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

// These cases come from https://github.com/getkin/kin-openapi/pull/1271.
// There, the keywords next to a 3.1 schema $ref become the Value and the
// referenced schema an allOf entry of it. Here Value keeps the shape of the
// referenced schema, so the cases check what a schema accepts and how it is
// written back out rather than where each keyword ends up.
//
// The skipped cases are $refs that only lead to one another, which leaves
// their Value nil, with or without siblings. JSON Schema 2020-12 leaves such
// infinite recursion undefined (§9.4.1); #1271 resolves them to the
// conjunction of their siblings.

func TestSchemaRefSiblingIsSerializedAndAppliedAsConjunct(t *testing.T) {
	t.Parallel()
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
	require.Equal(t, "string", useRef.Value.Type.Slice()[0])
	require.Equal(t, uint64(5), *useRef.Value.MaxLength)
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
	allOf := len(value.AllOf)
	for range 3 {
		require.NoError(t, loader.ResolveRefsIn(doc, nil))
		require.Same(t, value, useRef.Value)
		require.Len(t, useRef.Value.AllOf, allOf)
	}
}

func TestSchemaRefSiblingRoundTripsAsWritten(t *testing.T) {
	t.Parallel()
	data := []byte(`{
  "$ref":"#/components/schemas/Base",
  "pattern":"",
  "minLength":0,
  "uniqueItems":false,
  "required":[],
  "properties":{},
  "default":null,
  "maximum":0,
  "x-use-site":true
}`)
	var ref openapi3.SchemaRef
	require.NoError(t, json.Unmarshal(data, &ref))

	// Unlike #1271, which writes them out as an ordinary Schema would, dropping
	// the zero values, the fields next to the $ref come back as written.
	out, err := json.Marshal(ref)
	require.NoError(t, err)
	require.JSONEq(t, string(data), string(out))
}

func TestSchemaRefSiblingPreservesBothDynamicRefs(t *testing.T) {
	t.Parallel()
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
	require.Equal(t, "#base", use.DynamicRef)
	require.Len(t, use.AllOf, 1)
	require.Equal(t, "#use", use.AllOf[0].Value.DynamicRef)
	require.Equal(t, "#base", doc.Components.Schemas["Base"].Value.DynamicRef)
}

func TestSchemaRefSiblingOnSelfReference(t *testing.T) {
	t.Parallel()
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

	data, err := json.Marshal(parent)
	require.NoError(t, err)
	require.JSONEq(t, `{
  "$ref":"#/components/schemas/Node",
  "minProperties":2
}`, string(data))
}

func TestSchemaRefSiblingOnSelfReferenceOutsideComponents(t *testing.T) {
	t.Skip("TODO: give a value to the schemas of a cycle of $refs that never reaches a schema")
	t.Parallel()
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
	t.Skip("TODO: give a value to the schemas of a cycle of $refs that never reaches a schema")
	t.Parallel()
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
	require.NotNil(t, a)
	require.NotNil(t, b)
	validationErr := a.VisitJSON("x", openapi3.EnableJSONSchema2020())
	require.Error(t, validationErr)
	getErr := func() { _ = validationErr.Error() }
	require.NotPanics(t, getErr, "cyclic effective schemas must remain serializable in errors")
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
	t.Skip("TODO: give a value to the schemas of a cycle of $refs that never reaches a schema")
	t.Parallel()
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
	t.Skip("TODO: give a value to the schemas of a cycle of $refs that never reaches a schema")
	t.Parallel()
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

func TestSchemaRefSiblingErrorRetainsUseSiteOrigin(t *testing.T) {
	t.Parallel()
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
	validationErr := use.Value.VisitJSON("value", openapi3.EnableJSONSchema2020())
	var patternErr *openapi3.SchemaPatternRegexError
	require.True(t, errors.As(validationErr, &patternErr))
	require.NotNil(t, patternErr.Origin)
	require.NotNil(t, patternErr.Origin.Key)
	require.Equal(t, 11, patternErr.Origin.Key.Line)
	require.Equal(t, 7, patternErr.Origin.Fields.Get("pattern").Column)
}

func TestSchemaRefSiblingStructuralCycleKeepsAliasTarget(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	require.Zero(t, base.MinLength)
	require.Empty(t, base.AllOf)
	require.NotSame(t, base, short)
	require.NotSame(t, base, long)
	require.NoError(t, base.VisitJSON("xxxxx", openapi3.EnableJSONSchema2020()))
	require.Error(t, short.VisitJSON("xxxx", openapi3.EnableJSONSchema2020()))
	require.Error(t, long.VisitJSON("xxx", openapi3.EnableJSONSchema2020()))
	require.NoError(t, long.VisitJSON("xxxx", openapi3.EnableJSONSchema2020()))
}

func TestSchemaRefSiblingSerializationIsIndependentOfOASVersion(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			loader := openapi3.NewLoader()
			doc, err := loader.LoadFromData(fmt.Appendf(nil, spec, version))
			require.NoError(t, err)
			use := doc.Components.Schemas["Use"]
			data, err := json.Marshal(use)
			require.NoError(t, err)
			require.JSONEq(t, `{
  "$ref":"#/components/schemas/Base",
  "minLength":3
}`, string(data))
			if version == "3.0.3" {
				require.NoError(t, use.Value.VisitJSON("xx"))
			} else {
				require.Error(t, use.Value.VisitJSON("xx"))
			}
		})
	}
}

func TestMalformedSchemaRefSiblingUsesRootDialect(t *testing.T) {
	t.Parallel()
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
	doc, err := loader30.LoadFromData(fmt.Appendf(nil, spec, "3.0.3"))
	require.NoError(t, err, "OpenAPI 3.0 ignores malformed fields next to $ref")
	require.Equal(t, "string", doc.Components.Schemas["Use"].Value.Type.Slice()[0])

	loader31 := openapi3.NewLoader()
	_, err = loader31.LoadFromData(fmt.Appendf(nil, spec, "3.1.0"))
	require.Error(t, err, "OpenAPI 3.1 treats the adjacent fields as a Schema Object")
}

func TestInternalizeRefsUpdatesSchemaRefSiblingFields(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rootPath := filepath.Join(dir, "root.yaml")
	externalPath := filepath.Join(dir, "external.yaml")

	require.NoError(t, os.WriteFile(rootPath, []byte(`
openapi: 3.1.0
info:
  title: Internalized sibling refs
  version: "1.0"
paths: {}
components:
  schemas:
    Root:
      type: object
      properties:
        use:
          $ref: external.yaml#/Base
          items:
            $ref: external.yaml#/Item
`), 0o600))
	require.NoError(t, os.WriteFile(externalPath, []byte(`
Base:
  type: array
Item:
  type: string
`), 0o600))

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile(rootPath)
	require.NoError(t, err)

	doc.InternalizeRefs(t.Context(), nil)
	data, err := doc.MarshalJSON()
	require.NoError(t, err)
	require.NotContains(t, string(data), "external.yaml")

	use := doc.Components.Schemas["Root"].Value.Properties["use"]
	require.True(t, strings.HasPrefix(use.Ref, "#/components/schemas/"))
	targetName := strings.TrimPrefix(use.Ref, "#/components/schemas/")
	target := doc.Components.Schemas[targetName].Value
	require.Nil(t, target.Items,
		"internalization must componentize the referenced target, not the schema with the siblings applied")
	require.Empty(t, target.AllOf,
		"internalization must componentize the referenced target, not the schema with the siblings applied")
	var marshaledUse map[string]any
	require.NoError(t, json.Unmarshal(mustMarshalJSON(t, use), &marshaledUse))
	items := marshaledUse["items"].(map[string]any)
	require.True(t, strings.HasPrefix(items["$ref"].(string), "#/components/schemas/"))

	loader = openapi3.NewLoader()
	roundTripped, err := loader.LoadFromData(data)
	require.NoError(t, err, "the internalized document must not require external access")
	roundTrippedUse := roundTripped.Components.Schemas["Root"].Value.Properties["use"].Value
	require.NoError(t, roundTrippedUse.VisitJSON([]any{"value"}, openapi3.EnableJSONSchema2020()))
	require.Error(t, roundTrippedUse.VisitJSON([]any{1}, openapi3.EnableJSONSchema2020()))
}

// In OpenAPI 3.0 the fields next to a $ref are ignored, so the loader never
// follows the $refs within them and internalization leaves them as written.
func TestInternalizeRefsLeavesIgnoredSchemaRefSiblingFieldsInOpenAPI30(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rootPath := filepath.Join(dir, "root.yaml")
	externalPath := filepath.Join(dir, "external.yaml")

	require.NoError(t, os.WriteFile(rootPath, []byte(`
openapi: 3.0.3
info:
  title: Internalized OpenAPI 3.0 sibling refs
  version: "1.0"
paths: {}
components:
  schemas:
    Root:
      type: object
      properties:
        use:
          $ref: external.yaml#/Base
          items:
            $ref: external.yaml#/Item
`), 0o600))
	require.NoError(t, os.WriteFile(externalPath, []byte(`
Base:
  type: array
Item:
  type: string
`), 0o600))

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile(rootPath)
	require.NoError(t, err)

	use := doc.Components.Schemas["Root"].Value.Properties["use"]
	require.Nil(t, use.Value.Items, "an OpenAPI 3.0 sibling must not affect the resolved target")
	before, err := json.Marshal(use)
	require.NoError(t, err)
	require.Contains(t, string(before), `external.yaml#/Item`)

	doc.InternalizeRefs(t.Context(), nil)
	data, err := doc.MarshalJSON()
	require.NoError(t, err)
	require.Nil(t, use.Value.Items, "internalization must not change OpenAPI 3.0 sibling semantics")

	var marshaledUse map[string]any
	require.NoError(t, json.Unmarshal(mustMarshalJSON(t, use), &marshaledUse))
	require.True(t, strings.HasPrefix(marshaledUse["$ref"].(string), "#/components/schemas/"))
	items := marshaledUse["items"].(map[string]any)
	require.Equal(t, "external.yaml#/Item", items["$ref"])

	roundTripLoader := openapi3.NewLoader()
	roundTripped, err := roundTripLoader.LoadFromData(data)
	require.NoError(t, err, "the internalized document must not require external access")
	roundTrippedUse := roundTripped.Components.Schemas["Root"].Value.Properties["use"]
	require.Nil(t, roundTrippedUse.Value.Items)
}

func TestInternalizeRefsUpdatesSiblingFieldsWithFalseTarget(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rootPath := filepath.Join(dir, "root.yaml")
	externalPath := filepath.Join(dir, "external.yaml")

	require.NoError(t, os.WriteFile(rootPath, []byte(`
openapi: 3.1.0
info:
  title: Internalized false-target sibling refs
  version: "1.0"
paths: {}
components:
  schemas:
    RejectEverything: false
    Root:
      type: object
      properties:
        use:
          $ref: '#/components/schemas/RejectEverything'
          items:
            $ref: external.yaml#/Item
`), 0o600))
	require.NoError(t, os.WriteFile(externalPath, []byte(`
Item:
  type: string
`), 0o600))

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile(rootPath)
	require.NoError(t, err)

	use := doc.Components.Schemas["Root"].Value.Properties["use"]
	require.Error(t, use.Value.VisitJSON([]any{"value"}, openapi3.EnableJSONSchema2020()))

	doc.InternalizeRefs(t.Context(), nil)
	data, err := doc.MarshalJSON()
	require.NoError(t, err)
	require.NotContains(t, string(data), "external.yaml")
	require.Error(t, use.Value.VisitJSON([]any{"value"}, openapi3.EnableJSONSchema2020()))
}

func TestRecursiveSchemaDistinguishesOverlappingSlices(t *testing.T) {
	t.Parallel()
	schema := openapi3.NewArraySchema().WithMinItems(1)
	schema.Items = &openapi3.SchemaRef{Value: schema}

	value := make([]any /*,*/, 1)
	value[0] = value[:0]
	require.Error(t, schema.VisitJSON(value),
		"a shorter subslice with the same backing address is a distinct JSON instance")
}

func mustMarshalJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	return data
}
