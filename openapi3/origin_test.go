package openapi3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrigin_Info(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = context.Background()

	doc, err := loader.LoadFromFile("testdata/origin/simple.yaml")
	require.NoError(t, err)

	require.NotNil(t, doc.Info.Origin)
	require.Equal(t,
		&Location{
			File:   "testdata/origin/simple.yaml",
			Line:   2,
			Column: 1,
			Name:   "info",
		},
		doc.Info.Origin.Key)

	require.Equal(t,
		Location{
			File:   "testdata/origin/simple.yaml",
			Line:   3,
			Column: 3,
			Name:   "title",
		},
		doc.Info.Origin.Fields["title"])

	require.Equal(t,
		Location{
			File:   "testdata/origin/simple.yaml",
			Line:   4,
			Column: 3,
			Name:   "version",
		},
		doc.Info.Origin.Fields["version"])
}

func TestOrigin_Paths(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = context.Background()

	doc, err := loader.LoadFromFile("testdata/origin/simple.yaml")
	require.NoError(t, err)

	require.NotNil(t, doc.Paths.Origin)
	require.Equal(t,
		&Location{
			File:   "testdata/origin/simple.yaml",
			Line:   5,
			Column: 1,
			Name:   "paths",
		},
		doc.Paths.Origin.Key)

	base := doc.Paths.Find("/partner-api/test/another-method")

	require.NotNil(t, base.Origin)
	require.Equal(t,
		&Location{
			File:   "testdata/origin/simple.yaml",
			Line:   13,
			Column: 3,
			Name:   "/partner-api/test/another-method",
		},
		base.Origin.Key)

	require.NotNil(t, base.Get.Origin)
	require.Equal(t,
		&Location{
			File:   "testdata/origin/simple.yaml",
			Line:   14,
			Column: 5,
			Name:   "get",
		},
		base.Get.Origin.Key)
}

func TestOrigin_RequestBody(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = context.Background()

	doc, err := loader.LoadFromFile("testdata/origin/request_body.yaml")
	require.NoError(t, err)

	base := doc.Paths.Find("/subscribe").Post.RequestBody.Value
	require.NotNil(t, base.Origin)
	require.Equal(t,
		&Location{
			File:   "testdata/origin/request_body.yaml",
			Line:   8,
			Column: 7,
			Name:   "requestBody",
		},
		base.Origin.Key)

	require.NotNil(t, base.Content["application/json"].Origin)
	require.Equal(t,
		&Location{
			File:   "testdata/origin/request_body.yaml",
			Line:   10,
			Column: 11,
			Name:   "application/json",
		},
		base.Content["application/json"].Origin.Key)
}

func TestOrigin_Responses(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = context.Background()

	doc, err := loader.LoadFromFile("testdata/origin/simple.yaml")
	require.NoError(t, err)

	base := doc.Paths.Find("/partner-api/test/another-method").Get.Responses
	require.NotNil(t, base.Origin)
	require.Equal(t,
		&Location{
			File:   "testdata/origin/simple.yaml",
			Line:   17,
			Column: 7,
			Name:   "responses",
		},
		base.Origin.Key)

	require.NotNil(t, base.Origin)
	require.Nil(t, base.Value("200").Origin)
	require.Equal(t,
		&Location{
			File:   "testdata/origin/simple.yaml",
			Line:   18,
			Column: 9,
			Name:   "200",
		},
		base.Value("200").Value.Origin.Key)

	require.Equal(t,
		Location{
			File:   "testdata/origin/simple.yaml",
			Line:   19,
			Column: 11,
			Name:   "description",
		},
		base.Value("200").Value.Origin.Fields["description"])
}

func TestOrigin_Parameters(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = context.Background()

	doc, err := loader.LoadFromFile("testdata/origin/parameters.yaml")
	require.NoError(t, err)

	base := doc.Paths.Find("/api/test").Get.Parameters[0].Value
	require.NotNil(t, base)
	require.Equal(t,
		&Location{
			File:   "testdata/origin/parameters.yaml",
			Line:   9,
			Column: 11,
			Name:   "name",
		},
		base.Origin.Key)

	require.Equal(t,
		Location{
			File:   "testdata/origin/parameters.yaml",
			Line:   10,
			Column: 11,
			Name:   "in",
		},
		base.Origin.Fields["in"])

	require.Equal(t,
		Location{
			File:   "testdata/origin/parameters.yaml",
			Line:   9,
			Column: 11,
			Name:   "name",
		},
		base.Origin.Fields["name"])
}

func TestOrigin_SchemaInAdditionalProperties(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = context.Background()

	doc, err := loader.LoadFromFile("testdata/origin/additional_properties.yaml")
	require.NoError(t, err)

	base := doc.Paths.Find("/partner-api/test/some-method").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value.AdditionalProperties
	require.NotNil(t, base)

	require.NotNil(t, base.Schema.Value.Origin)
	require.Equal(t,
		&Location{
			File:   "testdata/origin/additional_properties.yaml",
			Line:   14,
			Column: 17,
			Name:   "additionalProperties",
		},
		base.Schema.Value.Origin.Key)

	require.Equal(t,
		Location{
			File:   "testdata/origin/additional_properties.yaml",
			Line:   15,
			Column: 19,
			Name:   "type",
		},
		base.Schema.Value.Origin.Fields["type"])
}

func TestOrigin_ExternalDocs(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = context.Background()

	doc, err := loader.LoadFromFile("testdata/origin/external_docs.yaml")
	require.NoError(t, err)

	base := doc.ExternalDocs
	require.NotNil(t, base.Origin)

	require.Equal(t,
		&Location{
			File:   "testdata/origin/external_docs.yaml",
			Line:   13,
			Column: 1,
			Name:   "externalDocs",
		},
		base.Origin.Key)

	require.Equal(t,
		Location{
			File:   "testdata/origin/external_docs.yaml",
			Line:   14,
			Column: 3,
			Name:   "description",
		},
		base.Origin.Fields["description"])

	require.Equal(t,
		Location{
			File:   "testdata/origin/external_docs.yaml",
			Line:   15,
			Column: 3,
			Name:   "url",
		},
		base.Origin.Fields["url"])
}

func TestOrigin_Security(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = context.Background()

	doc, err := loader.LoadFromFile("testdata/origin/security.yaml")
	require.NoError(t, err)

	base := doc.Components.SecuritySchemes["petstore_auth"].Value
	require.NotNil(t, base)

	require.Equal(t,
		&Location{
			File:   "testdata/origin/security.yaml",
			Line:   29,
			Column: 5,
			Name:   "petstore_auth",
		},
		base.Origin.Key)

	require.Equal(t,
		Location{
			File:   "testdata/origin/security.yaml",
			Line:   30,
			Column: 7,
			Name:   "type",
		},
		base.Origin.Fields["type"])

	require.Equal(t,
		&Location{
			File:   "testdata/origin/security.yaml",
			Line:   31,
			Column: 7,
			Name:   "flows",
		},
		base.Flows.Origin.Key)

	require.Equal(t,
		&Location{
			File:   "testdata/origin/security.yaml",
			Line:   32,
			Column: 9,
			Name:   "implicit",
		},
		base.Flows.Implicit.Origin.Key)

	require.Equal(t,
		Location{
			File:   "testdata/origin/security.yaml",
			Line:   33,
			Column: 11,
			Name:   "authorizationUrl",
		},
		base.Flows.Implicit.Origin.Fields["authorizationUrl"])
}

func TestOrigin_Example(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = context.Background()

	doc, err := loader.LoadFromFile("testdata/origin/example.yaml")
	require.NoError(t, err)

	base := doc.Paths.Find("/subscribe").Post.RequestBody.Value.Content["application/json"].Examples["bar"].Value
	require.NotNil(t, base.Origin)
	require.Equal(t,
		&Location{
			File:   "testdata/origin/example.yaml",
			Line:   14,
			Column: 15,
			Name:   "bar",
		},
		base.Origin.Key)

	require.Equal(t,
		Location{
			File:   "testdata/origin/example.yaml",
			Line:   15,
			Column: 17,
			Name:   "summary",
		},
		base.Origin.Fields["summary"])

	// Example.Value is an any-typed field, so __origin__ is stripped from it during unmarshaling.
	require.NotContains(t,
		base.Value,
		originKey)
}

func TestOrigin_XML(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = context.Background()

	doc, err := loader.LoadFromFile("testdata/origin/xml.yaml")
	require.NoError(t, err)

	base := doc.Paths.Find("/subscribe").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value.XML
	require.NotNil(t, base.Origin)
	require.Equal(t,
		&Location{
			File:   "testdata/origin/xml.yaml",
			Line:   21,
			Column: 19,
			Name:   "xml",
		},
		base.Origin.Key)

	require.Equal(t,
		Location{
			File:   "testdata/origin/xml.yaml",
			Line:   22,
			Column: 21,
			Name:   "namespace",
		},
		base.Origin.Fields["namespace"])

	require.Equal(t,
		Location{
			File:   "testdata/origin/xml.yaml",
			Line:   23,
			Column: 21,
			Name:   "prefix",
		},
		base.Origin.Fields["prefix"])
}

// TestOrigin_AnyFieldsStripped verifies that __origin__ is absent from all
// any-typed fields (Schema.Enum, Schema.Default, Schema.Example,
// Parameter.Example, MediaType.Example, Link.RequestBody) after loading.
// These fields have no dedicated UnmarshalJSON; extractOrigins strips
// __origin__ before JSON marshaling so it never reaches these values.
func TestOrigin_AnyFieldsStripped(t *testing.T) {
	loader := NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromFile("testdata/origin/any_fields.yaml")
	require.NoError(t, err)

	op := doc.Paths.Find("/items").Get
	resp := op.Responses.Value("200").Value

	// Parameter.Example
	paramEx := op.Parameters[0].Value.Example.(map[string]any)
	require.NotContains(t, paramEx, originKey, "Parameter.Example must not contain __origin__")

	// MediaType.Example
	mediaEx := resp.Content["application/json"].Example.(map[string]any)
	require.NotContains(t, mediaEx, originKey, "MediaType.Example must not contain __origin__")

	schema := resp.Content["application/json"].Schema.Value

	// Schema.Default
	schemaDefault := schema.Default.(map[string]any)
	require.NotContains(t, schemaDefault, originKey, "Schema.Default must not contain __origin__")

	// Schema.Example
	schemaEx := schema.Example.(map[string]any)
	require.NotContains(t, schemaEx, originKey, "Schema.Example must not contain __origin__")

	// Schema.Enum items
	for i, v := range schema.Enum {
		m, ok := v.(map[string]any)
		require.True(t, ok, "Schema.Enum[%d] must be a map", i)
		require.NotContains(t, m, originKey, "Schema.Enum[%d] must not contain __origin__", i)
	}

	// Link.RequestBody
	linkRB := resp.Links["self"].Value.RequestBody.(map[string]any)
	require.NotContains(t, linkRB, originKey, "Link.RequestBody must not contain __origin__")
}

func TestOrigin_ExampleWithArrayValue(t *testing.T) {
	loader := NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromFile("testdata/origin/example_with_array.yaml")
	require.NoError(t, err)

	example := doc.Paths.Find("/subscribe").Post.RequestBody.Value.Content["application/json"].Examples["bar"]
	require.NotNil(t, example.Value)

	// The example value contains a list of objects; __origin__ must be stripped from each.
	value := example.Value.Value.(map[string]any)
	items := value["items"].([]any)
	for _, item := range items {
		itemMap := item.(map[string]any)
		require.NotContains(t, itemMap, "__origin__")
	}
}

// TestOrigin_OriginExistsInProperties verifies that loading fails when a specification
// contains a property named "__origin__", highlighting a limitation in the current implementation.
func TestOrigin_OriginExistsInProperties(t *testing.T) {
	var data = `
paths:
  /foo:
    get:
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Foo"
components:
  schemas:
    Foo:
      type: object
      properties:
        __origin__:
          type: string
`

	loader := NewLoader()
	loader.IncludeOrigin = true

	_, err := loader.LoadFromData([]byte(data))
	require.Error(t, err)
	require.Equal(t, `failed to unmarshal data: json error: invalid character 'p' looking for beginning of value, yaml error: error converting YAML to JSON: yaml: unmarshal errors:
  line 0: mapping key "__origin__" already defined at line 17`, err.Error())
}

// TestOrigin_ExtensionValuesStripped verifies that __origin__ metadata injected
// by the YAML decoder is not present in any-typed extension values.
// Regression test: extension values that are YAML objects received __origin__
// from the yaml3 decoder but it was never stripped, causing spurious diffs
// between specs loaded from different file paths.
func TestOrigin_ExtensionValuesStripped(t *testing.T) {
	loader := NewLoader()
	loader.IncludeOrigin = true

	doc, err := loader.LoadFromFile("testdata/origin/extensions.yaml")
	require.NoError(t, err)

	val, ok := doc.Extensions["x-object-extension"]
	require.True(t, ok, "x-object-extension must be present")

	m, ok := val.(map[string]any)
	require.True(t, ok, "x-object-extension value must be a map")

	require.NotContains(t, m, originKey, "__origin__ must be stripped from extension object values")

	// Also verify stripping works for a nested type (Info), covering the 20
	// per-type UnmarshalJSON call sites with a single representative case.
	infoVal, ok := doc.Info.Extensions["x-info-extension"]
	require.True(t, ok, "x-info-extension must be present")

	infoMap, ok := infoVal.(map[string]any)
	require.True(t, ok, "x-info-extension value must be a map")

	require.NotContains(t, infoMap, originKey, "__origin__ must be stripped from nested extension object values")
}

func TestOrigin_WithExternalRef(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true

	loader.Context = context.Background()

	doc, err := loader.LoadFromFile("testdata/origin/external.yaml")
	require.NoError(t, err)

	base := doc.Paths.Find("/subscribe").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value
	require.NotNil(t, base.XML.Origin)
	require.Equal(t,
		&Location{
			File:   "testdata/origin/external-schema.yaml",
			Line:   2,
			Column: 1,
			Name:   "xml",
		},
		base.XML.Origin.Key)

	require.Equal(t,
		Location{
			File:   "testdata/origin/external-schema.yaml",
			Line:   3,
			Column: 3,
			Name:   "namespace",
		},
		base.XML.Origin.Fields["namespace"])

	require.Equal(t,
		Location{
			File:   "testdata/origin/external-schema.yaml",
			Line:   4,
			Column: 3,
			Name:   "prefix",
		},
		base.XML.Origin.Fields["prefix"])
}

// TestOrigin_MaplikeNoOriginKey verifies that __origin__ does not appear as a
// map key in Responses, Paths, or Callback maplike types after loading.
// The if k == originKey blocks in their UnmarshalJSON were removed; this
// confirms extractOrigins strips __origin__ before it reaches those iterators.
func TestOrigin_MaplikeNoOriginKey(t *testing.T) {
	loader := NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromFile("testdata/origin/simple.yaml")
	require.NoError(t, err)

	// Paths map must not contain __origin__ as a path key
	require.Nil(t, doc.Paths.Find(originKey), "Paths must not contain __origin__ as a key")

	// Responses map must not contain __origin__ as a status code key
	op := doc.Paths.Find("/partner-api/test/some-method").Get
	require.Nil(t, op.Responses.Value(originKey), "Responses must not contain __origin__ as a key")
}

func TestOrigin_NoSpuriousOriginsInComponents(t *testing.T) {
	loader := NewLoader()
	loader.IncludeOrigin = true

	doc, err := loader.LoadFromFile("testdata/origin/components.yaml")

	require.Nil(t, doc.Components.Schemas[originKey])
	require.Nil(t, doc.Components.Parameters[originKey])
	require.Nil(t, doc.Components.Headers[originKey])
	require.Nil(t, doc.Components.RequestBodies[originKey])
	require.Nil(t, doc.Components.Responses[originKey])
	require.Nil(t, doc.Components.SecuritySchemes[originKey])
	require.Nil(t, doc.Components.Examples[originKey])
	require.Nil(t, doc.Components.Links[originKey])
	require.Nil(t, doc.Components.Callbacks[originKey])

	require.NoError(t, err)
}

// TestOrigin_RequiredSequence verifies that Origin.Sequences records the
// file/line/column of each item in a required: [...] list.
// These locations are used by NewSourceFromSequenceItem to pinpoint
// breaking changes to individual required field names.
func TestOrigin_RequiredSequence(t *testing.T) {
	loader := NewLoader()
	loader.IncludeOrigin = true

	doc, err := loader.LoadFromFile("testdata/origin/required_sequence.yaml")
	require.NoError(t, err)

	schema := doc.Paths.Find("/items").Post.RequestBody.Value.Content["application/json"].Schema.Value
	require.NotNil(t, schema.Origin)

	// "required" must appear in Fields (it's a sequence-valued field)
	require.Contains(t, schema.Origin.Fields, "required")

	// Sequences must record per-item locations for "required"
	seqLocs, ok := schema.Origin.Sequences["required"]
	require.True(t, ok, "Origin.Sequences must contain 'required'")
	require.Len(t, seqLocs, 2)

	require.Equal(t, Location{
		File:   "testdata/origin/required_sequence.yaml",
		Line:   14,
		Column: 19,
		Name:   "name",
	}, seqLocs[0])

	require.Equal(t, Location{
		File:   "testdata/origin/required_sequence.yaml",
		Line:   15,
		Column: 19,
		Name:   "age",
	}, seqLocs[1])
}

// TestOrigin_YAMLAlias verifies that a schema referenced via YAML alias loads
// without error and carries origin metadata from the anchor definition.
// Multiple aliases of the same anchor must not produce duplicate __origin__ keys.
func TestOrigin_YAMLAlias(t *testing.T) {
	loader := NewLoader()
	loader.IncludeOrigin = true

	doc, err := loader.LoadFromFile("testdata/origin/alias.yaml")
	require.NoError(t, err)

	anchor := doc.Components.Schemas["Base"].Value
	alias1 := doc.Components.Schemas["Alias1"].Value
	alias2 := doc.Components.Schemas["Alias2"].Value

	// All three point to the same anchor node, so origin reflects the anchor location.
	anchorLoc := &Location{
		File:   "testdata/origin/alias.yaml",
		Line:   7,
		Column: 5,
		Name:   "Base",
	}
	require.Equal(t, anchorLoc, anchor.Origin.Key)
	require.Equal(t, anchorLoc, alias1.Origin.Key)
	require.Equal(t, anchorLoc, alias2.Origin.Key)
}

// TestOrigin_Headers verifies that response header origin is tracked correctly.
func TestOrigin_Headers(t *testing.T) {
	loader := NewLoader()
	loader.IncludeOrigin = true

	doc, err := loader.LoadFromFile("testdata/origin/headers.yaml")
	require.NoError(t, err)

	headers := doc.Paths.Find("/items").Get.Responses.Value("200").Value.Headers

	require.Equal(t,
		&Location{
			File:   "testdata/origin/headers.yaml",
			Line:   12,
			Column: 13,
			Name:   "X-Rate-Limit",
		},
		headers["X-Rate-Limit"].Value.Origin.Key)

	require.Equal(t,
		Location{
			File:   "testdata/origin/headers.yaml",
			Line:   13,
			Column: 15,
			Name:   "description",
		},
		headers["X-Rate-Limit"].Value.Origin.Fields["description"])

	require.Equal(t,
		&Location{
			File:   "testdata/origin/headers.yaml",
			Line:   16,
			Column: 13,
			Name:   "X-Request-Id",
		},
		headers["X-Request-Id"].Value.Origin.Key)
}

// TestOrigin_IntegerStatusCode verifies that response origin is tracked when
// HTTP status codes are written as bare integers (200:) rather than quoted
// strings ("200":). Bare integers produce map[interface{}]interface{} in the
// YAML decoder, which required a dedicated fix in extractOrigins.
func TestOrigin_IntegerStatusCode(t *testing.T) {
	loader := NewLoader()
	loader.IncludeOrigin = true

	doc, err := loader.LoadFromFile("testdata/origin/parameters.yaml")
	require.NoError(t, err)

	resp200 := doc.Paths.Find("/api/test").Get.Responses.Value("200").Value
	require.NotNil(t, resp200.Origin)
	require.Equal(t,
		&Location{
			File:   "testdata/origin/parameters.yaml",
			Line:   14,
			Column: 9,
			Name:   "200",
		},
		resp200.Origin.Key)

	resp201 := doc.Paths.Find("/api/test").Post.Responses.Value("201").Value
	require.NotNil(t, resp201.Origin)
	require.Equal(t,
		&Location{
			File:   "testdata/origin/parameters.yaml",
			Line:   18,
			Column: 9,
			Name:   "201",
		},
		resp201.Origin.Key)
}

// TestOrigin_Disabled verifies that all Origin fields are nil when
// IncludeOrigin is false (the default), ensuring no overhead in the common case.
func TestOrigin_Disabled(t *testing.T) {
	loader := NewLoader()
	// IncludeOrigin defaults to false

	doc, err := loader.LoadFromFile("testdata/origin/required_sequence.yaml")
	require.NoError(t, err)

	schema := doc.Paths.Find("/items").Post.RequestBody.Value.Content["application/json"].Schema.Value
	require.Nil(t, schema.Origin)
	require.Nil(t, doc.Info.Origin)
	require.Nil(t, doc.Paths.Origin)
}
