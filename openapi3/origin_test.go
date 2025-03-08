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
			Line:   2,
			Column: 1,
		},
		doc.Info.Origin.Key)

	require.Equal(t,
		Location{
			Line:   3,
			Column: 3,
		},
		doc.Info.Origin.Fields["title"])

	require.Equal(t,
		Location{
			Line:   4,
			Column: 3,
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
			Line:   5,
			Column: 1,
		},
		doc.Paths.Origin.Key)

	base := doc.Paths.Find("/partner-api/test/another-method")

	require.NotNil(t, base.Origin)
	require.Equal(t,
		&Location{
			Line:   13,
			Column: 3,
		},
		base.Origin.Key)

	require.NotNil(t, base.Get.Origin)
	require.Equal(t,
		&Location{
			Line:   14,
			Column: 5,
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
			Line:   8,
			Column: 7,
		},
		base.Origin.Key)

	require.NotNil(t, base.Content["application/json"].Origin)
	require.Equal(t,
		&Location{
			Line:   10,
			Column: 11,
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
			Line:   17,
			Column: 7,
		},
		base.Origin.Key)

	require.NotNil(t, base.Origin)
	require.Nil(t, base.Value("200").Origin)
	require.Equal(t,
		&Location{
			Line:   18,
			Column: 9,
		},
		base.Value("200").Value.Origin.Key)

	require.Equal(t,
		Location{
			Line:   19,
			Column: 11,
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
			Line:   9,
			Column: 11,
		},
		base.Origin.Key)

	require.Equal(t,
		Location{
			Line:   10,
			Column: 11,
		},
		base.Origin.Fields["in"])

	require.Equal(t,
		Location{
			Line:   9,
			Column: 11,
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
			Line:   14,
			Column: 17,
		},
		base.Schema.Value.Origin.Key)

	require.Equal(t,
		Location{
			Line:   15,
			Column: 19,
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
			Line:   13,
			Column: 1,
		},
		base.Origin.Key)

	require.Equal(t,
		Location{
			Line:   14,
			Column: 3,
		},
		base.Origin.Fields["description"])

	require.Equal(t,
		Location{
			Line:   15,
			Column: 3,
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
			Line:   29,
			Column: 5,
		},
		base.Origin.Key)

	require.Equal(t,
		Location{
			Line:   30,
			Column: 7,
		},
		base.Origin.Fields["type"])

	require.Equal(t,
		&Location{
			Line:   31,
			Column: 7,
		},
		base.Flows.Origin.Key)

	require.Equal(t,
		&Location{
			Line:   32,
			Column: 9,
		},
		base.Flows.Implicit.Origin.Key)

	require.Equal(t,
		Location{
			Line:   33,
			Column: 11,
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
			Line:   14,
			Column: 15,
		},
		base.Origin.Key)

	require.Equal(t,
		Location{
			Line:   15,
			Column: 17,
		},
		base.Origin.Fields["summary"])

	//	Note:
	//  Example.Value contains an extra field: "origin".
	//
	//	Explanation:
	//  The example value is defined in the original yaml file as a json object: {"bar": "baz"}
	//  This json object is also valid in YAML, so yaml.3 decodes it as a map and adds an "origin" field.
	require.Contains(t,
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
			Line:   21,
			Column: 19,
		},
		base.Origin.Key)

	require.Equal(t,
		Location{
			Line:   22,
			Column: 21,
		},
		base.Origin.Fields["namespace"])

	require.Equal(t,
		Location{
			Line:   23,
			Column: 21,
		},
		base.Origin.Fields["prefix"])
}

// TestOrigin_OriginExistsInProperties verifies that we don't remove pre-existing "origin" fields.
// See https://github.com/getkin/kin-openapi/issues/1051
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
        origin:
          type: string
`

	loader := NewLoader()
	doc, err := loader.LoadFromData([]byte(data))
	require.NoError(t, err)

	require.NotEmpty(t, doc.Paths.Find("/foo").Get.Responses.Status(200).Value.Content.Get("application/json").Schema.Value.Properties["origin"])
}

// TestOrigin_OriginExistsInContent verifies that we don't remove pre-existing "origin" fields.
// See https://github.com/getkin/kin-openapi/issues/1051
func TestOrigin_OriginExistsInContent(t *testing.T) {
	var data = `
paths:
  /foo:
    get:
      responses:
        "200":
          description: OK
          content:
            origin:
              schema:
                $ref: "#/components/schemas/Foo"
components:
  schemas:
    Foo:
      type: string
`

	loader := NewLoader()
	doc, err := loader.LoadFromData([]byte(data))
	require.NoError(t, err)

	require.NotEmpty(t, doc.Paths.Find("/foo").Get.Responses.Status(200).Value.Content.Get("origin"))
}

// TestOrigin_OriginExistsInPropertiesWithIncludeOrigin demonstrates a limitation in the current implementation:
// Include origin doesn't work if properties (or any other map) contains a key named "origin".
func TestOrigin_OriginExistsInPropertiesWithIncludeOrigin(t *testing.T) {
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
        origin:
          type: string
`

	loader := NewLoader()

	IncludeOrigin = true
	defer unsetIncludeOrigin()

	_, err := loader.LoadFromData([]byte(data))
	require.Equal(t, `failed to unmarshal data: json error: invalid character 'p' looking for beginning of value, yaml error: error converting YAML to JSON: yaml: unmarshal errors:
  line 0: mapping key "origin" already defined at line 17`, err.Error())
}
