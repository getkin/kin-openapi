package openapi2conv_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

func TestFromV3ParameterSchemas(t *testing.T) {
	t.Parallel()

	t.Run("ref to a binary schema", func(t *testing.T) {
		t.Parallel()

		components := &openapi3.Components{Schemas: openapi3.Schemas{
			"File": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "binary"}},
		}}
		param := &openapi3.ParameterRef{Value: &openapi3.Parameter{
			Name:   "file",
			In:     "query",
			Schema: &openapi3.SchemaRef{Ref: "#/components/schemas/File"},
		}}
		got, err := openapi2conv.FromV3Parameter(param, components)
		require.NoError(t, err)
		require.Equal(t, &openapi2.Parameter{
			Name:   "file",
			In:     "query",
			Type:   &openapi3.Types{"string"},
			Format: "binary",
		}, got)
	})

	t.Run("binary header keeps its location and name", func(t *testing.T) {
		t.Parallel()

		param := &openapi3.ParameterRef{Value: &openapi3.Parameter{
			Name:     "X-Blob",
			In:       "header",
			Required: true,
			Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{
				Type:      &openapi3.Types{"string"},
				Format:    "binary",
				MaxLength: openapi3.Ptr(uint64(64)),
			}},
		}}
		got, err := openapi2conv.FromV3Parameter(param, nil)
		require.NoError(t, err)
		require.Equal(t, &openapi2.Parameter{
			Name:      "X-Blob",
			In:        "header",
			Required:  true,
			Type:      &openapi3.Types{"string"},
			Format:    "binary",
			MaxLength: openapi3.Ptr(uint64(64)),
		}, got)
	})

	t.Run("schema without a value", func(t *testing.T) {
		t.Parallel()

		param := &openapi3.ParameterRef{Value: &openapi3.Parameter{
			Name:   "p",
			In:     "query",
			Schema: &openapi3.SchemaRef{},
		}}
		got, err := openapi2conv.FromV3Parameter(param, nil)
		require.NoError(t, err)
		require.Equal(t, &openapi2.Parameter{Name: "p", In: "query"}, got)
	})

}
