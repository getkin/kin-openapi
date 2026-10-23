package openapi3

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalExtensions(t *testing.T) {
	t.Parallel()

	for _, data := range []string{
		`{}`,
		`null`,
		` { "a" : 1 , "x-b" : {"c": [1, "}", {"d": "\"]"}]}, "x-c": null } `,
		`{"x-é": "é", "a\"b": true, "x-\\": [], "x-n": -1.5e3}`,
		`{"x-d": 1, "x-d": 2}`,
		"{\n\t\"x-a\":\r\n\"tab\\tquote\\\"\"}",
	} {
		t.Run(data, func(t *testing.T) {
			var want map[string]any
			require.NoError(t, json.Unmarshal([]byte(data), &want))
			delete(want, "a")
			if len(want) == 0 {
				want = nil
			}
			require.Equal(t, want, unmarshalExtensions([]byte(data), "a"))
		})
	}
}

func TestJSONMembersNotAnObject(t *testing.T) {
	t.Parallel()

	for _, data := range []string{``, `[]`, `"a"`, `1`, `{"a"}`, `{"a":1,}`, `{"a":"b`, `{"a":[}`} {
		err := jsonMembers([]byte(data), func(string, []byte) error { return nil })
		require.ErrorIs(t, err, errNotJSONObject, data)
	}
}

func TestRefNameIsCaseSensitive(t *testing.T) {
	t.Parallel()

	loader := NewLoader()
	doc, err := loader.LoadFromData([]byte(`{
		"openapi": "3.0.3",
		"info": {"title": "t", "version": "1"},
		"paths": {},
		"components": {
			"schemas": {
				"A": {"$REF": "#/components/schemas/B"},
				"B": {"type": "string"}
			},
			"parameters": {
				"P": {"$REF": "#/components/parameters/Q", "name": "p", "in": "query"},
				"Q": {"name": "q", "in": "query"}
			}
		}
	}`))
	require.NoError(t, err)
	require.Empty(t, doc.Components.Schemas["A"].Ref)
	require.Empty(t, doc.Components.Parameters["P"].Ref)
	require.Equal(t, "p", doc.Components.Parameters["P"].Value.Name)
}

func BenchmarkLoadFromData(b *testing.B) {
	data := benchmarkDocument(300)
	b.ReportAllocs()
	for b.Loop() {
		loader := NewLoader()
		if _, err := loader.LoadFromData(data); err != nil {
			b.Fatal(err)
		}
	}
}

// benchmarkDocument is a JSON document of the given number of paths, each
// with two operations whose parameters, bodies and responses refer to
// component schemas built with allOf.
func benchmarkDocument(paths int) []byte {
	var b strings.Builder
	b.WriteString(`{"openapi":"3.0.3","info":{"title":"t","version":"1","x-owner":"team"},"paths":{`)
	for i := range paths {
		if i > 0 {
			b.WriteByte(',')
		}
		s := i % 20
		fmt.Fprintf(&b, `"/items%d/{id}":{"parameters":[{"name":"id","in":"path","required":true,"schema":{"type":"string","pattern":"^[a-z0-9]+$"}}],`, i)
		fmt.Fprintf(&b, `"get":{"operationId":"get%d","x-rate":10,"parameters":[{"name":"limit","in":"query","schema":{"type":"integer","minimum":1,"maximum":100}}],`, i)
		fmt.Fprintf(&b, `"responses":{"200":{"description":"ok","content":{"application/json":{"schema":{"type":"array","items":{"$ref":"#/components/schemas/Item%d"}}}}},"default":{"$ref":"#/components/responses/Error"}}},`, s)
		fmt.Fprintf(&b, `"put":{"operationId":"put%d","requestBody":{"required":true,"content":{"application/json":{"schema":{"$ref":"#/components/schemas/Item%d"}}}},`, i, s)
		fmt.Fprintf(&b, `"responses":{"204":{"description":"done"},"default":{"$ref":"#/components/responses/Error"}}}}`)
	}
	b.WriteString(`},"components":{"responses":{"Error":{"description":"error","content":{"application/json":{"schema":{"$ref":"#/components/schemas/Error"}}}}},"schemas":{`)
	b.WriteString(`"Error":{"type":"object","required":["code"],"properties":{"code":{"type":"integer"},"message":{"type":"string"}}},`)
	b.WriteString(`"Base":{"type":"object","required":["id"],"properties":{"id":{"type":"string","format":"uuid"},"created":{"type":"string","format":"date-time"}}}`)
	for s := range 20 {
		fmt.Fprintf(&b, `,"Item%d":{"allOf":[{"$ref":"#/components/schemas/Base"},{"type":"object","x-kind":"item","properties":{"name":{"type":"string","maxLength":64},"tags":{"type":"array","items":{"type":"string","enum":["a","b","c"]}},"attributes":{"type":"object","additionalProperties":{"type":"string"}}}}]}`, s)
	}
	b.WriteString(`}}}`)
	return []byte(b.String())
}
