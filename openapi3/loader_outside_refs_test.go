package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadOutsideRefs(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile("testdata/303bis/service.yaml")
	require.NoError(t, err)
	require.NotNil(t, doc)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	require.Equal(t, "string", doc.Paths["/service"].Get.Responses["200"].Value.Content["application/json"].Schema.Value.Items.Value.AllOf[0].Value.Properties["created_at"].Value.Type)
}
