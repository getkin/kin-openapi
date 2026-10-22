package openapi3_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func loadPathItemRefSiblings(t *testing.T, version, siblings string) (*openapi3.T, error) {
	t.Helper()
	loader := openapi3.NewLoader()
	return loader.LoadFromData(fmt.Appendf(nil, `{
  "openapi": %q,
  "info": {"title": "t", "version": "1"},
  "paths": {
    "/x": {"$ref": "#/paths/~1y", %s},
    "/y": {
      "summary": "target",
      "get": {"responses": {"200": {"description": "target get"}}}
    }
  }
}`, version, siblings))
}

// A Path Item Object's $ref is not a Reference Object: the fields written next
// to it are part of the same path item. The specification leaves only the
// result for a field set on both sides undefined.
func TestPathItemRefSiblings(t *testing.T) {
	t.Parallel()
	for _, version := range []string{"3.0.3", "3.1.0", "3.2.0"} {
		t.Run(version, func(t *testing.T) {
			t.Parallel()
			t.Run("fields set on one side only are merged", func(t *testing.T) {
				t.Parallel()
				doc, err := loadPathItemRefSiblings(t, version, `"post": {"responses": {"200": {"description": "local post"}}}`)
				require.NoError(t, err)
				require.NoError(t, doc.Validate(t.Context()))

				x := doc.Paths.Value("/x")
				require.Equal(t, "target", x.Summary)
				require.Equal(t, "target get", *x.Get.Responses.Value("200").Value.Description)
				require.Equal(t, "local post", *x.Post.Responses.Value("200").Value.Description)

				// The referenced path item is not changed.
				require.Nil(t, doc.Paths.Value("/y").Post)
			})

			t.Run("a field set on both sides is taken from next to the $ref", func(t *testing.T) {
				t.Parallel()
				doc, err := loadPathItemRefSiblings(t, version, `"summary": "local", "get": {"responses": {"200": {"description": "local get"}}}`)
				require.NoError(t, err)
				require.NoError(t, doc.Validate(t.Context()))

				x := doc.Paths.Value("/x")
				require.Equal(t, "local", x.Summary)
				require.Equal(t, "local get", *x.Get.Responses.Value("200").Value.Description)

				y := doc.Paths.Value("/y")
				require.Equal(t, "target", y.Summary)
				require.Equal(t, "target get", *y.Get.Responses.Value("200").Value.Description)
			})

			t.Run("only the fields next to the $ref are marshaled back", func(t *testing.T) {
				t.Parallel()
				doc, err := loadPathItemRefSiblings(t, version, `"summary": "local", "x-local": 1`)
				require.NoError(t, err)

				data, err := doc.Paths.Value("/x").MarshalJSON()
				require.NoError(t, err)
				require.JSONEq(t, `{"$ref": "#/paths/~1y", "summary": "local", "x-local": 1}`, string(data))

				data, err = doc.MarshalJSON()
				require.NoError(t, err)
				loader := openapi3.NewLoader()
				reloaded, err := loader.LoadFromData(data)
				require.NoError(t, err)
				require.Equal(t, "local", reloaded.Paths.Value("/x").Summary)
				require.NotNil(t, reloaded.Paths.Value("/x").Get)
			})

			t.Run("a $ref to nothing is reported despite the siblings", func(t *testing.T) {
				t.Parallel()
				loader := openapi3.NewLoader()
				_, err := loader.LoadFromData(fmt.Appendf(nil, `{
  "openapi": %q,
  "info": {"title": "t", "version": "1"},
  "paths": {"/x": {"$ref": "#/paths/~1nowhere", "summary": "local"}}
}`, version))
				require.Error(t, err)
			})
		})
	}
}
