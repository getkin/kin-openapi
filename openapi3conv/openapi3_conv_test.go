package openapi3conv_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

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
// expected 3.1 output than walking structs.
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

func TestUpgradeTo31_BumpsVersion(t *testing.T) {
	doc := loadV30(t, `
openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
`)
	require.NoError(t, openapi3conv.UpgradeTo31(doc))
	assert.Equal(t, "3.1.1", doc.OpenAPI)
}

func TestUpgradeTo31_CustomTargetVersion(t *testing.T) {
	doc := loadV30(t, `
openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
`)
	require.NoError(t, openapi3conv.UpgradeTo31WithOptions(doc, openapi3conv.UpgradeOptions{
		TargetVersion: "3.1.0",
	}))
	assert.Equal(t, "3.1.0", doc.OpenAPI)
}

func TestUpgradeTo31_SkipVersionBump(t *testing.T) {
	doc := loadV30(t, `
openapi: 3.0.3
info: {title: t, version: '1'}
paths: {}
`)
	require.NoError(t, openapi3conv.UpgradeTo31WithOptions(doc, openapi3conv.UpgradeOptions{
		SkipVersionBump: true,
	}))
	assert.Equal(t, "3.0.3", doc.OpenAPI)
}

// ---------------------------------------------------------------------------
// Nullable rewrite
// ---------------------------------------------------------------------------

func TestUpgradeTo31_NullableWithType(t *testing.T) {
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
	require.NoError(t, openapi3conv.UpgradeTo31(doc))
	pet := doc.Components.Schemas["Pet"].Value
	require.NotNil(t, pet.Type)
	assert.Equal(t, openapi3.Types{"string", "null"}, *pet.Type)
	assert.False(t, pet.Nullable, "nullable should be cleared after rewrite")
}

func TestUpgradeTo31_NullableAlreadyHasNullInTypeArray(t *testing.T) {
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
	require.NoError(t, openapi3conv.UpgradeTo31(doc))
	pet := doc.Components.Schemas["Pet"].Value
	assert.Equal(t, openapi3.Types{"string", "null"}, *pet.Type, "no duplicate null appended")
	assert.False(t, pet.Nullable)
}

func TestUpgradeTo31_NullableNoType(t *testing.T) {
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
	require.NoError(t, openapi3conv.UpgradeTo31(doc))
	pet := doc.Components.Schemas["Pet"].Value
	assert.False(t, pet.Nullable)
	assert.Nil(t, pet.Type)
}

func TestUpgradeTo31_NullableInsideProperties(t *testing.T) {
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
	require.NoError(t, openapi3conv.UpgradeTo31(doc))
	props := doc.Components.Schemas["Pet"].Value.Properties
	assert.Equal(t, openapi3.Types{"string", "null"}, *props["name"].Value.Type)
	assert.Equal(t, openapi3.Types{"integer", "null"}, *props["ageInYears"].Value.Type)
}

// ---------------------------------------------------------------------------
// Exclusive-bound rewrite
// ---------------------------------------------------------------------------

func TestUpgradeTo31_ExclusiveMinTrueWithMinimum(t *testing.T) {
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
	require.NoError(t, openapi3conv.UpgradeTo31(doc))
	score := doc.Components.Schemas["Score"].Value
	assert.Nil(t, score.Min, "Min cleared")
	require.NotNil(t, score.ExclusiveMin.Value)
	assert.Equal(t, 5.0, *score.ExclusiveMin.Value)
	assert.Nil(t, score.ExclusiveMin.Bool)
}

func TestUpgradeTo31_ExclusiveMaxTrueWithMaximum(t *testing.T) {
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
	require.NoError(t, openapi3conv.UpgradeTo31(doc))
	score := doc.Components.Schemas["Score"].Value
	assert.Nil(t, score.Max)
	require.NotNil(t, score.ExclusiveMax.Value)
	assert.Equal(t, 100.0, *score.ExclusiveMax.Value)
	assert.Nil(t, score.ExclusiveMax.Bool)
}

func TestUpgradeTo31_ExclusiveMinFalseDropped(t *testing.T) {
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
	require.NoError(t, openapi3conv.UpgradeTo31(doc))
	score := doc.Components.Schemas["Score"].Value
	require.NotNil(t, score.Min)
	assert.Equal(t, 5.0, *score.Min)
	assert.False(t, score.ExclusiveMin.IsSet(), "exclusiveMinimum: false should be dropped")
}

func TestUpgradeTo31_ExclusiveBoundsLeavesNumericIntact(t *testing.T) {
	// A document already in 3.1 numeric form should round-trip unchanged.
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

	require.NoError(t, openapi3conv.UpgradeTo31(doc))
	score := doc.Components.Schemas["Score"].Value
	require.NotNil(t, score.ExclusiveMin.Value)
	assert.Equal(t, 5.0, *score.ExclusiveMin.Value)
	assert.Nil(t, score.Min)
}

// ---------------------------------------------------------------------------
// Example -> Examples rewrite
// ---------------------------------------------------------------------------

func TestUpgradeTo31_ExampleToExamples(t *testing.T) {
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
	require.NoError(t, openapi3conv.UpgradeTo31(doc))
	pet := doc.Components.Schemas["Pet"].Value
	assert.Nil(t, pet.Example)
	require.Len(t, pet.Examples, 1)
	assert.Equal(t, "fido", pet.Examples[0])
}

func TestUpgradeTo31_ExampleAppendsToExistingExamples(t *testing.T) {
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
	require.NoError(t, openapi3conv.UpgradeTo31(doc))
	pet := doc.Components.Schemas["Pet"].Value
	assert.Nil(t, pet.Example)
	assert.Equal(t, []any{"rex", "fido"}, pet.Examples)
}

// ---------------------------------------------------------------------------
// Idempotence
// ---------------------------------------------------------------------------

func TestUpgradeTo31_Idempotent(t *testing.T) {
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
	require.NoError(t, openapi3conv.UpgradeTo31(doc))
	first := marshalJSON(t, doc)

	require.NoError(t, openapi3conv.UpgradeTo31(doc))
	second := marshalJSON(t, doc)

	assert.Equal(t, first, second, "second pass must be a no-op")
}

// ---------------------------------------------------------------------------
// Walks reachable from operations and request/response bodies
// ---------------------------------------------------------------------------

func TestUpgradeTo31_WalksOperationSchemas(t *testing.T) {
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
	require.NoError(t, openapi3conv.UpgradeTo31(doc))

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

func TestUpgradeTo31_CycleSafe(t *testing.T) {
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

	done := make(chan error, 1)
	go func() { done <- openapi3conv.UpgradeTo31(doc) }()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-context.Background().Done():
		t.Fatal("UpgradeTo31 hung on a cyclic schema")
	}

	// Sanity: the rewrite still ran on the non-cyclic property.
	name := cycle.Properties["name"].Value
	assert.Equal(t, openapi3.Types{"string", "null"}, *name.Type)
}

// ---------------------------------------------------------------------------
// Verbose logging
// ---------------------------------------------------------------------------

func TestUpgradeTo31_VerboseLogsRewrites(t *testing.T) {
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
	require.NoError(t, openapi3conv.UpgradeTo31WithOptions(doc, openapi3conv.UpgradeOptions{
		Verbose: &buf,
	}))
	out := buf.String()
	assert.Contains(t, out, "openapi: 3.0.3 -> 3.1.1")
	assert.Contains(t, out, "nullable")
	assert.Contains(t, out, "example")
}

// ---------------------------------------------------------------------------
// Nil safety
// ---------------------------------------------------------------------------

func TestUpgradeTo31_NilDoc(t *testing.T) {
	err := openapi3conv.UpgradeTo31(nil)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "nil"))
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
