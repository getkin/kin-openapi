package openapi3

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrigin_All(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = context.Background()

	const dir = "testdata/origin/"
	items, _ := os.ReadDir(dir)
	for _, item := range items {
		t.Run(item.Name(), func(t *testing.T) {
			doc, err := loader.LoadFromFile(fmt.Sprintf("%s/%s", dir, item.Name()))
			require.NoError(t, err)
			if doc.Paths == nil {
				t.Skip("no paths")
			}
			require.NotEmpty(t, doc.Paths.Origin)
		})
	}
}

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
	loader.IncludeOrigin = true
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
	loader.IncludeOrigin = true
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
	loader.IncludeOrigin = true
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
	loader.IncludeOrigin = true
	loader.Context = context.Background()

	doc, err := loader.LoadFromFile("testdata/origin/external_docs.yaml")
	require.NoError(t, err)

	base := doc.ExternalDocs.Origin
	require.NotNil(t, base)

	require.Equal(t,
		&Location{
			Line:   13,
			Column: 1,
		},
		base.Key)

	require.Equal(t,
		Location{
			Line:   14,
			Column: 3,
		},
		base.Fields["description"])

	require.Equal(t,
		Location{
			Line:   15,
			Column: 3,
		},
		base.Fields["url"])
}
