package openapi2

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/internal/fuzzing"
	"github.com/getkin/kin-openapi/internal/yamlconv"
)

// maxFuzzInput bounds fuzzer inputs so the fuzzer spends its budget on
// decoding rather than on allocating huge documents.
const maxFuzzInput = 32 << 10

// A swagger document, JSON or YAML, loads back to itself from its own JSON and
// YAML forms.
func FuzzUnmarshalFromData(f *testing.F) {
	for _, seed := range fuzzing.Testdata(f, "testdata") {
		f.Add(seed)
	}
	// Read from openapi3, the one copy the fuzz workflow fetches.
	for _, seed := range fuzzing.APIsGuru(f, "../openapi3/testdata", "swagger:") {
		f.Add(seed)
	}
	f.Add([]byte(`swagger: "2.0"
info: {title: t, version: "1.0"}
x-date: 2020-06-11
paths:
  /a/{id}:
    get:
      parameters:
        - &id {name: id, in: path, required: true, type: string}
      responses:
        200: &ok {description: ok, schema: {type: integer, enum: [0x2A, 4_2]}}
        201: *ok
  /b:
    <<: {get: {responses: {default: {description: d}}}}
definitions:
  A:
    type: object
    x-keys: {18_24: a, 1.5: b}
    properties: {b: {$ref: '#/definitions/A'}}
`))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxFuzzInput {
			t.Skip()
		}

		var doc T
		if unmarshal(data, &doc) != nil {
			return
		}
		once, err := json.Marshal(&doc)
		require.NoError(t, err)

		var fromJSON T
		require.NoError(t, unmarshal(once, &fromJSON), "json:\n%s", once)
		twice, err := json.Marshal(&fromJSON)
		require.NoError(t, err)
		require.JSONEq(t, string(once), string(twice))

		var generic any
		require.NoError(t, json.Unmarshal(once, &generic))
		if fuzzing.LosesInYAML(generic) {
			return
		}
		y, err := yamlconv.Marshal(&doc)
		require.NoError(t, err)
		var fromYAML T
		require.NoError(t, unmarshal(y, &fromYAML), "yaml:\n%s", y)
		thrice, err := json.Marshal(&fromYAML)
		require.NoError(t, err)
		require.JSONEq(t, string(once), string(thrice), "yaml:\n%s", y)
	})
}
