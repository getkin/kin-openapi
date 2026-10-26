package openapi3

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	yaml "go.yaml.in/yaml/v3"

	"github.com/getkin/kin-openapi/internal/fuzzing"
)

// maxYAMLFuzzInput bounds fuzzer inputs, see maxFuzzInput.
const maxYAMLFuzzInput = 32 << 10

// yamlFuzzSeeds exercise what only the YAML decode path sees: anchors and
// aliases, merge keys, tags, block scalars, non-string keys and YAML 1.1
// implicit resolution. Shared by every fuzz target that loads documents.
var yamlFuzzSeeds = []string{
	// Anchors and aliases, on a schema, a parameter and a $ref.
	`openapi: 3.0.0
info: {title: t, version: "1"}
paths:
  /a/{id}:
    parameters:
      - &id {name: id, in: path, required: true, schema: {type: string}}
    get:
      parameters: [*id]
      responses:
        200: &ok
          description: ok
          content:
            application/json:
              schema: &s {type: object, properties: {a: {type: integer}}}
        201: *ok
components:
  schemas:
    A: *s
    B: {$ref: &r '#/components/schemas/A'}
    C: {$ref: *r}
`,
	// Merge keys: in a path item, in Paths itself, from a sequence of sources.
	`openapi: 3.0.0
info: {title: t, version: "1"}
x-base: &base
  description: base
  responses: {"200": {description: ok}}
x-more: &more
  summary: more
  description: overridden
paths:
  /m:
    get:
      <<: [*base, *more]
      operationId: m
  <<: {/n: {get: {responses: {default: {description: d}}}}}
components:
  schemas:
    S:
      <<: {type: object, required: [a]}
      properties: {a: {type: string}}
`,
	// Block scalars, including one ending the block, and comments.
	`openapi: 3.0.0
info:
  title: t
  version: "1"
  description: |
    line one

    line three
paths:
  /b:
    get:
      responses:
        "200":
          description: >-
            folded
            text   # not a comment
        # trailing comment
`,
	// YAML 1.1 implicit resolution: timestamps, hex, underscores, bools,
	// non-string keys and a tagged scalar.
	`openapi: 3.0.0
info: {title: t, version: 1.0}
paths: {}
components:
  schemas:
    T:
      type: string
      format: date-time
      example: 2020-06-11T16:32:50-03:00
      x-ts: !!timestamp 2001-01-01
      x-keys: {18_24: a, 0x2A: b, 1.5: c, true: d}
    I:
      type: integer
      enum: [0x2A, 4_2, 052, 42]
      example: 42
      default: 0x2A
      nullable: yes
      maximum: 1e3
`,
	// OpenAPI 3.1: boolean schemas, type lists, exclusive bounds as numbers,
	// explicitly null responses, webhooks.
	`openapi: 3.1.0
info: {title: t, version: "1"}
paths:
  /w:
    get:
      responses: ~
webhooks:
  hook:
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: [object, "null"]
              additionalProperties: false
              unevaluatedProperties: true
              properties:
                n: {type: number, exclusiveMinimum: 0, exclusiveMaximum: 10}
                any: true
                none: false
`,
	// Flow style, CRLF line endings, a document marker.
	"---\r\n{openapi: 3.0.0, info: {title: t, version: \"1\"},\r\n paths: {/f: {get: {responses: {\"200\": {description: ok}}}}}}\r\n",
	// $ref siblings and extensions on maplike collections.
	`openapi: 3.1.0
info: {title: t, version: "1"}
paths:
  x-p: 1
  /r:
    $ref: '#/components/pathItems/R'
    summary: s
components:
  pathItems:
    R:
      get:
        responses:
          x-r: 2
          "200": {$ref: '#/components/responses/OK', description: d}
        callbacks:
          cb:
            x-c: 3
            '{$request.body#/u}': {post: {responses: {"200": {description: ok}}}}
  responses:
    OK: {description: ok}
`,
}

// decodeYAML decodes YAML (or JSON, which parses as YAML) straight into v,
// the way unmarshal does before falling back to JSON.
func decodeYAML(data []byte, v any) error {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return err
	}
	if err := prepareScalars(&root); err != nil {
		return err
	}
	return root.Decode(v)
}

func requireSameDoc(t *testing.T, expected, actual *T, msgAndArgs ...any) {
	t.Helper()
	e, err := expected.MarshalJSON()
	require.NoError(t, err)
	a, err := actual.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, string(e), string(a), msgAndArgs...)
}

// The YAML decode path must agree with the JSON one:
//   - a JSON document both paths accept decodes to the same thing, and
//   - any document, re-encoded as block YAML, decodes back to itself.
func FuzzYAMLMatchesJSON(f *testing.F) {
	for _, seed := range yamlFuzzSeeds {
		f.Add([]byte(seed))
	}
	for _, seed := range fuzzing.Testdata(f, "testdata") {
		f.Add(seed)
	}
	for _, seed := range fuzzing.APIsGuru(f, "testdata", "openapi:") {
		f.Add(seed)
	}
	f.Add([]byte(`{"openapi":"3.0.0","info":{"title":"t","version":"1"},"paths":{"/a":{"get":{"responses":{"200":{"description":"ok"}}}}}}`))
	f.Add([]byte(`{"openapi":"3.0.0","paths":{"x-a":{"<<":1},"/x":{}},"components":{"schemas":{"<<":{"type":"string"}}}}`))
	f.Add([]byte(`{"openapi":"3.1.0","components":{"schemas":{"a":{"enum":[1,1.5,"1",true,null],"example":12345678901234567890,"x-e":{"1":[{}]}}}}}`))
	f.Add([]byte(`{"openapi":"3.0.0","info":{"title":"é\/\t","version":"2020-01-01"},"paths":{}}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxYAMLFuzzInput {
			t.Skip()
		}

		if json.Valid(data) {
			var viaJSON, viaYAML T
			if unmarshalJSON(data, &viaJSON) == nil && decodeYAML(data, &viaYAML) == nil {
				requireSameDoc(t, &viaJSON, &viaYAML, "data:\n%s", data)
			}
		}

		var doc T
		if decodeYAML(data, &doc) != nil {
			return
		}
		j, err := doc.MarshalJSON()
		require.NoError(t, err)
		var generic any
		require.NoError(t, json.Unmarshal(j, &generic))
		if fuzzing.LosesInYAML(generic) {
			return
		}
		y, err := yaml.Marshal(generic)
		require.NoError(t, err)

		var back T
		require.NoError(t, decodeYAML(y, &back), "yaml:\n%s", y)
		requireSameDoc(t, &doc, &back, "yaml:\n%s\ndata:\n%s", y, data)
	})
}

// YAMLFuzzSeeds exposes yamlFuzzSeeds to the external test package.
var YAMLFuzzSeeds = yamlFuzzSeeds
