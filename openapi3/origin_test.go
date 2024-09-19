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

	require.NotNil(t, doc.Paths.Find("/partner-api/test/another-method").Origin)
	require.Equal(t,
		&Location{
			Line:   13,
			Column: 3,
		},
		doc.Paths.Find("/partner-api/test/another-method").Origin.Key)

	require.NotNil(t, doc.Paths.Find("/partner-api/test/another-method").Get.Origin)
	require.Equal(t,
		&Location{
			Line:   14,
			Column: 5,
		},
		doc.Paths.Find("/partner-api/test/another-method").Get.Origin.Key)
}

func TestOrigin_SchemaInAdditionalProperties(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = context.Background()

	doc, err := loader.LoadFromFile("testdata/origin/additional_properties.yaml")
	require.NoError(t, err)

	require.NotNil(t, doc.Paths.Find("/partner-api/test/some-method").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value.AdditionalProperties)
	additionalProperties := doc.Paths.Find("/partner-api/test/some-method").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value.AdditionalProperties

	require.NotNil(t, additionalProperties.Schema.Value.Origin)
	require.Equal(t,
		&Location{
			Line:   14,
			Column: 17,
		},
		additionalProperties.Schema.Value.Origin.Key)

	require.Equal(t,
		Location{
			Line:   15,
			Column: 19,
		},
		additionalProperties.Schema.Value.Origin.Fields["type"])
}

func TestOrigin_ExternalDocs(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = context.Background()

	doc, err := loader.LoadFromFile("testdata/origin/external_docs.yaml")
	require.NoError(t, err)

	require.NotNil(t, doc.ExternalDocs.Origin)

	require.Equal(t,
		&Location{
			Line:   13,
			Column: 1,
		},
		doc.ExternalDocs.Origin.Key)

	require.Equal(t,
		Location{
			Line:   14,
			Column: 3,
		},
		doc.ExternalDocs.Origin.Fields["description"])

	require.Equal(t,
		Location{
			Line:   15,
			Column: 3,
		},
		doc.ExternalDocs.Origin.Fields["url"])
}
