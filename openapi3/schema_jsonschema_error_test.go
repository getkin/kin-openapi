package openapi3_test

import (
	"errors"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func jsonSchemaErrorSubject() *openapi3.Schema {
	return &openapi3.Schema{
		Type: &openapi3.Types{"object"},
		Properties: openapi3.Schemas{
			"n":    openapi3.NewSchemaRef("", &openapi3.Schema{Type: &openapi3.Types{"integer"}}),
			"tags": openapi3.NewSchemaRef("", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema().WithMaxLength(2))),
		},
	}
}

// A SchemaError of the JSON Schema 2020-12 validator carries the same structure
// as one of the built-in validator: the path of the failing value, the failing
// keyword and the failing value. Callers tell a malformed body from a broken
// field by that structure, and it must not depend on the validator in use.
func TestJSONSchema2020_SchemaErrorCarriesItsStructure(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		pointer []string
		field   string
		failing any
	}{
		{"a field of the wrong type", map[string]any{"n": "x"}, []string{"n"}, "type", "x"},
		{"an item breaking its bound", map[string]any{"tags": []any{"ok", "long"}}, []string{"tags", "1"}, "maxLength", "long"},
		{"the root of the wrong type", []any{}, nil, "type", []any{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for mode, opts := range map[string][]openapi3.SchemaValidationOption{
				"built-in":         {openapi3.MultiErrors()},
				"JSON Schema 2020": {openapi3.EnableJSONSchema2020()},
			} {
				err := jsonSchemaErrorSubject().VisitJSON(tt.value, opts...)
				require.Error(t, err, mode)

				var schemaErr *openapi3.SchemaError
				require.True(t, errors.As(err, &schemaErr), mode)
				require.Equal(t, tt.pointer, nilIfEmpty(schemaErr.JSONPointer()), mode)
				require.Equal(t, tt.field, schemaErr.SchemaField, mode)
				require.Equal(t, tt.failing, schemaErr.Value, mode)
			}
		})
	}
}

// The error of the JSON Schema 2020-12 validator stays reachable, so a caller
// can read its typed kind.
func TestJSONSchema2020_SchemaErrorUnwrapsToTheValidatorError(t *testing.T) {
	err := jsonSchemaErrorSubject().VisitJSON(map[string]any{"n": "x"}, openapi3.EnableJSONSchema2020())
	require.Error(t, err)

	var validationErr *jsonschema.ValidationError
	require.True(t, errors.As(err, &validationErr))
	require.Equal(t, []string{"n"}, validationErr.InstanceLocation)

	mistyped, ok := validationErr.ErrorKind.(*kind.Type)
	require.True(t, ok, "kind %T", validationErr.ErrorKind)
	require.Equal(t, "string", mistyped.Got)
}

// The message of a converted error is the one it had before it carried its
// structure: the path is named once.
func TestJSONSchema2020_SchemaErrorMessageNamesThePathOnce(t *testing.T) {
	err := jsonSchemaErrorSubject().VisitJSON(map[string]any{"n": "x"}, openapi3.EnableJSONSchema2020())
	require.Error(t, err)
	require.ErrorContains(t, err, `error at "/n"`)
	message := err.Error()
	require.NotContains(t, message, `Error at "/n"`)
}

func nilIfEmpty(path []string) []string {
	if len(path) == 0 {
		return nil
	}
	return path
}
