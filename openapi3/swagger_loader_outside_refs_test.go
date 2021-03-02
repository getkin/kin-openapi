package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadOutsideRefs(t *testing.T) {
	loader := NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadSwaggerFromFile("testdata/303bis/service.yaml")
	require.NoError(t, err)
	require.NotNil(t, doc)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	props := doc.Paths["/service"].Get.Responses["200"].Value.Content["application/json"].Schema.Value.Items.Value.Properties
	require.NotNil(t, props)
	require.Equal(t, "string", props["created_at"].Value.Type)
}
