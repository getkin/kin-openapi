package openapi3filter

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

var benchBodyDecoder BodyDecoder

func BenchmarkGetBodyDecoder(b *testing.B) {
	benchmarks := []struct {
		name        string
		contentType string
	}{
		// Registered verbatim: resolves on the first map lookup.
		{"exact", "application/json"},
		{"case_variant", "APPLICATION/JSON"},
		// Unregistered: binary multipart parts take this path on every request.
		{"unregistered", "image/png"},
	}
	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchBodyDecoder, _ = getBodyDecoder(bb.contentType)
			}
		})
	}
}

// BenchmarkValidateRequestBody puts the media type lookups above in proportion
// to a whole body validation.
func BenchmarkValidateRequestBody(b *testing.B) {
	schema := openapi3.NewObjectSchema().
		WithProperty("name", openapi3.NewStringSchema()).
		WithProperty("code", openapi3.NewIntegerSchema())
	requestBody := openapi3.NewRequestBody().WithJSONSchema(schema).WithRequired(true)
	payload := []byte(`{"name":"foo","code":123}`)

	for _, contentType := range []string{
		"application/json",
		"application/json;charset=utf-8",
		"APPLICATION/JSON",
	} {
		b.Run(contentType, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				req, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))
				if err != nil {
					b.Fatal(err)
				}
				req.Header.Set(headerCT, contentType)
				input := &RequestValidationInput{Request: req}
				if err := ValidateRequestBody(b.Context(), input, requestBody); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
