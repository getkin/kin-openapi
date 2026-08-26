package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// JSON Schema 2020-12, which OpenAPI 3.1 adopts, allows a boolean wherever a
// schema is expected: `true` accepts every instance and `false` accepts none.
const booleanSchemaSpec = `
openapi: 3.1.0
info:
  title: boolean schemas
  version: "1.0"
paths: {}
components:
  schemas:
    ClosedTuple:
      type: array
      prefixItems:
        - type: string
        - type: integer
      items: false
    OpenPosition:
      type: array
      prefixItems:
        - true
        - type: string
    Anything:
      not: true
    Nothing:
      not: false
    NoExtraProps:
      type: object
      properties:
        name:
          type: string
        anything: true
`

func loadBooleanSchemas(t *testing.T) *openapi3.T {
	t.Helper()

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(booleanSchemaSpec))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(loader.Context))

	return doc
}

// Every position that holds a SchemaRef accepts a boolean, since they all
// decode through Schema.UnmarshalJSON.
func TestBooleanSchema_Positions(t *testing.T) {
	schemas := loadBooleanSchemas(t).Components.Schemas

	for _, test := range []struct {
		name   string
		schema *openapi3.Schema
		want   bool
	}{
		{"items", schemas["ClosedTuple"].Value.Items.Value, false},
		{"prefixItems", schemas["OpenPosition"].Value.PrefixItems[0].Value, true},
		{"not true", schemas["Anything"].Value.Not.Value, true},
		{"not false", schemas["Nothing"].Value.Not.Value, false},
		{"properties", schemas["NoExtraProps"].Value.Properties["anything"].Value, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			require.NotNil(t, test.schema.Boolean, "should have decoded as a boolean schema")
			require.Equal(t, test.want, *test.schema.Boolean)
		})
	}
}

// `true` accepts anything, including null; `false` accepts nothing, including
// null. A boolean schema constrains the instance and nothing else.
func TestBooleanSchema_VisitJSON(t *testing.T) {
	schemas := loadBooleanSchemas(t).Components.Schemas

	trueSchema := schemas["OpenPosition"].Value.PrefixItems[0].Value
	falseSchema := schemas["ClosedTuple"].Value.Items.Value

	for _, value := range []any{"a string", float64(1), true, nil, []any{}, map[string]any{}} {
		require.NoError(t, trueSchema.VisitJSON(value))
		require.Error(t, falseSchema.VisitJSON(value))
	}
}

// An `items: false` schema is reached and rejects, so an array carrying any
// element fails while an empty one passes.
//
// Instance validation applies `items` to every element rather than only to
// those past `prefixItems`, so this does not yet express 2020-12 tuple
// semantics. That is a separate gap in visitJSONArray, which does not consult
// PrefixItems at all.
func TestBooleanSchema_FalseItemsRejectsElements(t *testing.T) {
	closed := loadBooleanSchemas(t).Components.Schemas["ClosedTuple"].Value

	require.NoError(t, closed.VisitJSON([]any{}))
	require.Error(t, closed.VisitJSON([]any{"a"}))
}

// A boolean schema marshals back to the bare boolean it was written as, not to
// the object it is equivalent to.
func TestBooleanSchema_RoundTrip(t *testing.T) {
	const spec = `{"openapi":"3.1.0","info":{"title":"t","version":"1.0"},"paths":{},"components":{"schemas":{"Tuple":{"items":false,"prefixItems":[{"type":"string"}],"type":"array"},"Any":{"not":true}}}}`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	data, err := doc.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, spec, string(data))
}

// A boolean carries no other keyword, so nothing else is populated and the
// schema is empty only when it accepts everything.
func TestBooleanSchema_IsEmpty(t *testing.T) {
	schemas := loadBooleanSchemas(t).Components.Schemas

	require.True(t, schemas["OpenPosition"].Value.PrefixItems[0].Value.IsEmpty(), "true constrains nothing")
	require.False(t, schemas["ClosedTuple"].Value.Items.Value.IsEmpty(), "false rejects everything")
}

// An ordinary schema is unaffected: Boolean stays nil so nothing short-circuits.
func TestBooleanSchema_ObjectSchemaUnaffected(t *testing.T) {
	schemas := loadBooleanSchemas(t).Components.Schemas

	name := schemas["NoExtraProps"].Value.Properties["name"].Value
	require.Nil(t, name.Boolean)
	require.NoError(t, name.VisitJSON("a string"))
	require.Error(t, name.VisitJSON(float64(1)))
}
