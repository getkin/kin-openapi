package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// The empty schema matches every instance, so `not: {}` matches none.
func TestNot_EmptySubschemaRejectsEverything(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: t
  version: "1.0"
paths: {}
components:
  schemas:
    Nothing:
      not: {}
    NothingNested:
      properties:
        a:
          not: {}
`
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(loader.Context))

	nothing := doc.Components.Schemas["Nothing"].Value
	require.False(t, nothing.IsEmpty(), "a not constrains the schema whatever it holds")

	// null included: `{}` places no type constraint, so it matches every
	// instance and `not: {}` matches none of them.
	for _, value := range []any{"a string", float64(1), true, nil, []any{}, map[string]any{}} {
		require.Error(t, nothing.VisitJSON(value))
	}

	// The same through a property, not only on a root schema.
	nested := doc.Components.Schemas["NothingNested"].Value
	require.Error(t, nested.VisitJSON(map[string]any{"a": "anything"}))
	require.NoError(t, nested.VisitJSON(map[string]any{}))
}

// 3.1 reaches the same result. `{}` is a JSON Schema 2020-12 schema there
// rather than an OAS Schema Object, but an empty one is unconstrained either
// way, so there is nothing version dependent to gate.
func TestNot_EmptySubschemaRejectsEverythingIn31(t *testing.T) {
	const spec = `
openapi: 3.1.0
info:
  title: t
  version: "1.0"
paths: {}
components:
  schemas:
    Nothing:
      not: {}
`
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(loader.Context))

	nothing := doc.Components.Schemas["Nothing"].Value
	for _, value := range []any{"a string", float64(1), true, nil, []any{}, map[string]any{}} {
		require.Error(t, nothing.VisitJSON(value))
	}
}

// A `not` holding a schema rejects exactly the values that schema matches.
func TestNot_NonEmptySubschemaUnchanged(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: t
  version: "1.0"
paths: {}
components:
  schemas:
    NotAString:
      not:
        type: string
`
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	s := doc.Components.Schemas["NotAString"].Value
	require.Error(t, s.VisitJSON("a string"))
	require.NoError(t, s.VisitJSON(float64(1)))
}
