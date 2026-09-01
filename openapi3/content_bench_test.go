package openapi3

import (
	"fmt"
	"testing"
)

var benchMediaType *MediaType

// benchmarkContent returns a document content with n declared media types.
func benchmarkContent(n int) Content {
	content := Content{
		"application/json":    NewMediaType(),
		"text/plain":          NewMediaType(),
		"multipart/form-data": NewMediaType(),
	}
	for i := len(content); i < n; i++ {
		content[fmt.Sprintf("application/vnd.example.v%d+json", i)] = NewMediaType()
	}
	return content
}

func BenchmarkContent_Get(b *testing.B) {
	benchmarks := []struct {
		name string
		keys int
		mime string
	}{
		// Declared verbatim: resolves on the first map lookup.
		{"exact", 3, "application/json"},
		// The most common shape on the wire: matches once parameters are stripped.
		{"parameters", 3, "application/json;charset=utf-8"},
		{"case_variant", 3, "APPLICATION/JSON"},
		{"case_variant_parameters", 3, "APPLICATION/JSON;charset=utf-8"},
		// Undeclared: walks every matching stage before giving up.
		{"no_match", 3, "image/png"},
		{"parameters_many_media_types", 20, "application/json;charset=utf-8"},
	}
	for _, bb := range benchmarks {
		content := benchmarkContent(bb.keys)
		b.Run(bb.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchMediaType = content.Get(bb.mime)
			}
		})
	}
}
