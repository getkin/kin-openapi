package openapi3

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultRefNameResolverDocumentExtension(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		path, want string
	}{
		{"schemas/Item.yaml", "schemas_Item"},
		{"schemas/Item.yml", "schemas_Item"},
		{"schemas/Item.json", "schemas_Item"},
		{"schemas/Models.User.yaml", "schemas_Models.User"},
		{"schemas/Models.User.YAML", "schemas_Models.User"},
		{"schemas/Models.User.JsOn?version=1#/components/schemas/Item", "schemas_Models.User_Item"},
		{"schemas/Models.User", "schemas_Models.User"},
		{"schemas/Models.User.schema", "schemas_Models.User.schema"},
		{"schemas/Models.User.json.yaml", "schemas_Models.User.json"},
		{"schemas/Models.User.yaml.json", "schemas_Models.User.yaml"},
		{"schemas.v1/Item.yaml", "schemas.v1_Item"},
		{"schemas/Item", "schemas_Item"},
		{"schemas/Item.yaml#/components/schemas/Models.User", "schemas_Item_Models.User"},
		{"schemas/Models.User.json#/components/schemas/Item", "schemas_Models.User_Item"},
	} {
		t.Run(test.path, func(t *testing.T) {
			t.Parallel()
			uri, err := url.Parse(test.path)
			require.NoError(t, err)
			ref := &SchemaRef{Ref: test.path}
			ref.setRefPath(uri)
			require.Equal(t, test.want, DefaultRefNameResolver(&T{}, ref))
		})
	}
}

func TestDefaultRefNameResolverRootDocument(t *testing.T) {
	t.Parallel()
	root := &url.URL{Path: "/specs/root.yaml"}
	ref := &SchemaRef{Ref: "#/components/schemas/Models.User"}
	ref.setRefPath(&url.URL{Path: root.Path, Fragment: "/components/schemas/Models.User"})
	require.Equal(t, "Models.User", DefaultRefNameResolver(&T{url: root}, ref))
}
