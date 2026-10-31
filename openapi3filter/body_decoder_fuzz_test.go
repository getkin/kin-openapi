package openapi3filter_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
)

var fuzzedBodyDecoders = []string{
	"application/json",
	"application/octet-stream",
	"application/x-www-form-urlencoded",
	"application/yaml",
	"multipart/form-data",
	"text/csv",
	"text/plain",
}

// The registered body decoders never panic, and return only what the
// BodyDecoder contract allows: a primitive, []any or map[string]any.
func FuzzBodyDecoder(f *testing.F) {
	const object = `{"type":"object","properties":{"a":{"type":"integer"},"b":{"type":"array","items":{"type":"string"}},"o":{"type":"object","properties":{"x":{"type":"boolean"}}},"f":{"type":"string","format":"binary"}}}`
	f.Add(uint8(0), "application/json", []byte(`{"a":1,"b":["x"]}`), []byte(object))
	f.Add(uint8(1), "application/octet-stream", []byte("\x00\x01"), []byte(`{"type":"string","format":"binary"}`))
	f.Add(uint8(2), "application/x-www-form-urlencoded", []byte(`a=1&b=x&b=y&o[x]=true`), []byte(object))
	f.Add(uint8(2), "application/x-www-form-urlencoded", []byte(`a=1`), []byte(`{"type":"object","additionalProperties":{"type":"integer"}}`))
	f.Add(uint8(3), "application/yaml", []byte("a: 1\nb: [x]\n"), []byte(object))
	f.Add(uint8(4), "multipart/form-data; boundary=b", []byte("--b\r\nContent-Disposition: form-data; name=\"a\"\r\n\r\n1\r\n--b\r\nContent-Disposition: form-data; name=\"b\"\r\n\r\nx\r\n--b\r\nContent-Disposition: form-data; name=\"o\"\r\nContent-Type: application/json\r\n\r\n{\"x\":true}\r\n--b\r\nContent-Disposition: form-data; name=\"f\"; filename=\"f\"\r\n\r\nraw\r\n--b--\r\n"), []byte(object))
	f.Add(uint8(5), "text/csv", []byte("a,b\n1,2\n"), []byte(`{"type":"string"}`))
	f.Add(uint8(6), "text/plain", []byte("hi"), []byte(`{"type":"string"}`))

	f.Fuzz(func(t *testing.T, which uint8, contentType string, body, schemaJSON []byte) {
		if len(body) > maxFuzzInput || len(schemaJSON) > maxFuzzInput || len(contentType) > 1<<10 {
			t.Skip()
		}
		decoder := openapi3filter.RegisteredBodyDecoder(fuzzedBodyDecoders[int(which)%len(fuzzedBodyDecoders)])

		var schema *openapi3.SchemaRef
		if len(schemaJSON) != 0 {
			schema = &openapi3.SchemaRef{}
			if err := schema.UnmarshalJSON(schemaJSON); err != nil {
				t.Skip()
			}
			if schema.Ref != "" {
				t.Skip() // unresolved
			}
		}
		encFn := func(string) *openapi3.Encoding { return nil }

		value, err := decoder(bytes.NewReader(body), http.Header{"Content-Type": {contentType}}, schema, encFn)
		if err != nil {
			return
		}
		requireBodyDecoderValue(t, value)

		if schema != nil && schema.Value != nil {
			_ = schema.Value.VisitJSON(value) // no crash
		}
	})
}

func requireBodyDecoderValue(t *testing.T, value any) {
	t.Helper()
	switch v := value.(type) {
	case nil, bool, string, json.Number, float64, int64, int, uint64, float32, int32:
	case []any:
		for _, e := range v {
			requireBodyDecoderValue(t, e)
		}
	case map[string]any:
		for _, e := range v {
			requireBodyDecoderValue(t, e)
		}
	default:
		require.Failf(t, "not a primitive, []any nor map[string]any", "%T: %#v", value, value)
	}
}
