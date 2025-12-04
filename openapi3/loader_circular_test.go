package openapi3

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadCircular(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile("testdata/circularRef2/circular2.yaml")
	require.NoError(t, err)
	require.NotNil(t, doc)

	ref := "./AwsEnvironmentSettings.yaml"

	arr := NewArraySchema()
	obj := NewObjectSchema()
	arr.Items = &SchemaRef{
		Ref:   ref,
		Value: obj,
	}
	arr.Items.setRefPath(&url.URL{Path: "testdata/circularRef2/AwsEnvironmentSettings.yaml"})
	obj.Description = "test"
	obj.Properties = map[string]*SchemaRef{
		"children": {
			Value: arr,
		},
	}

	expected := &SchemaRef{
		Ref:   ref,
		Value: obj,
	}

	actual := doc.Paths.Map()["/sample"].Put.RequestBody.Value.Content.Get("application/json").Schema

	require.Equal(t, expected.Value, actual.Value)
}
