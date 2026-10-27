package openapi3_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// Members are checked in map order, which Go varies from run to run, but
// their errors must still come out in order of member name.
func TestObjectMemberErrorsInNameOrder(t *testing.T) {
	t.Parallel()

	schema := openapi3.NewObjectSchema()
	value := map[string]any{}
	var want []string
	for i := range 20 {
		name := fmt.Sprintf("p%02d", i)
		schema.WithProperty(name, openapi3.NewIntegerSchema())
		value[name] = "not an integer"
		want = append(want, name)
	}

	for range 50 {
		err := schema.VisitJSON(value, openapi3.MultiErrors())
		var me openapi3.MultiError
		require.ErrorAs(t, err, &me)
		var got []string
		for _, e := range me {
			var se *openapi3.SchemaError
			require.ErrorAs(t, e, &se)
			got = append(got, se.JSONPointer()[0])
		}
		require.Equal(t, want, got)

		// Without MultiErrors the error is the first member's in name order.
		err = schema.VisitJSON(value)
		var se *openapi3.SchemaError
		require.ErrorAs(t, err, &se)
		require.Equal(t, []string{"p00"}, se.JSONPointer())
	}
}

// An unsupported member and a member that does not fit its property are
// ordered together by name.
func TestObjectUnsupportedMemberErrorsInNameOrder(t *testing.T) {
	t.Parallel()

	schema := openapi3.NewObjectSchema().
		WithProperty("b", openapi3.NewIntegerSchema()).
		WithoutAdditionalProperties()
	value := map[string]any{"a": 1, "b": "x", "c": 1}

	for range 50 {
		err := schema.VisitJSON(value, openapi3.MultiErrors())
		var me openapi3.MultiError
		require.ErrorAs(t, err, &me)
		require.Len(t, me, 3)
		require.ErrorContains(t, me[0], `"a"`)
		require.ErrorContains(t, me[1], `"/b"`)
		require.ErrorContains(t, me[2], `"c"`)
	}
}

// An unresolved reference is reported rather than the errors of members
// that come before it, whichever order the members are checked in.
func TestObjectUnresolvedMemberErrorInNameOrder(t *testing.T) {
	t.Parallel()

	schema := openapi3.NewObjectSchema().
		WithProperty("a", openapi3.NewIntegerSchema()).
		WithProperty("z", openapi3.NewIntegerSchema())
	schema.Properties["m"] = &openapi3.SchemaRef{Ref: "#/components/schemas/Missing"}
	value := map[string]any{"a": "x", "m": 1, "z": "x"}

	for range 50 {
		for _, opts := range [][]openapi3.SchemaValidationOption{nil, {openapi3.MultiErrors()}} {
			err := schema.VisitJSON(value, opts...)
			var ur *openapi3.UnresolvedRefError
			if len(opts) > 0 {
				require.ErrorAs(t, err, &ur)
			} else {
				// "a" comes before "m", so its error is the one returned.
				var se *openapi3.SchemaError
				require.ErrorAs(t, err, &se)
				require.Equal(t, []string{"a"}, se.JSONPointer())
			}
		}
	}
}

func BenchmarkVisitJSONObject(b *testing.B) {
	item := openapi3.NewObjectSchema()
	for i := range 12 {
		item.WithProperty(fmt.Sprintf("field%02d", i), openapi3.NewStringSchema())
	}
	schema := openapi3.NewObjectSchema().WithProperty("items", openapi3.NewArraySchema().WithItems(item))
	var items []any
	for range 50 {
		o := map[string]any{}
		for i := range 12 {
			o[fmt.Sprintf("field%02d", i)] = "value"
		}
		items = append(items, o)
	}
	value := map[string]any{"items": items}

	b.ReportAllocs()
	for b.Loop() {
		if err := schema.VisitJSON(value, openapi3.VisitAsRequest(), openapi3.MultiErrors()); err != nil {
			b.Fatal(err)
		}
	}
}
