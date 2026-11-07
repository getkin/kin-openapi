package openapi3_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestOrderedObjectErrorsInMemberOrder(t *testing.T) {
	t.Parallel()

	schema := openapi3.NewObjectSchema().
		WithProperty("a", openapi3.NewIntegerSchema()).
		WithProperty("b", openapi3.NewIntegerSchema()).
		WithProperty("c", openapi3.NewIntegerSchema())
	value := openapi3.NewOrderedObject()
	value.Set("c", "x")
	value.Set("a", "x")
	value.Set("b", "x")

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
		require.Equal(t, []string{"c", "a", "b"}, got)
	}
}

func TestOrderedObjectValidation(t *testing.T) {
	t.Parallel()

	obj := func(kv ...any) *openapi3.OrderedObject {
		o := openapi3.NewOrderedObject()
		for i := 0; i < len(kv); i += 2 {
			o.Set(kv[i].(string), kv[i+1])
		}
		return o
	}

	t.Run("valid", func(t *testing.T) {
		schema := openapi3.NewObjectSchema().
			WithProperty("a", openapi3.NewIntegerSchema()).
			WithRequired([]string{"a"})
		require.NoError(t, schema.VisitJSON(obj("a", 1)))
		require.Error(t, schema.VisitJSON(obj("b", 1)))
	})

	t.Run("enum ignores member order", func(t *testing.T) {
		schema := openapi3.NewObjectSchema()
		schema.Enum = []any{map[string]any{"a": 1, "b": 2}}
		require.NoError(t, schema.VisitJSON(obj("b", 2, "a", 1)))
		require.Error(t, schema.VisitJSON(obj("a", 1)))
	})

	t.Run("unique items ignore member order", func(t *testing.T) {
		schema := openapi3.NewArraySchema().WithItems(openapi3.NewObjectSchema())
		schema.UniqueItems = true
		require.Error(t, schema.VisitJSON([]any{obj("a", 1, "b", 2), obj("b", 2, "a", 1)}))
		require.NoError(t, schema.VisitJSON([]any{obj("a", 1), obj("a", 2)}))
	})

	t.Run("default is added after the members", func(t *testing.T) {
		schema := openapi3.NewObjectSchema().
			WithProperty("a", openapi3.NewStringSchema()).
			WithProperty("b", openapi3.NewStringSchema().WithDefault("d"))
		value := obj("a", "x")
		require.NoError(t, schema.VisitJSON(value, openapi3.VisitAsRequest(), openapi3.DefaultsSet(func() {})))
		require.Equal(t, []string{"a", "b"}, value.Names())
		b, ok := value.Get("b")
		require.True(t, ok)
		require.Equal(t, "d", b)
	})

	t.Run("default for a null member keeps its place", func(t *testing.T) {
		schema := openapi3.NewObjectSchema().
			WithProperty("a", openapi3.NewStringSchema().WithDefault("d")).
			WithProperty("b", openapi3.NewStringSchema())
		value := obj("a", nil, "b", "x")
		require.NoError(t, schema.VisitJSON(value, openapi3.VisitAsRequest(), openapi3.DefaultsSet(func() {})))
		require.Equal(t, []string{"a", "b"}, value.Names())
	})

	t.Run("JSON Schema 2020-12", func(t *testing.T) {
		schema := openapi3.NewObjectSchema().
			WithProperty("a", openapi3.NewIntegerSchema()).
			WithRequired([]string{"a"})
		require.NoError(t, schema.VisitJSON(obj("a", 1), openapi3.EnableJSONSchema2020()))
		items := openapi3.NewArraySchema().WithItems(schema)
		require.NoError(t, items.VisitJSON([]any{obj("a", 1)}, openapi3.EnableJSONSchema2020()))
		require.Error(t, items.VisitJSON([]any{obj("a", "x")}, openapi3.EnableJSONSchema2020()))
		require.Error(t, schema.VisitJSON(obj("a", "x"), openapi3.EnableJSONSchema2020()))
		require.Error(t, schema.VisitJSON(obj("b", 1), openapi3.EnableJSONSchema2020()))
	})

	t.Run("zero and nil values", func(t *testing.T) {
		var o openapi3.OrderedObject
		o.Set("a", 1)
		require.Equal(t, []string{"a"}, o.Names())

		var nilObject *openapi3.OrderedObject
		require.NoError(t, openapi3.NewObjectSchema().VisitJSON(nilObject))
		require.Error(t, openapi3.NewStringSchema().VisitJSON(nilObject))
	})

	t.Run("marshals in order", func(t *testing.T) {
		b, err := obj("b", 1, "a", obj("z", true, "y", nil)).MarshalJSON()
		require.NoError(t, err)
		require.Equal(t, `{"b":1,"a":{"z":true,"y":null}}`, string(b))
	})

	t.Run("errors in the unsupported member order", func(t *testing.T) {
		schema := openapi3.NewObjectSchema().WithoutAdditionalProperties()
		err := schema.VisitJSON(obj("z", 1, "a", 1), openapi3.MultiErrors())
		var me openapi3.MultiError
		require.ErrorAs(t, err, &me)
		require.ErrorContains(t, me[0], fmt.Sprintf("%q", "z"))
		require.ErrorContains(t, me[1], fmt.Sprintf("%q", "a"))
	})
}
