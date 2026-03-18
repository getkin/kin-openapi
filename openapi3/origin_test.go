package openapi3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func unsetIncludeOrigin() {
	IncludeOrigin = false
}

func TestOrigin_Info(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true

	IncludeOrigin = true
	defer unsetIncludeOrigin()

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

	IncludeOrigin = true
	defer unsetIncludeOrigin()

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

	IncludeOrigin = true
	defer unsetIncludeOrigin()

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

	IncludeOrigin = true
	defer unsetIncludeOrigin()

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

	IncludeOrigin = true
	defer unsetIncludeOrigin()

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

	IncludeOrigin = true
	defer unsetIncludeOrigin()

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

	IncludeOrigin = true
	defer unsetIncludeOrigin()

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

	IncludeOrigin = true
	defer unsetIncludeOrigin()

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

	IncludeOrigin = true
	defer unsetIncludeOrigin()

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

	IncludeOrigin = true
	defer unsetIncludeOrigin()

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

func TestStripOriginFromAny_Slice(t *testing.T) {
	// Simulates what the YAML origin-tracking loader produces for example
	// values that contain arrays of objects.
	input := map[string]any{
		"name": "test",
		"items": []any{
			map[string]any{
				"__origin__": map[string]any{"file": "a.yaml", "line": 1},
				"id":         1,
			},
			map[string]any{
				"__origin__": map[string]any{"file": "a.yaml", "line": 2},
				"id":         2,
			},
		},
	}

	result := stripOriginFromAny(input)
	m := result.(map[string]any)
	items := m["items"].([]any)
	for _, item := range items {
		itemMap := item.(map[string]any)
		require.NotContains(t, itemMap, "__origin__")
	}
}

func TestOrigin_ExampleWithArrayValue(t *testing.T) {
	loader := NewLoader()

	IncludeOrigin = true
	defer unsetIncludeOrigin()

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

	IncludeOrigin = true
	defer unsetIncludeOrigin()

	_, err := loader.LoadFromData([]byte(data))
	require.Error(t, err)
	require.Equal(t, `failed to unmarshal data: json error: invalid character 'p' looking for beginning of value, yaml error: error converting YAML to JSON: yaml: unmarshal errors:
  line 0: mapping key "__origin__" already defined at line 17`, err.Error())
}

func TestOrigin_WithExternalRef(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true

	IncludeOrigin = true
	defer unsetIncludeOrigin()

	loader.Context = context.Background()

	doc, err := loader.LoadFromFile("testdata/origin/external.yaml")
	require.NoError(t, err)

	base := doc.Paths.Find("/subscribe").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value
	require.NotNil(t, base.XML.Origin)
	require.Equal(t, base.XML.Origin.Key.File, "testdata/origin/external-schema.yaml")
	// require.Equal(t,
	// 	&Location{
	// 		Line:   1,
	// 		Column: 1,
	// 	},
	// 	base.Origin.Key)

	// require.Equal(t,
	// 	Location{
	// 		Line:   2,
	// 		Column: 3,
	// 	},
	// 	base.Origin.Fields["namespace"])

	// require.Equal(t,
	// 	Location{
	// 		Line:   3,
	// 		Column: 3,
	// 	},
	// 	base.Origin.Fields["prefix"])
}
