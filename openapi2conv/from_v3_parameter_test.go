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
		require.Equal(t, &openapi2.Parameter{Ref: "#/parameters/File"}, got)
	})

	t.Run("inline binary schema", func(t *testing.T) {
		t.Parallel()

		param := &openapi3.ParameterRef{Value: &openapi3.Parameter{
			Name:        "token",
			In:          "query",
			Description: "an opaque blob",
			Required:    true,
			Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{
				Type:   &openapi3.Types{"string"},
				Format: "binary",
			}},
		}}
		got, err := openapi2conv.FromV3Parameter(param, nil)
		require.NoError(t, err)
		require.Equal(t, &openapi2.Parameter{
			Name:        "token",
			In:          "query",
			Description: "an opaque blob",
			Required:    true,
			Type:        &openapi3.Types{"string"},
			Format:      "binary",
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

	t.Run("binary schema keeps the keywords the other branch keeps", func(t *testing.T) {
		t.Parallel()

		// A schema may hold keywords that say nothing about a string. The
		// conversion of a schema that is not binary carries them over, so this
		// one does too.
		schema := &openapi3.Schema{
			Type:        &openapi3.Types{"string"},
			Format:      "binary",
			Pattern:     "^[a-z]+$",
			MultipleOf:  openapi3.Ptr(2.0),
			UniqueItems: true,
			MinItems:    1,
			MaxItems:    openapi3.Ptr(uint64(3)),
			Items:       &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
		}
		param := &openapi3.ParameterRef{Value: &openapi3.Parameter{
			Name: "blob", In: "query",
			Schema: &openapi3.SchemaRef{Value: schema},
		}}
		got, err := openapi2conv.FromV3Parameter(param, nil)
		require.NoError(t, err)
		require.Equal(t, "^[a-z]+$", got.Pattern)
		require.Equal(t, openapi3.Ptr(2.0), got.MultipleOf)
		require.True(t, got.UniqueItems)
		require.EqualValues(t, 1, got.MinItems)
		require.Equal(t, openapi3.Ptr(uint64(3)), got.MaxItems)
		require.NotNil(t, got.Items)
		require.Equal(t, &openapi3.Types{"string"}, got.Items.Value.Type)
		require.Empty(t, got.CollectionFormat)
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
