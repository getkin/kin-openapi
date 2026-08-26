package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestOrigin_T(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = t.Context()

	doc, err := loader.LoadFromFile("testdata/origin/simple.yaml")
	require.NoError(t, err)

	require.NotNil(t, doc.Origin)
	require.NotNil(t, doc.Origin.Key)
	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/simple.yaml",
			Line:   1,
			Column: 1,
			Name:   "openapi",
		},
		doc.Origin.Fields.Get("openapi"))
}

func TestOrigin_Info(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = t.Context()

	doc, err := loader.LoadFromFile("testdata/origin/simple.yaml")
	require.NoError(t, err)

	require.NotNil(t, doc.Info.Origin)
	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/simple.yaml",
			Line:      2,
			Column:    1,
			Name:      "info",
			EndLine:   4,
			EndColumn: 14,
		},
		doc.Info.Origin.Key)

	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/simple.yaml",
			Line:   3,
			Column: 3,
			Name:   "title",
		},
		doc.Info.Origin.Fields.Get("title"))

	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/simple.yaml",
			Line:   4,
			Column: 3,
			Name:   "version",
		},
		doc.Info.Origin.Fields.Get("version"))
}

func TestOrigin_Paths(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = t.Context()

	doc, err := loader.LoadFromFile("testdata/origin/simple.yaml")
	require.NoError(t, err)

	require.NotNil(t, doc.Paths.Origin)
	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/simple.yaml",
			Line:      5,
			Column:    1,
			Name:      "paths",
			EndLine:   19,
			EndColumn: 31,
		},
		doc.Paths.Origin.Key)

	base := doc.Paths.Find("/partner-api/test/another-method")

	require.NotNil(t, base.Origin)
	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/simple.yaml",
			Line:      13,
			Column:    3,
			Name:      "/partner-api/test/another-method",
			EndLine:   19,
			EndColumn: 31,
		},
		base.Origin.Key)

	// The operation's Origin.Key spans the whole endpoint block: it starts at
	// `get:` (line 14) and ends at the operation's last content (line 19). This
	// is what lets a consumer point at the endpoint's location, not just the
	// exact change site.
	require.NotNil(t, base.Get.Origin)
	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/simple.yaml",
			Line:      14,
			Column:    5,
			Name:      "get",
			EndLine:   19,
			EndColumn: 31,
		},
		base.Get.Origin.Key)
}

func TestOrigin_RequestBody(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = t.Context()

	doc, err := loader.LoadFromFile("testdata/origin/request_body.yaml")
	require.NoError(t, err)

	base := doc.Paths.Find("/subscribe").Post.RequestBody.Value
	require.NotNil(t, base.Origin)
	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/request_body.yaml",
			Line:      8,
			Column:    7,
			Name:      "requestBody",
			EndLine:   19,
			EndColumn: 31,
		},
		base.Origin.Key)

	require.NotNil(t, base.Content["application/json"].Origin)
	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/request_body.yaml",
			Line:      10,
			Column:    11,
			Name:      "application/json",
			EndLine:   19,
			EndColumn: 31,
		},
		base.Content["application/json"].Origin.Key)
}

func TestOrigin_Responses(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = t.Context()

	doc, err := loader.LoadFromFile("testdata/origin/simple.yaml")
	require.NoError(t, err)

	base := doc.Paths.Find("/partner-api/test/another-method").Get.Responses
	require.NotNil(t, base.Origin)
	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/simple.yaml",
			Line:      17,
			Column:    7,
			Name:      "responses",
			EndLine:   19,
			EndColumn: 31,
		},
		base.Origin.Key)

	require.NotNil(t, base.Origin)
	// ResponseRef.Origin is populated with the same data as Value.Origin
	require.NotNil(t, base.Value("200").Origin)
	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/simple.yaml",
			Line:      18,
			Column:    9,
			Name:      "200",
			EndLine:   19,
			EndColumn: 31,
		},
		base.Value("200").Origin.Key)
	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/simple.yaml",
			Line:      18,
			Column:    9,
			Name:      "200",
			EndLine:   19,
			EndColumn: 31,
		},
		base.Value("200").Value.Origin.Key)

	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/simple.yaml",
			Line:   19,
			Column: 11,
			Name:   "description",
		},
		base.Value("200").Value.Origin.Fields.Get("description"))
}

func TestOrigin_Parameters(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = t.Context()

	doc, err := loader.LoadFromFile("testdata/origin/parameters.yaml")
	require.NoError(t, err)

	base := doc.Paths.Find("/api/test").Get.Parameters[0].Value
	require.NotNil(t, base)
	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/parameters.yaml",
			Line:      9,
			Column:    11,
			Name:      "name",
			EndLine:   12,
			EndColumn: 26,
		},
		base.Origin.Key)

	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/parameters.yaml",
			Line:   10,
			Column: 11,
			Name:   "in",
		},
		base.Origin.Fields.Get("in"))

	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/parameters.yaml",
			Line:   9,
			Column: 11,
			Name:   "name",
		},
		base.Origin.Fields.Get("name"))
}

func TestOrigin_SchemaInAdditionalProperties(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = t.Context()

	doc, err := loader.LoadFromFile("testdata/origin/additional_properties.yaml")
	require.NoError(t, err)

	base := doc.Paths.Find("/partner-api/test/some-method").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value.AdditionalProperties
	require.NotNil(t, base)

	require.NotNil(t, base.Schema.Value.Origin)
	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/additional_properties.yaml",
			Line:      14,
			Column:    17,
			Name:      "additionalProperties",
			EndLine:   20,
			EndColumn: 35,
		},
		base.Schema.Value.Origin.Key)

	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/additional_properties.yaml",
			Line:   15,
			Column: 19,
			Name:   "type",
		},
		base.Schema.Value.Origin.Fields.Get("type"))
}

func TestOrigin_ExternalDocs(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = t.Context()

	doc, err := loader.LoadFromFile("testdata/origin/external_docs.yaml")
	require.NoError(t, err)

	base := doc.ExternalDocs
	require.NotNil(t, base.Origin)

	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/external_docs.yaml",
			Line:      13,
			Column:    1,
			Name:      "externalDocs",
			EndLine:   15,
			EndColumn: 38,
		},
		base.Origin.Key)

	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/external_docs.yaml",
			Line:   14,
			Column: 3,
			Name:   "description",
		},
		base.Origin.Fields.Get("description"))

	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/external_docs.yaml",
			Line:   15,
			Column: 3,
			Name:   "url",
		},
		base.Origin.Fields.Get("url"))
}

func TestOrigin_Security(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = t.Context()

	doc, err := loader.LoadFromFile("testdata/origin/security.yaml")
	require.NoError(t, err)

	base := doc.Components.SecuritySchemes["petstore_auth"].Value
	require.NotNil(t, base)

	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/security.yaml",
			Line:      29,
			Column:    5,
			Name:      "petstore_auth",
			EndLine:   36,
			EndColumn: 38,
		},
		base.Origin.Key)

	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/security.yaml",
			Line:   30,
			Column: 7,
			Name:   "type",
		},
		base.Origin.Fields.Get("type"))

	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/security.yaml",
			Line:      31,
			Column:    7,
			Name:      "flows",
			EndLine:   36,
			EndColumn: 38,
		},
		base.Flows.Origin.Key)

	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/security.yaml",
			Line:      32,
			Column:    9,
			Name:      "implicit",
			EndLine:   36,
			EndColumn: 38,
		},
		base.Flows.Implicit.Origin.Key)

	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/security.yaml",
			Line:   33,
			Column: 11,
			Name:   "authorizationUrl",
		},
		base.Flows.Implicit.Origin.Fields.Get("authorizationUrl"))

	// scopes is a map[string]string, which decodes without an Origin of its own,
	// so its per-key locations are recorded on the flow's Origin as a named
	// sequence (sorted by key).
	require.Equal(t,
		[]openapi3.Location{
			{File: "testdata/origin/security.yaml", Line: 36, Column: 13, Name: "read:pets"},
			{File: "testdata/origin/security.yaml", Line: 35, Column: 13, Name: "write:pets"},
		},
		base.Flows.Implicit.Origin.Sequences["scopes"])
}

func TestOrigin_Example(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = t.Context()

	doc, err := loader.LoadFromFile("testdata/origin/example.yaml")
	require.NoError(t, err)

	base := doc.Paths.Find("/subscribe").Post.RequestBody.Value.Content["application/json"].Examples["bar"].Value
	require.NotNil(t, base.Origin)
	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/example.yaml",
			Line:      14,
			Column:    15,
			Name:      "bar",
			EndLine:   16,
			EndColumn: 38, // just past the closing `}` of the flow map `{"bar": "baz"}`
		},
		base.Origin.Key)

	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/example.yaml",
			Line:   15,
			Column: 17,
			Name:   "summary",
		},
		base.Origin.Fields.Get("summary"))

}

func TestOrigin_XML(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = t.Context()

	doc, err := loader.LoadFromFile("testdata/origin/xml.yaml")
	require.NoError(t, err)

	base := doc.Paths.Find("/subscribe").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value.XML
	require.NotNil(t, base.Origin)
	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/xml.yaml",
			Line:      21,
			Column:    19,
			Name:      "xml",
			EndLine:   23,
			EndColumn: 35,
		},
		base.Origin.Key)

	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/xml.yaml",
			Line:   22,
			Column: 21,
			Name:   "namespace",
		},
		base.Origin.Fields.Get("namespace"))

	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/xml.yaml",
			Line:   23,
			Column: 21,
			Name:   "prefix",
		},
		base.Origin.Fields.Get("prefix"))
}

func TestOrigin_ExampleWithArrayValue(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromFile("testdata/origin/example_with_array.yaml")
	require.NoError(t, err)

	example := doc.Paths.Find("/subscribe").Post.RequestBody.Value.Content["application/json"].Examples["bar"]
	require.NotNil(t, example.Value)

	// The example value is a list of objects and decodes as plain data.
	value := example.Value.Value.(map[string]any)
	require.Len(t, value["items"].([]any), 2)
}

func TestOrigin_WithExternalRef(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true

	loader.Context = t.Context()

	doc, err := loader.LoadFromFile("testdata/origin/external.yaml")
	require.NoError(t, err)

	base := doc.Paths.Find("/subscribe").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value
	require.NotNil(t, base.XML.Origin)
	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/external-schema.yaml",
			Line:      2,
			Column:    1,
			Name:      "xml",
			EndLine:   4,
			EndColumn: 17,
		},
		base.XML.Origin.Key)

	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/external-schema.yaml",
			Line:   3,
			Column: 3,
			Name:   "namespace",
		},
		base.XML.Origin.Fields.Get("namespace"))

	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/external-schema.yaml",
			Line:   4,
			Column: 3,
			Name:   "prefix",
		},
		base.XML.Origin.Fields.Get("prefix"))
}

// The root-level schema of an externally $ref'd file carries an Origin, not
// only the schemas nested inside a parent mapping.
func TestOrigin_WithExternalRefRootOrigin(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = t.Context()

	doc, err := loader.LoadFromFile("testdata/origin/external.yaml")
	require.NoError(t, err)

	// base is the root schema of external-schema.yaml ($ref resolved)
	base := doc.Paths.Find("/subscribe").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value

	// Root schema Origin must now be set (fixed in yaml3 document() injection)
	require.NotNil(t, base.Origin)
	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/external-schema.yaml",
			Line:      1,
			Column:    1,
			Name:      "",
			EndLine:   4,
			EndColumn: 17,
		},
		base.Origin.Key)

	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/external-schema.yaml",
			Line:   1,
			Column: 1,
			Name:   "type",
		},
		base.Origin.Fields.Get("type"))
}

// TestOrigin_RequiredSequence verifies that Origin.Sequences records the
// file/line/column of each item in a required: [...] list.
// These locations are used by NewSourceFromSequenceItem to pinpoint
// breaking changes to individual required field names.
func TestOrigin_RequiredSequence(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true

	doc, err := loader.LoadFromFile("testdata/origin/required_sequence.yaml")
	require.NoError(t, err)

	schema := doc.Paths.Find("/items").Post.RequestBody.Value.Content["application/json"].Schema.Value
	require.NotNil(t, schema.Origin)

	// "required" must appear in Fields (it's a sequence-valued field)
	require.True(t, mustHaveField(schema.Origin.Fields, "required"))

	// Sequences must record per-item locations for "required"
	seqLocs, ok := schema.Origin.Sequences["required"]
	require.True(t, ok, "Origin.Sequences must contain 'required'")
	require.Len(t, seqLocs, 2)

	require.Equal(t, openapi3.Location{
		File:   "testdata/origin/required_sequence.yaml",
		Line:   14,
		Column: 19,
		Name:   "name",
	}, seqLocs[0])

	require.Equal(t, openapi3.Location{
		File:   "testdata/origin/required_sequence.yaml",
		Line:   15,
		Column: 19,
		Name:   "age",
	}, seqLocs[1])
}

// TestOrigin_YAMLAlias verifies that a schema referenced via YAML alias loads
// without error and carries origin metadata from the anchor definition.
// Multiple aliases of the same anchor must each resolve to their own origin.
func TestOrigin_YAMLAlias(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true

	doc, err := loader.LoadFromFile("testdata/origin/alias.yaml")
	require.NoError(t, err)

	anchor := doc.Components.Schemas["Base"].Value
	alias1 := doc.Components.Schemas["Alias1"].Value
	alias2 := doc.Components.Schemas["Alias2"].Value

	// All three point to the same anchor node, so origin reflects the anchor location.
	anchorLoc := &openapi3.Location{
		File:      "testdata/origin/alias.yaml",
		Line:      7,
		Column:    5,
		Name:      "Base",
		EndLine:   13,
		EndColumn: 23,
	}
	require.Equal(t, anchorLoc, anchor.Origin.Key)
	require.Equal(t, anchorLoc, alias1.Origin.Key)
	require.Equal(t, anchorLoc, alias2.Origin.Key)
}

// TestOrigin_Headers verifies that response header origin is tracked correctly.
func TestOrigin_Headers(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true

	doc, err := loader.LoadFromFile("testdata/origin/headers.yaml")
	require.NoError(t, err)

	headers := doc.Paths.Find("/items").Get.Responses.Value("200").Value.Headers

	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/headers.yaml",
			Line:      12,
			Column:    13,
			Name:      "X-Rate-Limit",
			EndLine:   15,
			EndColumn: 30,
		},
		headers["X-Rate-Limit"].Value.Origin.Key)

	require.Equal(t,
		openapi3.Location{
			File:   "testdata/origin/headers.yaml",
			Line:   13,
			Column: 15,
			Name:   "description",
		},
		headers["X-Rate-Limit"].Value.Origin.Fields.Get("description"))

	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/headers.yaml",
			Line:      16,
			Column:    13,
			Name:      "X-Request-Id",
			EndLine:   19,
			EndColumn: 29,
		},
		headers["X-Request-Id"].Value.Origin.Key)
}

// TestOrigin_IntegerStatusCode verifies that response origin is tracked when
// HTTP status codes are written as bare integers (200:) rather than quoted
// strings ("200":). Bare integers produce map[any]any in the
// YAML decoder, which required a dedicated fix in extractOrigins.
func TestOrigin_IntegerStatusCode(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true

	doc, err := loader.LoadFromFile("testdata/origin/parameters.yaml")
	require.NoError(t, err)

	resp200 := doc.Paths.Find("/api/test").Get.Responses.Value("200").Value
	require.NotNil(t, resp200.Origin)
	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/parameters.yaml",
			Line:      14,
			Column:    9,
			Name:      "200",
			EndLine:   15,
			EndColumn: 26,
		},
		resp200.Origin.Key)

	resp201 := doc.Paths.Find("/api/test").Post.Responses.Value("201").Value
	require.NotNil(t, resp201.Origin)
	require.Equal(t,
		&openapi3.Location{
			File:      "testdata/origin/parameters.yaml",
			Line:      18,
			Column:    9,
			Name:      "201",
			EndLine:   19,
			EndColumn: 26,
		},
		resp201.Origin.Key)
}

// TestOrigin_Disabled verifies that all Origin fields are nil when
// IncludeOrigin is false (the default), ensuring no overhead in the common case.
func TestOrigin_Disabled(t *testing.T) {
	loader := openapi3.NewLoader()
	// IncludeOrigin defaults to false

	doc, err := loader.LoadFromFile("testdata/origin/required_sequence.yaml")
	require.NoError(t, err)

	schema := doc.Paths.Find("/items").Post.RequestBody.Value.Content["application/json"].Schema.Value
	require.Nil(t, schema.Origin)
	require.Nil(t, doc.Info.Origin)
	require.Nil(t, doc.Paths.Origin)
}

// TestOrigin_MappingFields verifies that mapping-valued schema fields
// (dependentRequired, dependentSchemas, patternProperties) have their
// key locations tracked in Origin.Fields. Before yaml3 v0.0.12,
// buildOriginSeq only tracked scalar and sequence values, so these
// mapping-valued fields were missing from the origin and source
// location lookups returned nil.
func TestOrigin_MappingFields(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true

	doc, err := loader.LoadFromFile("testdata/origin/mapping_fields.yaml")
	require.NoError(t, err)

	schema := doc.Paths.Find("/test").Get.Responses.Value("200").Value.
		Content["application/json"].Schema.Value.Properties["metadata"].Value
	require.NotNil(t, schema.Origin)

	file := "testdata/origin/mapping_fields.yaml"

	// dependentRequired is a map[string][]string — mapping-valued
	require.True(t, mustHaveField(schema.Origin.Fields, "dependentRequired"))
	require.Equal(t, openapi3.Location{
		File:   file,
		Line:   18,
		Column: 21,
		Name:   "dependentRequired",
	}, schema.Origin.Fields.Get("dependentRequired"))

	// dependentSchemas is a Schemas map — mapping-valued
	require.True(t, mustHaveField(schema.Origin.Fields, "dependentSchemas"))
	require.Equal(t, openapi3.Location{
		File:   file,
		Line:   22,
		Column: 21,
		Name:   "dependentSchemas",
	}, schema.Origin.Fields.Get("dependentSchemas"))

	// patternProperties is a Schemas map — mapping-valued
	require.True(t, mustHaveField(schema.Origin.Fields, "patternProperties"))
	require.Equal(t, openapi3.Location{
		File:   file,
		Line:   25,
		Column: 21,
		Name:   "patternProperties",
	}, schema.Origin.Fields.Get("patternProperties"))
}

// mustHaveField reports whether the named field carries a location, the
// slice-shaped equivalent of asserting a key is present in a map.
func mustHaveField(f openapi3.FieldLocations, name string) bool {
	_, ok := f.Lookup(name)
	return ok
}
