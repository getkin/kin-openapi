package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue1243FormatValidatorsReachJSONSchema2020(t *testing.T) {
	schema := &openapi3.Schema{
		Type:   &openapi3.Types{"string"},
		Format: "uuid",
	}

	uuid := openapi3.WithStringFormatValidator("uuid", openapi3.NewRegexpFormatValidator(openapi3.FormatOfStringForUUIDOfRFC4122))

	t.Run("per-validation validator asserts", func(t *testing.T) {
		require.NoError(t, schema.VisitJSON("9d1c8f74-4e5b-41f9-a2f5-6b1f0b6dbb9e", openapi3.EnableJSONSchema2020(), uuid))
		require.ErrorContains(t, schema.VisitJSON("not-a-uuid", openapi3.EnableJSONSchema2020(), uuid), "uuid")
	})

	t.Run("global validator asserts", func(t *testing.T) {
		openapi3.DefineStringFormatValidator("uuid", openapi3.NewRegexpFormatValidator(openapi3.FormatOfStringForUUIDOfRFC4122))
		defer delete(openapi3.SchemaStringFormats, "uuid")

		require.NoError(t, schema.VisitJSON("9d1c8f74-4e5b-41f9-a2f5-6b1f0b6dbb9e", openapi3.EnableJSONSchema2020()))
		require.ErrorContains(t, schema.VisitJSON("not-a-uuid", openapi3.EnableJSONSchema2020()), "uuid")
	})

	t.Run("unregistered format stays an annotation", func(t *testing.T) {
		require.NoError(t, schema.VisitJSON("not-a-uuid", openapi3.EnableJSONSchema2020()))
	})

	t.Run("integer format asserts", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:   &openapi3.Types{"integer"},
			Format: "int32",
		}

		require.NoError(t, schema.VisitJSON(float64(42), openapi3.EnableJSONSchema2020()))
		require.ErrorContains(t, schema.VisitJSON(float64(1<<31), openapi3.EnableJSONSchema2020()), "int32")
	})
}
