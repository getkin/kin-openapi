package openapi3_test

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/internal/fuzzing"
	"github.com/getkin/kin-openapi/openapi3"
)

// maxFuzzInput bounds fuzzer inputs so the fuzzer spends its budget on
// conversion logic rather than on allocating huge documents.
const maxFuzzInput = 32 << 10

func newFuzzLoader() *openapi3.Loader {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(*openapi3.Loader, *url.URL) ([]byte, error) { return []byte(`{}`), nil } // deterministic
	return loader
}

// seedFromTestdata adds the YAML-specific seeds and the specs under testdata,
// then those of the APIs-guru corpus when it has been downloaded.
func seedFromTestdata(f *testing.F) {
	f.Helper()
	for _, seed := range openapi3.YAMLFuzzSeeds {
		f.Add([]byte(seed))
	}
	for _, seed := range fuzzing.Testdata(f, "testdata") {
		f.Add(seed)
	}
	for _, seed := range fuzzing.APIsGuru(f, "testdata", "openapi:") {
		f.Add(seed)
	}
}

// enc(dec) == enc(dec(enc(dec)))
func FuzzLoadFromData(f *testing.F) {
	seedFromTestdata(f)

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

// Asking for origins changes nothing about the document, and every recorded
// position lies inside the input.
func FuzzLoadWithOrigins(f *testing.F) {
	seedFromTestdata(f)

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxFuzzInput {
			t.Skip()
		}

		plain, plainErr := newFuzzLoader().LoadFromData(data)

		loader := newFuzzLoader()
		loader.IncludeOrigin = true
		withOrigins, err := loader.LoadFromData(data)
		require.Equal(t, plainErr == nil, err == nil, "without origins: %v\nwith origins: %v", plainErr, err)
		if err != nil {
			return
		}

		a, err := plain.MarshalJSON()
		require.NoError(t, err)
		b, err := withOrigins.MarshalJSON()
		require.NoError(t, err)
		require.JSONEq(t, string(a), string(b))

		lines := yamlLines(data)
		checkLocation := func(what string, loc openapi3.Location) {
			if loc.File != "" || loc.Line == 0 {
				return // another document, or not recorded
			}
			require.GreaterOrEqual(t, loc.Line, 1, what)
			require.LessOrEqual(t, loc.Line, lines, what)
			require.GreaterOrEqual(t, loc.Column, 1, what)
			if loc.EndLine != 0 {
				require.GreaterOrEqual(t, loc.EndLine, loc.Line, "%s: %+v", what, loc)
				require.LessOrEqual(t, loc.EndLine, lines, "%s: %+v", what, loc)
				require.GreaterOrEqual(t, loc.EndColumn, 1, "%s: %+v", what, loc)
			}
		}
		walkOrigins(withOrigins, func(o *openapi3.Origin) {
			if o.Key != nil {
				checkLocation("key", *o.Key)
			}
			for _, loc := range o.Fields {
				checkLocation("field "+loc.Name, loc)
			}
			for name, locs := range o.Sequences {
				for _, loc := range locs {
					checkLocation("sequence "+name, loc)
				}
			}
		})
	})
}

// yamlLines counts lines the way the YAML parser numbers them, which breaks
// on CR, LF, CRLF, NEL, LS and PS alike.
func yamlLines(data []byte) int {
	s := strings.ReplaceAll(string(data), "\r\n", "\n")
	return 1 + strings.Count(s, "\n") + strings.Count(s, "\r") +
		strings.Count(s, "\u0085") + strings.Count(s, "\u2028") + strings.Count(s, "\u2029")
}

var originType = reflect.TypeFor[openapi3.Origin]()

// walkOrigins calls fn on every *Origin reachable from v, unexported fields
// included since that is where Paths and Responses keep their entries. A
// resolved document is a graph, so each pointer is visited once.
func walkOrigins(v any, fn func(*openapi3.Origin)) {
	seen := make(map[uintptr]struct{})
	var walk func(reflect.Value)
	walk = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.Pointer:
			if v.IsNil() {
				return
			}
			if _, ok := seen[v.Pointer()]; ok {
				return
			}
			seen[v.Pointer()] = struct{}{}
			if v.Type().Elem() == originType {
				fn((*openapi3.Origin)(v.UnsafePointer()))
				return
			}
			walk(v.Elem())
		case reflect.Interface:
			if !v.IsNil() {
				walk(v.Elem())
			}
		case reflect.Struct:
			for i := range v.NumField() {
				walk(v.Field(i))
			}
		case reflect.Slice, reflect.Array:
			for i := range v.Len() {
				walk(v.Index(i))
			}
		case reflect.Map:
			for it := v.MapRange(); it.Next(); {
				walk(it.Value())
			}
		}
	}
	walk(reflect.ValueOf(v))
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
