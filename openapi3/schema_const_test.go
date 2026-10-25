package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// const is OpenAPI 3.1 only, so it is validated by the JSON Schema 2020-12
// validator; the built-in one ignores it.
func TestSchemaConst_BuiltInValidatorIgnoresIt(t *testing.T) {
	t.Parallel()

	schema := &openapi3.Schema{Const: "production"}
	err := schema.VisitJSON("development")
	require.NoError(t, err)
	err = schema.VisitJSON("development", openapi3.EnableJSONSchema2020())
	require.Error(t, err)
}

func TestSchemaConst_JSONSchema2020(t *testing.T) {
	t.Parallel()

	opt := openapi3.EnableJSONSchema2020()

	t.Run("string const", func(t *testing.T) {
		schema := &openapi3.Schema{
			Const: "production",
		}

		err := schema.VisitJSON("production", opt)
		require.NoError(t, err)

		err = schema.VisitJSON("development", opt)
		require.Error(t, err)
		require.ErrorContains(t, err, "production")
	})

	t.Run("number const", func(t *testing.T) {
		schema := &openapi3.Schema{
			Const: float64(42),
		}

		err := schema.VisitJSON(float64(42), opt)
		require.NoError(t, err)

		err = schema.VisitJSON(float64(43), opt)
		require.Error(t, err)
	})

	t.Run("boolean const", func(t *testing.T) {
		schema := &openapi3.Schema{
			Const: true,
		}

		err := schema.VisitJSON(true, opt)
		require.NoError(t, err)

		err = schema.VisitJSON(false, opt)
		require.Error(t, err)
	})

	t.Run("null const", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:  &openapi3.Types{"null"},
			Const: nil,
		}

		// nil const means "not set", so this should pass as empty schema
		err := schema.VisitJSON(nil, opt)
		require.NoError(t, err)
	})

	t.Run("object const", func(t *testing.T) {
		schema := &openapi3.Schema{
			Const: map[string]any{"key": "value"},
		}

		err := schema.VisitJSON(map[string]any{"key": "value"}, opt)
		require.NoError(t, err)

		err = schema.VisitJSON(map[string]any{"key": "other"}, opt)
		require.Error(t, err)
	})

	t.Run("const with type constraint", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:  &openapi3.Types{"string"},
			Const: "fixed",
		}

		err := schema.VisitJSON("fixed", opt)
		require.NoError(t, err)

		err = schema.VisitJSON("other", opt)
		require.Error(t, err)
	})

	t.Run("const with multiError", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:  &openapi3.Types{"string"},
			Const: "fixed",
		}

		err := schema.VisitJSON("other", opt, openapi3.MultiErrors())
		require.Error(t, err)
	})
}
