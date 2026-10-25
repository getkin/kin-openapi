package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// TestRefSiblingsSurviveMarshal covers what a $ref carries next to it: it has
// to be written back out.
//
// For 3.1+ the loader overlays sibling keywords onto the referenced schema,
// which is what JSON Schema 2020-12 prescribes. Marshaling used to throw the
// siblings away and emit a bare {"$ref": ...}, so a document meant one thing
// before being written out and another after being read back, with nothing to
// report the change.
//
// The generated *Ref types now keep their non-extension siblings verbatim and
// copy them back over the $ref. They do not need to know the document version:
// the version decision was already made at load.
func TestRefSiblingsSurviveMarshal(t *testing.T) {
	t.Parallel()

	t.Run("schema", func(t *testing.T) {
		t.Parallel()

		const spec = `
openapi: 3.1.0
info:
  title: schema sibling demo
  version: "1"
paths: {}
components:
  schemas:
    Base:
      type: string
    Alias:
      $ref: "#/components/schemas/Base"
      description: what Alias means
`

		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromData([]byte(spec))
		require.NoError(t, err)

		// Alias resolves to Base's type with the sibling description layered on
		// top, and Base itself is left alone.
		alias := doc.Components.Schemas["Alias"]
		require.Equal(t, `#/components/schemas/Base`, alias.Ref)
		require.Equal(t, &openapi3.Types{"string"}, alias.Value.Type)
		require.Equal(t, "what Alias means", alias.Value.Description)
		require.Empty(t, doc.Components.Schemas["Base"].Value.Description,
			"the overlay must not leak into the referenced schema")

		once, err := doc.MarshalJSON()
		require.NoError(t, err)
		require.JSONEq(t, `{
			"openapi": "3.1.0",
			"info": {"title": "schema sibling demo", "version": "1"},
			"paths": {},
			"components": {
				"schemas": {
					"Base": {"type": "string"},
					"Alias": {
						"$ref": "#/components/schemas/Base",
						"description": "what Alias means"
					}
				}
			}
		}`, string(once))

		// So the meaning of Alias survives a save/load cycle.
		loader = openapi3.NewLoader()
		reloaded, err := loader.LoadFromData(once)
		require.NoError(t, err)
		require.Equal(t, "what Alias means",
			reloaded.Components.Schemas["Alias"].Value.Description)
	})

	t.Run("a sibling that is not a schema keyword is kept too", func(t *testing.T) {
		t.Parallel()

		// Preserving only recognized keywords would still lose data. This is
		// the shape the fuzzer reduced to; see
		// testdata/fuzz/FuzzLoadFromData/schema-ref-junk-sibling.
		const spec = `
openapi: 3.0.0
info:
  title: junk sibling
  version: "1"
paths: {}
components:
  schemas:
    Base:
      type: string
    Alias:
      $ref: "#/components/schemas/Base"
      nonsense: 1
`

		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromData([]byte(spec))
		require.NoError(t, err)

		// A Reference Object ignores it.
		before := doc.Validate(t.Context())
		require.NoError(t, before)

		once, err := doc.MarshalJSON()
		require.NoError(t, err)
		require.Contains(t, string(once), `"nonsense":1`)

		// The same document validates the same way after a round trip.
		loader = openapi3.NewLoader()
		reloaded, err := loader.LoadFromData(once)
		require.NoError(t, err)
		require.Equal(t, before, reloaded.Validate(t.Context()))
	})
}
