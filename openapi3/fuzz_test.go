package openapi3_test

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// maxFuzzInput bounds fuzzer inputs so the fuzzer spends its budget on
// conversion logic rather than on allocating huge documents.
const maxFuzzInput = 32 << 10

// maxFuzzSeed keeps the seed corpus small enough that `go test ./...` still
// replays it on every run without noticeably slowing the suite down.
const maxFuzzSeed = 16 << 10

func newFuzzLoader() *openapi3.Loader {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(*openapi3.Loader, *url.URL) ([]byte, error) { return []byte(`{}`), nil } // deterministic
	return loader
}

func seedFromTestdata(f *testing.F, globs ...string) {
	f.Helper()
	for _, glob := range globs {
		paths, err := filepath.Glob(glob)
		require.NoError(f, err)
		for _, path := range paths {
			data, err := os.ReadFile(path)
			require.NoError(f, err)
			if len(data) <= maxFuzzSeed {
				f.Add(data)
			}
		}
	}
}

// enc(dec) == enc(dec(enc(dec)))
func FuzzLoadFromData(f *testing.F) {
	seedFromTestdata(f, "testdata/*.yml", "testdata/*.yaml", "testdata/*.json")

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxFuzzInput {
			t.Skip()
		}

		loader := newFuzzLoader()
		doc, err := loader.LoadFromData(data)
		if err != nil {
			return
		}
		require.NotNil(t, doc)

		// Validation runs over attacker-shaped documents: it may reject them but
		// must never panic nor fail to terminate.
		valid := doc.Validate(t.Context())

		once, err := doc.MarshalJSON()
		require.NoError(t, err)

		loader = newFuzzLoader()
		// The loader must accept its own output
		reloaded, err := loader.LoadFromData(once)
		require.NoError(t, err, "\nonce: %s\nvalid: %+v\ndata:\n%s", string(once), valid, string(data))

		twice, err := reloaded.MarshalJSON()
		require.NoError(t, err)

		require.JSONEq(t, string(once), string(twice))

		// Validation should be deterministic
		require.Equal(t, valid, reloaded.Validate(t.Context()))
	})
}

// No crash when validating values against schemas
func FuzzSchemaVisitJSON(f *testing.F) {
	f.Add([]byte(`{"type":"string","minLength":1,"pattern":"^a+$"}`), []byte(`"aaa"`))
	f.Add([]byte(`{"type":"integer","multipleOf":0.01,"maximum":10}`), []byte(`3`))
	f.Add([]byte(`{"type":"array","items":{"type":"number"},"uniqueItems":true}`), []byte(`[1,2,1]`))
	f.Add([]byte(`{"type":"object","required":["a"],"additionalProperties":false}`), []byte(`{"a":null,"b":1}`))
	f.Add([]byte(`{"allOf":[{"type":"string"},{"maxLength":2}]}`), []byte(`"abc"`))
	f.Add([]byte(`{"oneOf":[{"type":"string"},{"type":"number"}],"nullable":true}`), []byte(`null`))
	f.Add([]byte(`{"type":"string","format":"date-time"}`), []byte(`"not a date"`))

	f.Fuzz(func(t *testing.T, schemaJSON, valueJSON []byte) {
		if len(schemaJSON) > maxFuzzInput || len(valueJSON) > maxFuzzInput {
			t.Skip()
		}

		var schema openapi3.Schema
		if err := schema.UnmarshalJSON(schemaJSON); err != nil {
			t.Skip()
		}

		var value any
		if err := json.Unmarshal(valueJSON, &value); err != nil {
			t.Skip()
		}

		_ = schema.VisitJSON(value)
		_ = schema.VisitJSON(value, openapi3.EnableFormatValidation())
		_ = schema.VisitJSON(value, openapi3.EnableJSONSchema2020())
	})
}
