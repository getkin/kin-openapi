package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// prefixItems is positional: the schema at index i validates the item at index
// i, and items governs the positions past the end of the list. prefixItems is
// validated by the JSON Schema 2020-12 validator; the built-in one ignores it.
const prefixItemsSpec = `
openapi: 3.1.0
info:
  title: prefix items
  version: "1.0"
paths: {}
components:
  schemas:
    Tuple:
      type: array
      prefixItems:
        - type: string
        - type: integer
    Tail:
      type: array
      prefixItems:
        - type: string
        - type: integer
      items:
        type: boolean
`

func prefixItemsSchema(t *testing.T, name string) *openapi3.Schema {
	t.Helper()

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(prefixItemsSpec))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(loader.Context))

	return doc.Components.Schemas[name].Value
}

// Each entry validates its own index, so swapping two values that satisfy the
// list in a different order fails.
func TestPrefixItems_Positional(t *testing.T) {
	tuple := prefixItemsSchema(t, "Tuple")
	opt := openapi3.EnableJSONSchema2020()

	require.NoError(t, tuple.VisitJSON([]any{"a", float64(1)}, opt))
	require.Error(t, tuple.VisitJSON([]any{float64(1), "a"}, opt), "the values are in the wrong positions")
	require.Error(t, tuple.VisitJSON([]any{float64(1), float64(1)}, opt), "index 0 must be a string")
	require.Error(t, tuple.VisitJSON([]any{"a", "b"}, opt), "index 1 must be an integer")
}

// The error names the offending index rather than the array as a whole.
func TestPrefixItems_ErrorCarriesTheIndex(t *testing.T) {
	opt := openapi3.EnableJSONSchema2020()
	err := prefixItemsSchema(t, "Tuple").VisitJSON([]any{"a", "b"}, opt)
	require.Error(t, err)
	require.ErrorContains(t, err, `error at "/1"`)
}

// A short array leaves the later entries unapplied; prefixItems constrains the
// positions that exist, and minItems is what requires them to exist.
func TestPrefixItems_ShorterThanTheList(t *testing.T) {
	tuple := prefixItemsSchema(t, "Tuple")
	opt := openapi3.EnableJSONSchema2020()

	require.NoError(t, tuple.VisitJSON([]any{}, opt))
	require.NoError(t, tuple.VisitJSON([]any{"a"}, opt))
	require.Error(t, tuple.VisitJSON([]any{float64(1)}, opt), "index 0 is still checked")
}

// items governs only the positions past prefixItems, so the prefix keeps its
// own types and the tail takes the items schema.
func TestPrefixItems_ItemsAppliesPastThePrefix(t *testing.T) {
	tail := prefixItemsSchema(t, "Tail")
	opt := openapi3.EnableJSONSchema2020()

	require.NoError(t, tail.VisitJSON([]any{"a", float64(1)}, opt))
	require.NoError(t, tail.VisitJSON([]any{"a", float64(1), true, false}, opt))
	require.Error(t, tail.VisitJSON([]any{"a", float64(1), "not a boolean"}, opt), "the tail must match items")
	require.Error(t, tail.VisitJSON([]any{true, float64(1)}, opt), "items must not apply to the prefix")
}

// Without prefixItems, items applies to every position, which is what OAS 3.0
// documents rely on.
func TestPrefixItems_AbsentLeavesItemsUnchanged(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: t
  version: "1.0"
paths: {}
components:
  schemas:
    Strings:
      type: array
      items:
        type: string
`
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	strings := doc.Components.Schemas["Strings"].Value
	require.NoError(t, strings.VisitJSON([]any{"a", "b"}))
	require.Error(t, strings.VisitJSON([]any{"a", float64(1)}))
}

// Every failing position is reported when multiple errors are requested.
func TestPrefixItems_MultiError(t *testing.T) {
	opt := openapi3.EnableJSONSchema2020()
	err := prefixItemsSchema(t, "Tuple").VisitJSON([]any{float64(1), "a"}, opt, openapi3.MultiErrors())
	require.Error(t, err)

	var multi openapi3.MultiError
	require.ErrorAs(t, err, &multi)
	require.Len(t, multi, 2, "both positions are wrong")
}
