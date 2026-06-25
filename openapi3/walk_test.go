package openapi3_test

import (
	"errors"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// walkSpec exercises schemas in components and under a path operation
// (parameter, request body, response content and header), inline nested
// sub-schemas (deterministic locations), a shared $ref (dedup), and a
// self-referential schema (cycle safety).
var walkSpec = []byte(`
openapi: 3.0.3
info: {title: t, version: "1"}
paths:
  /pets:
    get:
      parameters:
        - name: limit
          in: query
          schema: {type: integer}
      responses:
        "200":
          description: ok
          headers:
            X-Next:
              schema: {type: string}
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Pet'
components:
  schemas:
    Pet:
      type: object
      properties:
        id: {type: integer}
        friends:
          type: array
          items: {type: string}
        tag:
          $ref: '#/components/schemas/Tag'
    Tag:
      type: string
    Node:
      type: object
      properties:
        next:
          $ref: '#/components/schemas/Node'
`)

func loadWalkSpec(t *testing.T) *openapi3.T {
	t.Helper()
	doc, err := openapi3.NewLoader().LoadFromData(walkSpec)
	require.NoError(t, err)
	return doc
}

func TestWalkSchemas(t *testing.T) {
	doc := loadWalkSpec(t)

	locsByPtr := map[*openapi3.Schema][]string{}
	var visited []string
	err := doc.WalkSchemas(func(loc string, sr *openapi3.SchemaRef) error {
		require.NotNil(t, sr)
		require.NotNil(t, sr.Value)
		locsByPtr[sr.Value] = append(locsByPtr[sr.Value], loc)
		visited = append(visited, loc)
		return nil
	})
	require.NoError(t, err)

	// Every distinct schema is visited exactly once: shared $refs are deduped
	// and the self-referential Node does not loop forever.
	for s, locs := range locsByPtr {
		require.Lenf(t, locs, 1, "schema %p visited %d times: %v", s, len(locs), locs)
	}

	// Deterministic locations for inline (non-$ref) schemas across components,
	// nested sub-schemas, and the operation's parameter/response surfaces.
	for _, want := range []string{
		"/components/schemas/Pet",
		"/components/schemas/Pet/properties/id",
		"/components/schemas/Pet/properties/friends",
		"/components/schemas/Pet/properties/friends/items",
		"/components/schemas/Node",
		"/paths/~1pets/get/parameters/0/schema",
		"/paths/~1pets/get/responses/200/headers/X-Next/schema",
		"/paths/~1pets/get/responses/200/content/application~1json/schema",
	} {
		require.Contains(t, visited, want)
	}
}

func TestWalkSchemas_SkipSubtree(t *testing.T) {
	doc := loadWalkSpec(t)

	var visited []string
	err := doc.WalkSchemas(func(loc string, sr *openapi3.SchemaRef) error {
		visited = append(visited, loc)
		if loc == "/components/schemas/Pet" {
			return openapi3.SkipSubtree
		}
		return nil
	})
	require.NoError(t, err)

	// Pet itself is visited, but SkipSubtree prevents descent into its members.
	require.Contains(t, visited, "/components/schemas/Pet")
	require.NotContains(t, visited, "/components/schemas/Pet/properties/id")
	require.NotContains(t, visited, "/components/schemas/Pet/properties/friends")
}

func TestWalkSchemas_ErrorAborts(t *testing.T) {
	doc := loadWalkSpec(t)

	boom := errors.New("boom")
	count := 0
	err := doc.WalkSchemas(func(loc string, sr *openapi3.SchemaRef) error {
		count++
		return boom
	})
	require.ErrorIs(t, err, boom)
	require.Equal(t, 1, count, "walk should stop at the first error")
}
