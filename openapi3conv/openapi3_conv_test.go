package openapi3conv_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3conv"
)

func loadV30(t *testing.T, raw string) *openapi3.T {
	t.Helper()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(raw))
	require.NoError(t, err)
	return doc
}

// Round-trips through JSON so the result reflects what tooling consumers see
// (omitted zero-value fields, normalized ordering). Easier to compare against
// expected output than walking structs.
func marshalJSON(t *testing.T, doc *openapi3.T) map[string]any {
	t.Helper()
	b, err := json.Marshal(doc)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	return m
}

// ---------------------------------------------------------------------------
// Version bump
// ---------------------------------------------------------------------------

func TestUpgrade_BumpsVersion(t *testing.T) {
	doc := loadV30(t, `
openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
`)
	openapi3conv.Upgrade(doc)
	assert.Equal(t, "3.2.0", doc.OpenAPI)
}

// 3.2 is purely additive over 3.1 — no schema-level rewrites between them.
// Upgrade applies the 3.0 → 3.1 rewrites (still required) and writes the
// 3.2 version string. The result is canonical and version-correct.
func TestUpgrade_RewritesAppliedAlongsideVersionBump(t *testing.T) {
	doc := loadV30(t, `
openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
components:
  schemas:
    Pet:
      type: string
      nullable: true
`)
	openapi3conv.Upgrade(doc)
	assert.Equal(t, "3.2.0", doc.OpenAPI)
	assert.Equal(t, openapi3.Types{"string", "null"}, *doc.Components.Schemas["Pet"].Value.Type)
}

// ---------------------------------------------------------------------------
// Nullable rewrite
// ---------------------------------------------------------------------------

func TestUpgrade_NullableWithType(t *testing.T) {
	doc := loadV30(t, `
openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
components:
  schemas:
    Pet:
      type: string
      nullable: true
`)
	openapi3conv.Upgrade(doc)
	pet := doc.Components.Schemas["Pet"].Value
	require.NotNil(t, pet.Type)
	assert.Equal(t, openapi3.Types{"string", "null"}, *pet.Type)
	assert.False(t, pet.Nullable, "nullable should be cleared after rewrite")
}

func TestUpgrade_NullableAlreadyHasNullInTypeArray(t *testing.T) {
	doc := loadV30(t, `
openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
components:
  schemas:
    Pet:
      type: ['string', 'null']
      nullable: true
`)
	openapi3conv.Upgrade(doc)
	pet := doc.Components.Schemas["Pet"].Value
	assert.Equal(t, openapi3.Types{"string", "null"}, *pet.Type, "no duplicate null appended")
	assert.False(t, pet.Nullable)
}

func TestUpgrade_NullableNoType(t *testing.T) {
	// nullable without an accompanying type is ambiguous in 3.0 (the spec is
	// silent on whether nullable applies). Drop nullable; the schema then
	// accepts any type, which subsumes null.
	doc := loadV30(t, `
openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
components:
  schemas:
    Pet:
      nullable: true
`)
	openapi3conv.Upgrade(doc)
	pet := doc.Components.Schemas["Pet"].Value
	assert.False(t, pet.Nullable)
	assert.Nil(t, pet.Type)
}

func TestUpgrade_NullableInsideProperties(t *testing.T) {
	doc := loadV30(t, `
openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
components:
  schemas:
    Pet:
      type: object
      properties:
        name:
          type: string
          nullable: true
        ageInYears:
          type: integer
          nullable: true
`)
	openapi3conv.Upgrade(doc)
	props := doc.Components.Schemas["Pet"].Value.Properties
	assert.Equal(t, openapi3.Types{"string", "null"}, *props["name"].Value.Type)
	assert.Equal(t, openapi3.Types{"integer", "null"}, *props["ageInYears"].Value.Type)
}

// ---------------------------------------------------------------------------
// Exclusive-bound rewrite
// ---------------------------------------------------------------------------

func TestUpgrade_ExclusiveMinTrueWithMinimum(t *testing.T) {
	doc := loadV30(t, `
openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
components:
  schemas:
    Score:
      type: integer
      minimum: 5
      exclusiveMinimum: true
`)
	openapi3conv.Upgrade(doc)
	score := doc.Components.Schemas["Score"].Value
	assert.Nil(t, score.Min, "Min cleared")
	require.NotNil(t, score.ExclusiveMin.Value)
	assert.Equal(t, 5.0, *score.ExclusiveMin.Value)
	assert.Nil(t, score.ExclusiveMin.Bool)
}

func TestUpgrade_ExclusiveMaxTrueWithMaximum(t *testing.T) {
	doc := loadV30(t, `
openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
components:
  schemas:
    Score:
      type: integer
      maximum: 100
      exclusiveMaximum: true
`)
	openapi3conv.Upgrade(doc)
	score := doc.Components.Schemas["Score"].Value
	assert.Nil(t, score.Max)
	require.NotNil(t, score.ExclusiveMax.Value)
	assert.Equal(t, 100.0, *score.ExclusiveMax.Value)
	assert.Nil(t, score.ExclusiveMax.Bool)
}

func TestUpgrade_ExclusiveMinFalseDropped(t *testing.T) {
	// `exclusiveMinimum: false` is the default — it carries no information.
	// The merged 3.1 form drops it.
	doc := loadV30(t, `
openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
components:
  schemas:
    Score:
      type: integer
      minimum: 5
      exclusiveMinimum: false
`)
	openapi3conv.Upgrade(doc)
	score := doc.Components.Schemas["Score"].Value
	require.NotNil(t, score.Min)
	assert.Equal(t, 5.0, *score.Min)
	assert.False(t, score.ExclusiveMin.IsSet(), "exclusiveMinimum: false should be dropped")
}

func TestUpgrade_ExclusiveBoundsLeavesNumericIntact(t *testing.T) {
	// A document already in 3.1 numeric form should round-trip unchanged
	// (apart from the version bump to 3.2.0).
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.1.1
info: {title: t, version: '1'}
paths: {}
components:
  schemas:
    Score:
      type: integer
      exclusiveMinimum: 5
`))
	require.NoError(t, err)

	openapi3conv.Upgrade(doc)
	score := doc.Components.Schemas["Score"].Value
	require.NotNil(t, score.ExclusiveMin.Value)
	assert.Equal(t, 5.0, *score.ExclusiveMin.Value)
	assert.Nil(t, score.Min)
}

// ---------------------------------------------------------------------------
// Example -> Examples rewrite
// ---------------------------------------------------------------------------

func TestUpgrade_ExampleToExamples(t *testing.T) {
	doc := loadV30(t, `
openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
components:
  schemas:
    Pet:
      type: string
      example: fido
`)
	openapi3conv.Upgrade(doc)
	pet := doc.Components.Schemas["Pet"].Value
	assert.Nil(t, pet.Example)
	require.Len(t, pet.Examples, 1)
	assert.Equal(t, "fido", pet.Examples[0])
}

func TestUpgrade_ExampleAppendsToExistingExamples(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
components:
  schemas:
    Pet:
      type: string
      example: fido
      examples: [rex]
`))
	require.NoError(t, err)
	openapi3conv.Upgrade(doc)
	pet := doc.Components.Schemas["Pet"].Value
	assert.Nil(t, pet.Example)
	assert.Equal(t, []any{"rex", "fido"}, pet.Examples)
}

// ---------------------------------------------------------------------------
// Idempotence
// ---------------------------------------------------------------------------

func TestUpgrade_Idempotent(t *testing.T) {
	doc := loadV30(t, `
openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
components:
  schemas:
    Pet:
      type: object
      properties:
        name:
          type: string
          nullable: true
          example: fido
        score:
          type: integer
          minimum: 0
          exclusiveMinimum: true
`)
	openapi3conv.Upgrade(doc)
	first := marshalJSON(t, doc)

	openapi3conv.Upgrade(doc)
	second := marshalJSON(t, doc)

	assert.Equal(t, first, second, "second pass must be a no-op")
}

// ---------------------------------------------------------------------------
// Walks reachable from operations and request/response bodies
// ---------------------------------------------------------------------------

func TestUpgrade_WalksOperationSchemas(t *testing.T) {
	doc := loadV30(t, `
openapi: 3.0.3
info: {title: t, version: '1'}
paths:
  /pets:
    get:
      parameters:
        - name: id
          in: query
          schema:
            type: string
            nullable: true
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: object
                properties:
                  total:
                    type: integer
                    minimum: 0
                    exclusiveMinimum: true
`)
	openapi3conv.Upgrade(doc)

	getOp := doc.Paths.Value("/pets").Get
	param := getOp.Parameters[0].Value.Schema.Value
	assert.Equal(t, openapi3.Types{"string", "null"}, *param.Type)

	body := getOp.Responses.Value("200").Value.Content["application/json"].Schema.Value
	total := body.Properties["total"].Value
	assert.Nil(t, total.Min)
	require.NotNil(t, total.ExclusiveMin.Value)
	assert.Equal(t, 0.0, *total.ExclusiveMin.Value)
}

// ---------------------------------------------------------------------------
// Cycle safety
// ---------------------------------------------------------------------------

func TestUpgrade_CycleSafe(t *testing.T) {
	// Hand-build a cycle without going through YAML — the loader resolves
	// $ref into shared *Schema pointers, so a self-referential schema
	// becomes a true graph cycle. The walker must terminate.
	cycle := &openapi3.Schema{Type: &openapi3.Types{"object"}}
	cycle.Properties = openapi3.Schemas{
		"self": &openapi3.SchemaRef{Value: cycle},
		"name": &openapi3.SchemaRef{Value: &openapi3.Schema{
			Type:     &openapi3.Types{"string"},
			Nullable: true,
		}},
	}
	doc := &openapi3.T{
		OpenAPI: "3.0.3",
		Info:    &openapi3.Info{Title: "t", Version: "1"},
		Paths:   openapi3.NewPaths(),
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{"Cycle": &openapi3.SchemaRef{Value: cycle}},
		},
	}

	done := make(chan struct{})
	go func() { openapi3conv.Upgrade(doc); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Upgrade hung on a cyclic schema")
	}

	// Sanity: the rewrite still ran on the non-cyclic property.
	name := cycle.Properties["name"].Value
	assert.Equal(t, openapi3.Types{"string", "null"}, *name.Type)
}

// ---------------------------------------------------------------------------
// Verbose logging
// ---------------------------------------------------------------------------

func TestUpgrade_VerboseLogsRewrites(t *testing.T) {
	doc := loadV30(t, `
openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
components:
  schemas:
    Pet:
      type: string
      nullable: true
      example: fido
`)
	var buf bytes.Buffer
	openapi3conv.Upgrade(doc, openapi3conv.WithWriter(&buf))
	out := buf.String()
	assert.Contains(t, out, "openapi: 3.0.3 -> 3.2.0")
	assert.Contains(t, out, "nullable")
	assert.Contains(t, out, "example")
}

// ---------------------------------------------------------------------------
// Nil safety
// ---------------------------------------------------------------------------

func TestUpgrade_NilDocDoesNotPanic(t *testing.T) {
	// Per Pierre's guidance, Upgrade no longer validates input — doc
	// must be Validate()'d first. Nil is the one case we still defend
	// against, returning silently rather than panicking.
	openapi3conv.Upgrade(nil)
}

func TestUpgradeSchema_NilSchema(t *testing.T) {
	// Should not panic.
	openapi3conv.UpgradeSchema(nil)
}

func TestUpgradeSchema_OperatesOnSubtree(t *testing.T) {
	s := &openapi3.Schema{
		Type: &openapi3.Types{"object"},
		Properties: openapi3.Schemas{
			"x": &openapi3.SchemaRef{Value: &openapi3.Schema{
				Type:     &openapi3.Types{"string"},
				Nullable: true,
			}},
		},
	}
	openapi3conv.UpgradeSchema(s)
	x := s.Properties["x"].Value
	assert.Equal(t, openapi3.Types{"string", "null"}, *x.Type)
	assert.False(t, x.Nullable)
}
