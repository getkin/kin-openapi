package openapi3_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestSchemaEnumJSONNumbers(t *testing.T) {
	for _, tc := range []struct {
		name  string
		enum  []any
		value any
	}{
		{"json.Number in enum", []any{json.Number("1")}, float64(1)},
		{"json.Number in enum and value", []any{json.Number("1.0")}, json.Number("1")},
		{"int in enum", []any{1}, json.Number("1")},
		{"json.Number in value object", []any{map[string]any{"a": float64(1)}}, map[string]any{"a": json.Number("1")}},
		{"json.Number in enum object", []any{map[string]any{"a": json.Number("1")}}, map[string]any{"a": float64(1)}},
		{"json.Number in value array", []any{[]any{float64(1), "b"}}, []any{json.Number("1"), "b"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, enum := range [][]any{tc.enum, append(tc.enum, "other")} {
				schema := &openapi3.Schema{Enum: enum}
				require.NoError(t, schema.VisitJSON(tc.value))
				require.NoError(t, schema.VisitJSON(tc.value, openapi3.EnableJSONSchema2020()))
			}
		})
	}

	for _, tc := range []struct {
		name  string
		enum  []any
		value any
	}{
		{"different numbers", []any{json.Number("1")}, float64(2)},
		{"number and string", []any{json.Number("1")}, "1"},
		{"different nested numbers", []any{map[string]any{"a": json.Number("1")}}, map[string]any{"a": float64(2)}},
		{"extra member", []any{map[string]any{"a": json.Number("1")}}, map[string]any{"a": float64(1), "b": nil}},
		{"longer array", []any{[]any{json.Number("1")}}, []any{float64(1), float64(1)}},
		{"invalid json.Number", []any{json.Number("1")}, json.Number("one")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			schema := &openapi3.Schema{Enum: tc.enum}
			err := schema.VisitJSON(tc.value)
			var schemaErr *openapi3.SchemaError
			require.ErrorAs(t, err, &schemaErr)
			require.Equal(t, "enum", schemaErr.SchemaField)

			if tc.name != "invalid json.Number" {
				require.ErrorAs(t, schema.VisitJSON(tc.value, openapi3.EnableJSONSchema2020()), &schemaErr)
			}
		})
	}
}
