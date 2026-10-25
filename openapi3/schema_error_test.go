package openapi3

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSchemaErrorMessageCustomizerAppliesEverywhere(t *testing.T) {
	t.Parallel()

	customizer := SetSchemaErrorMessageCustomizer(func(err *SchemaError) string {
		return "custom " + err.SchemaField
	})

	requireCustomized := func(t *testing.T, err error, field string) {
		t.Helper()
		require.Error(t, err)
		var schemaErr *SchemaError
		require.True(t, errors.As(err, &schemaErr))
		require.Equal(t, field, schemaErr.SchemaField)
		require.EqualError(t, schemaErr, "custom "+field)
	}

	t.Run("discriminator", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
		schema := NewStringSchema()
		schema.Pattern = "^[a-"
		err := schema.VisitJSON("abc", customizer)
		requireCustomized(t, err, "pattern")
	})

	t.Run("JSON Schema 2020-12 validator", func(t *testing.T) {
		t.Parallel()
		schema := NewStringSchema()
		err := schema.VisitJSON(42, customizer, EnableJSONSchema2020())
		requireCustomized(t, err, "type")
	})
}

func TestDisableSchemaErrorDetails(t *testing.T) {
	t.Parallel()

	schema := NewStringSchema()
	schema.Pattern = "[A-Z]"

	err := schema.VisitJSON("secret")
	require.ErrorContains(t, err, "Value:")
	require.ErrorContains(t, err, "secret")

	err = schema.VisitJSON("secret", DisableSchemaErrorDetails())
	require.EqualError(t, err, `string doesn't match the regular expression "[A-Z]"`)
}

func TestDisableSchemaErrorDetailsOfFormatValidators(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		validate func(string) error
		message  string
	}{
		"direct": {
			validate: func(s string) error {
				return newSchemaError(nil, s).withReason("bad")
			},
			message: `string doesn't match the format "custom": bad`,
		},
		"wrapped": {
			validate: func(s string) error {
				return fmt.Errorf("wrapped: %w", newSchemaError(nil, s).withReason("bad"))
			},
			message: `string doesn't match the format "custom": bad`,
		},
		"joined": {
			validate: func(s string) error {
				return errors.Join(errors.New("other"), fmt.Errorf("wrapped: %w", newSchemaError(nil, s).withReason("bad")))
			},
			message: `string doesn't match the format "custom": bad`,
		},
		"multi": {
			validate: func(s string) error {
				return fmt.Errorf("wrapped: %w", MultiError{newSchemaError(nil, s).withReason("bad")})
			},
			message: `string doesn't match the format "custom": bad`,
		},
		"several": {
			validate: func(s string) error {
				return errors.Join(fmt.Errorf("wrapped: %w", newSchemaError(nil, s).withReason("bad")),
					MultiError{newSchemaError(nil, s).withReason("worse")},
				)
			},
			message: `string doesn't match the format "custom": bad | worse`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			schema := NewStringSchema().WithFormat("custom")
			formats := WithStringFormatValidators(map[string]StringFormatValidator{
				"custom": NewCallbackValidator(tc.validate),
			})

			err := schema.VisitJSON("secret", formats)
			require.ErrorContains(t, err, "secret")

			err = schema.VisitJSON("secret", formats, DisableSchemaErrorDetails())
			require.EqualError(t, err, tc.message)
		})
	}
}

func TestDisableSchemaErrorDetailsWithJSONSchema2020(t *testing.T) {
	t.Parallel()

	schema := NewStringSchema()

	err := schema.VisitJSON("valid", DisableSchemaErrorDetails(), EnableJSONSchema2020())
	require.ErrorIs(t, err, ErrSchemaErrorDetailsWithJSONSchema2020)

	err = schema.VisitJSON("valid", DisableSchemaErrorDetails())
	require.NoError(t, err)
}
