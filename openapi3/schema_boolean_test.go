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
      not: false
    Nothing:
      not: true
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
		{"not false", schemas["Anything"].Value.Not.Value, false},
		{"not true", schemas["Nothing"].Value.Not.Value, true},
		{"properties", schemas["NoExtraProps"].Value.Properties["anything"].Value, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			require.NotNil(t, test.schema.Always, "should have decoded as a boolean schema")
			require.Equal(t, test.want, *test.schema.Always)
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

// `items: false` closes a tuple at the prefixItems length: an array that is a
// prefix of the tuple is valid, and any element past the prefix is rejected,
// since no prefixItems entry covers that position and items forbids it.
func TestBooleanSchema_FalseItemsClosesTuple(t *testing.T) {
	closed := loadBooleanSchemas(t).Components.Schemas["ClosedTuple"].Value

	require.NoError(t, closed.VisitJSON([]any{}))
	require.NoError(t, closed.VisitJSON([]any{"a"}))
	require.NoError(t, closed.VisitJSON([]any{"a", float64(1)}))

	require.Error(t, closed.VisitJSON([]any{"a", float64(1), "extra"}), "items: false forbids a third element")
	require.Error(t, closed.VisitJSON([]any{float64(1), "a"}), "prefixItems is positional")
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

// Neither boolean is empty. `false` rejects every instance, and `true` is
// accepted by visitJSON explicitly rather than by skipping validation, which is
// what leaves `not: true` free to reject.
func TestBooleanSchema_IsEmpty(t *testing.T) {
	schemas := loadBooleanSchemas(t).Components.Schemas

	require.False(t, schemas["OpenPosition"].Value.PrefixItems[0].Value.IsEmpty())
	require.False(t, schemas["ClosedTuple"].Value.Items.Value.IsEmpty())
}

// An ordinary schema is unaffected: Boolean stays nil so nothing short-circuits.
func TestBooleanSchema_ObjectSchemaUnaffected(t *testing.T) {
	schemas := loadBooleanSchemas(t).Components.Schemas

	name := schemas["NoExtraProps"].Value.Properties["name"].Value
	require.Nil(t, name.Always)
	require.NoError(t, name.VisitJSON("a string"))
	require.Error(t, name.VisitJSON(float64(1)))
}

// `not` inverts the boolean, so the composition is what matters and not just
// the decoded value: `not: true` matches nothing and `not: false` matches
// everything.
func TestBooleanSchema_UnderNot(t *testing.T) {
	schemas := loadBooleanSchemas(t).Components.Schemas

	anything := schemas["Anything"].Value
	nothing := schemas["Nothing"].Value

	// Null is left out: whether a schema admits it is decided by kin's
	// nullability rules before `not` is reached, and a boolean does not change
	// that either way.
	for _, value := range []any{"a string", float64(1), []any{}, map[string]any{}} {
		require.NoError(t, anything.VisitJSON(value), "not: false matches everything")
		require.Error(t, nothing.VisitJSON(value), "not: true matches nothing")
	}
}

// A Schema built in code can carry a boolean alongside other keywords. It would
// marshal back as the bare boolean and drop them, so Validate rejects it rather
// than let the next serialization lose content.
func TestBooleanSchema_RejectsOtherKeywords(t *testing.T) {
	boolean := true

	// A boolean schema is 3.1, so say so; without it the version gate fires
	// first, as it does for every other 2020-12 construct.
	as31 := openapi3.IsOpenAPI31OrLater()

	require.NoError(t, (&openapi3.Schema{Always: &boolean}).Validate(t.Context(), as31))

	for name, schema := range map[string]*openapi3.Schema{
		"type":       {Always: &boolean, Type: &openapi3.Types{"string"}},
		"properties": {Always: &boolean, Properties: openapi3.Schemas{"a": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}}},
		"extensions": {Always: &boolean, Extensions: map[string]any{"x-a": 1}},
	} {
		t.Run(name, func(t *testing.T) {
			err := schema.Validate(t.Context(), as31)
			require.Error(t, err)
			var exclusive *openapi3.MutuallyExclusiveFieldsError
			require.ErrorAs(t, err, &exclusive)
		})
	}
}

// VisitJSON documents EnableJSONSchema2020 as the way to validate a 3.1
// schema, and a boolean schema exists only in 3.1, so the option's path has to
// agree with the built-in one.
//
// It is not purely external: useJSONSchema2020 falls back to visitJSON when the
// schema does not compile, which is what a bare boolean at the root does, so
// both paths are exercised here.
func TestBooleanSchema_JSONSchema2020(t *testing.T) {
	schemas := loadBooleanSchemas(t).Components.Schemas
	opt := openapi3.EnableJSONSchema2020()

	trueSchema := schemas["OpenPosition"].Value.PrefixItems[0].Value
	falseSchema := schemas["ClosedTuple"].Value.Items.Value

	for _, value := range []any{"a string", float64(1), []any{}, map[string]any{}} {
		require.NoError(t, trueSchema.VisitJSON(value, opt))
		require.Error(t, falseSchema.VisitJSON(value, opt))
	}

	require.NoError(t, schemas["Anything"].Value.VisitJSON(float64(1), opt), "not: false matches everything")
	require.Error(t, schemas["Nothing"].Value.VisitJSON(float64(1), opt), "not: true matches nothing")
}

// The 2020-12 validator has to agree with the built-in one about a closed
// tuple, since a document does not say which validator will read it.
func TestBooleanSchema_ClosesTupleUnder2020(t *testing.T) {
	closed := loadBooleanSchemas(t).Components.Schemas["ClosedTuple"].Value
	opt := openapi3.EnableJSONSchema2020()

	require.NoError(t, closed.VisitJSON([]any{"a", float64(1)}, opt))
	require.Error(t, closed.VisitJSON([]any{"a", float64(1), "extra"}, opt),
		"items: false must reject a position past prefixItems")
	require.Error(t, closed.VisitJSON([]any{float64(1), "a"}, opt),
		"prefixItems is positional")
}

// A boolean schema is JSON Schema 2020-12, so it does not exist in 3.0, where a
// schema MUST be a Schema Object. Validate says so rather than accepting it.
func TestBooleanSchema_RejectedBefore31(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: t
  version: "1.0"
paths: {}
components:
  schemas:
    Closed:
      type: array
      items: false
`
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err, "it still parses; the version is a validation matter")

	err = doc.Validate(loader.Context)
	require.Error(t, err)

	var mismatch *openapi3.FieldVersionMismatchError
	require.ErrorAs(t, err, &mismatch)
	require.Equal(t, "3.1", mismatch.MinVersion)
}
