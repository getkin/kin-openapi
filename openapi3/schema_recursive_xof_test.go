package openapi3_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// A tree whose node is one of two shapes, each holding a child node: both
// subschemas of anyOf or oneOf lead to the same child value.
const recursiveXOfSpec = `
openapi: 3.0.3
info: {title: t, version: "1"}
paths: {}
components:
  schemas:
    AnyOf:
      anyOf:
        - {type: object, properties: {c: {$ref: '#/components/schemas/AnyOf'}}, required: [z]}
        - {type: object, properties: {c: {$ref: '#/components/schemas/AnyOf'}}, required: [y]}
    OneOf:
      oneOf:
        - {type: object, properties: {c: {$ref: '#/components/schemas/OneOf'}}, required: [z]}
        - {type: object, properties: {c: {$ref: '#/components/schemas/OneOf'}}, required: [y]}
`

func recursiveXOfSchemas(tb testing.TB) (anyOf, oneOf *openapi3.Schema) {
	tb.Helper()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(recursiveXOfSpec))
	require.NoError(tb, err)
	require.NoError(tb, doc.Validate(context.Background()))
	return doc.Components.Schemas["AnyOf"].Value, doc.Components.Schemas["OneOf"].Value
}

// nestedValue is depth objects, each wrapping the next under "c".
func nestedValue(tb testing.TB, wrap, leaf string, depth int) any {
	tb.Helper()
	var v any
	err := json.Unmarshal([]byte(strings.Repeat(wrap, depth)+leaf+strings.Repeat("}", depth)), &v)
	require.NoError(tb, err)
	return v
}

// Each node fits only the last subschema, so a check that visited every
// subschema afresh would take time exponential in the depth.
func TestVisitJSONRecursiveXOfDeepValue(t *testing.T) {
	t.Parallel()

	anyOf, oneOf := recursiveXOfSchemas(t)
	valid := nestedValue(t, `{"y":1,"c":`, `{"y":1}`, 60)
	invalid := nestedValue(t, `{"y":1,"c":`, `{}`, 60)
	for _, opts := range [][]openapi3.SchemaValidationOption{
		nil,
		{openapi3.MultiErrors()},
		{openapi3.FailFast()},
		{openapi3.VisitAsRequest(), openapi3.MultiErrors()},
		{openapi3.VisitAsResponse()},
	} {
		require.NoError(t, anyOf.VisitJSON(valid, opts...))
		require.NoError(t, oneOf.VisitJSON(valid, opts...))
		require.Error(t, anyOf.VisitJSON(invalid, opts...))
	}
	// Every subschema's errors are part of a oneOf error, so only a check
	// that stops at the first error decides an invalid value this deep.
	require.Error(t, oneOf.VisitJSON(invalid, openapi3.FailFast()))
}

// A value that occurs twice in the input is reported at each place it
// does not fit.
func TestVisitJSONRepeatedValueErrors(t *testing.T) {
	t.Parallel()

	item := openapi3.NewObjectSchema().WithProperty("n", openapi3.NewIntegerSchema())
	schema := openapi3.NewAllOfSchema(openapi3.NewObjectSchema().
		WithProperty("a", item).
		WithProperty("b", item))
	shared := map[string]any{"n": "x"}
	err := schema.VisitJSON(map[string]any{"a": shared, "b": shared}, openapi3.MultiErrors())
	require.ErrorContains(t, err, `Error at "/a/n"`)
	require.ErrorContains(t, err, `Error at "/b/n"`)
}

func BenchmarkVisitJSONRecursiveAnyOf(b *testing.B) {
	anyOf, _ := recursiveXOfSchemas(b)
	value := nestedValue(b, `{"y":1,"c":`, `{"y":1}`, 10)

	b.ReportAllocs()
	for b.Loop() {
		if err := anyOf.VisitJSON(value, openapi3.VisitAsRequest(), openapi3.MultiErrors()); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVisitJSONRecursiveOneOf(b *testing.B) {
	_, oneOf := recursiveXOfSchemas(b)
	value := nestedValue(b, `{"y":1,"c":`, `{"y":1}`, 10)

	b.ReportAllocs()
	for b.Loop() {
		if err := oneOf.VisitJSON(value, openapi3.VisitAsRequest(), openapi3.MultiErrors()); err != nil {
			b.Fatal(err)
		}
	}
}
