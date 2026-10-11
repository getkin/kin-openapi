package openapi3

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSchemaErrorMessageCustomizerAppliesEverywhere(t *testing.T) {
	customizer := SetSchemaErrorMessageCustomizer(func(err *SchemaError) string {
		return "custom " + err.SchemaField
	})

	requireCustomized := func(t *testing.T, err error, field string) {
		t.Helper()
		require.Error(t, err)
		var schemaErr *SchemaError
		require.True(t, errors.As(err, &schemaErr))
		require.Equal(t, field, schemaErr.SchemaField)
		require.Equal(t, "custom "+field, schemaErr.Error())
	}

	t.Run("discriminator", func(t *testing.T) {
		schema := &Schema{
			OneOf: SchemaRefs{
				{Ref: "#/components/schemas/A", Value: NewObjectSchema()},
			},
			Discriminator: &Discriminator{PropertyName: "kind"},
		}
		err := schema.VisitJSON(map[string]any{}, customizer)
		requireCustomized(t, err, "discriminator")
	})

	t.Run("pattern that does not compile", func(t *testing.T) {
		schema := NewStringSchema()
		schema.Pattern = "^[a-"
		err := schema.VisitJSON("abc", customizer)
		requireCustomized(t, err, "pattern")
	})

	t.Run("JSON Schema 2020-12 validator", func(t *testing.T) {
		schema := NewStringSchema()
		err := schema.VisitJSON(42, customizer, EnableJSONSchema2020())
		requireCustomized(t, err, "type")
	})
}
