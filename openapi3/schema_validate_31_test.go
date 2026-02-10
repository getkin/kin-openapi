package openapi3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSchemaValidate31SubSchemas(t *testing.T) {
	ctx := context.Background()

	// Helper: a schema with an invalid nested schema (pattern with bad regex)
	invalidSchema := &Schema{
		Type:    &Types{"string"},
		Pattern: "[invalid",
	}

	t.Run("prefixItems with invalid sub-schema", func(t *testing.T) {
		schema := &Schema{
			Type: &Types{"array"},
			PrefixItems: SchemaRefs{
				{Value: invalidSchema},
			},
		}
		err := schema.Validate(ctx)
		require.Error(t, err, "should detect invalid sub-schema in prefixItems")
	})

	t.Run("contains with invalid sub-schema", func(t *testing.T) {
		schema := &Schema{
			Type:     &Types{"array"},
			Contains: &SchemaRef{Value: invalidSchema},
		}
		err := schema.Validate(ctx)
		require.Error(t, err, "should detect invalid sub-schema in contains")
	})

	t.Run("patternProperties with invalid sub-schema", func(t *testing.T) {
		schema := &Schema{
			Type: &Types{"object"},
			PatternProperties: Schemas{
				"^x-": {Value: invalidSchema},
			},
		}
		err := schema.Validate(ctx)
		require.Error(t, err, "should detect invalid sub-schema in patternProperties")
	})

	t.Run("dependentSchemas with invalid sub-schema", func(t *testing.T) {
		schema := &Schema{
			Type: &Types{"object"},
			DependentSchemas: Schemas{
				"name": {Value: invalidSchema},
			},
		}
		err := schema.Validate(ctx)
		require.Error(t, err, "should detect invalid sub-schema in dependentSchemas")
	})

	t.Run("propertyNames with invalid sub-schema", func(t *testing.T) {
		schema := &Schema{
			Type:          &Types{"object"},
			PropertyNames: &SchemaRef{Value: invalidSchema},
		}
		err := schema.Validate(ctx)
		require.Error(t, err, "should detect invalid sub-schema in propertyNames")
	})

	t.Run("unevaluatedItems with invalid sub-schema", func(t *testing.T) {
		schema := &Schema{
			Type:             &Types{"array"},
			UnevaluatedItems: &SchemaRef{Value: invalidSchema},
		}
		err := schema.Validate(ctx)
		require.Error(t, err, "should detect invalid sub-schema in unevaluatedItems")
	})

	t.Run("unevaluatedProperties with invalid sub-schema", func(t *testing.T) {
		schema := &Schema{
			Type:                  &Types{"object"},
			UnevaluatedProperties: &SchemaRef{Value: invalidSchema},
		}
		err := schema.Validate(ctx)
		require.Error(t, err, "should detect invalid sub-schema in unevaluatedProperties")
	})

	t.Run("contentSchema with invalid sub-schema", func(t *testing.T) {
		schema := &Schema{
			Type:             &Types{"string"},
			ContentMediaType: "application/json",
			ContentSchema:    &SchemaRef{Value: invalidSchema},
		}
		err := schema.Validate(ctx)
		require.Error(t, err, "should detect invalid sub-schema in contentSchema")
	})

	t.Run("valid 3.1 sub-schemas pass validation", func(t *testing.T) {
		validSubSchema := &Schema{Type: &Types{"string"}}
		schema := &Schema{
			Type:  &Types{"array"},
			Items: &SchemaRef{Value: validSubSchema},
			PrefixItems: SchemaRefs{
				{Value: validSubSchema},
			},
			Contains:         &SchemaRef{Value: validSubSchema},
			UnevaluatedItems: &SchemaRef{Value: validSubSchema},
		}
		err := schema.Validate(ctx)
		require.NoError(t, err, "valid sub-schemas should pass validation")
	})
}
