package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestSchemaValidate31SubSchemas(t *testing.T) {
	ctx := openapi3.WithValidationOptions(t.Context(), openapi3.IsOpenAPI31OrLater())

	// Helper: a schema with an invalid nested schema (pattern with bad regex)
	invalidSchema := &openapi3.Schema{
		Type:    &openapi3.Types{"string"},
		Pattern: "[invalid",
	}

	t.Run("prefixItems with invalid sub-schema", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:        &openapi3.Types{"array"},
			PrefixItems: openapi3.SchemaRefs{{Value: invalidSchema}},
		}
		err := schema.Validate(ctx)
		require.Error(t, err, "should detect invalid sub-schema in prefixItems")
	})

	t.Run("contains with invalid sub-schema", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:     &openapi3.Types{"array"},
			Contains: &openapi3.SchemaRef{Value: invalidSchema},
		}
		err := schema.Validate(ctx)
		require.Error(t, err, "should detect invalid sub-schema in contains")
	})

	t.Run("patternProperties with invalid sub-schema", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			PatternProperties: openapi3.Schemas{
				"^x-": {Value: invalidSchema},
			},
		}
		err := schema.Validate(ctx)
		require.Error(t, err, "should detect invalid sub-schema in patternProperties")
	})

	t.Run("dependentSchemas with invalid sub-schema", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			DependentSchemas: openapi3.Schemas{
				"name": {Value: invalidSchema},
			},
		}
		err := schema.Validate(ctx)
		require.Error(t, err, "should detect invalid sub-schema in dependentSchemas")
	})

	t.Run("propertyNames with invalid sub-schema", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:          &openapi3.Types{"object"},
			PropertyNames: &openapi3.SchemaRef{Value: invalidSchema},
		}
		err := schema.Validate(ctx)
		require.Error(t, err, "should detect invalid sub-schema in propertyNames")
	})

	t.Run("unevaluatedItems with invalid sub-schema", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:             &openapi3.Types{"array"},
			UnevaluatedItems: openapi3.BoolSchema{Schema: &openapi3.SchemaRef{Value: invalidSchema}},
		}
		err := schema.Validate(ctx)
		require.Error(t, err, "should detect invalid sub-schema in unevaluatedItems")
	})

	t.Run("unevaluatedProperties with invalid sub-schema", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:                  &openapi3.Types{"object"},
			UnevaluatedProperties: openapi3.BoolSchema{Schema: &openapi3.SchemaRef{Value: invalidSchema}},
		}
		err := schema.Validate(ctx)
		require.Error(t, err, "should detect invalid sub-schema in unevaluatedProperties")
	})

	t.Run("contentSchema with invalid sub-schema", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:             &openapi3.Types{"string"},
			ContentMediaType: "application/json",
			ContentSchema:    &openapi3.SchemaRef{Value: invalidSchema},
		}
		err := schema.Validate(ctx)
		require.Error(t, err, "should detect invalid sub-schema in contentSchema")
	})

	t.Run("valid 3.1 sub-schemas pass validation", func(t *testing.T) {
		validSubSchema := &openapi3.Schema{Type: &openapi3.Types{"string"}}
		schema := &openapi3.Schema{
			Type:  &openapi3.Types{"array"},
			Items: &openapi3.SchemaRef{Value: validSubSchema},
			PrefixItems: openapi3.SchemaRefs{
				{Value: validSubSchema},
			},
			Contains:         &openapi3.SchemaRef{Value: validSubSchema},
			UnevaluatedItems: openapi3.BoolSchema{Schema: &openapi3.SchemaRef{Value: validSubSchema}},
		}
		err := schema.Validate(ctx)
		require.NoError(t, err, "valid sub-schemas should pass validation")
	})
}
