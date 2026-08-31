package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContent_Get(t *testing.T) {
	fallback := NewMediaType()
	wildcard := NewMediaType()
	stripped := NewMediaType()
	fullMatch := NewMediaType()
	caseVariant := NewMediaType()
	malformed := NewMediaType()
	content := Content{
		"*/*":                             fallback,
		"application/*":                   wildcard,
		"application/json":                stripped,
		"application/json;encoding=utf-8": fullMatch,
	}
	contentWithoutWildcards := Content{
		"application/json":                stripped,
		"application/json;encoding=utf-8": fullMatch,
	}
	contentWithCaseVariants := Content{
		"application/json": stripped,
		"APPLICATION/JSON": caseVariant,
	}
	contentWithMalformedType := Content{
		"text": malformed,
	}
	tests := []struct {
		name    string
		content Content
		mime    string
		want    *MediaType
	}{
		{
			name:    "missing",
			content: contentWithoutWildcards,
			mime:    "text/plain;encoding=utf-8",
			want:    nil,
		},
		{
			name:    "full match",
			content: content,
			mime:    "application/json;encoding=utf-8",
			want:    fullMatch,
		},
		{
			name:    "full match case insensitive",
			content: content,
			mime:    "APPLICATION/JSON;encoding=utf-8",
			want:    fullMatch,
		},
		{
			name:    "parameter value case sensitive",
			content: content,
			mime:    "APPLICATION/JSON;encoding=UTF-8",
			want:    stripped,
		},
		{
			name:    "stripped match",
			content: content,
			mime:    "application/json;encoding=utf-16",
			want:    stripped,
		},
		{
			name:    "stripped match case insensitive",
			content: content,
			mime:    "APPLICATION/JSON;encoding=utf-16",
			want:    stripped,
		},
		{
			name:    "wildcard match",
			content: content,
			mime:    "application/yaml;encoding=utf-16",
			want:    wildcard,
		},
		{
			name:    "wildcard match case insensitive",
			content: content,
			mime:    "APPLICATION/YAML;encoding=utf-16",
			want:    wildcard,
		},
		{
			name:    "fallback match",
			content: content,
			mime:    "text/plain;encoding=utf-16",
			want:    fallback,
		},
		{
			name:    "invalid mime type",
			content: content,
			mime:    "text;encoding=utf16",
			want:    nil,
		},
		{
			name:    "missing no encoding",
			content: contentWithoutWildcards,
			mime:    "text/plain",
			want:    nil,
		},
		{
			name:    "stripped match no encoding",
			content: content,
			mime:    "application/json",
			want:    stripped,
		},
		{
			name:    "wildcard match no encoding",
			content: content,
			mime:    "application/yaml",
			want:    wildcard,
		},
		{
			name:    "fallback match no encoding",
			content: content,
			mime:    "text/plain",
			want:    fallback,
		},
		{
			name:    "invalid mime type no encoding",
			content: content,
			mime:    "text",
			want:    nil,
		},
		{
			name:    "exact match takes precedence",
			content: contentWithCaseVariants,
			mime:    "APPLICATION/JSON",
			want:    caseVariant,
		},
		{
			name:    "invalid mime type remains case sensitive",
			content: contentWithMalformedType,
			mime:    "TEXT",
			want:    nil,
		},
		{
			name:    "missing mime type",
			content: content,
			mime:    "",
			want:    fallback,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.content.Get(tt.mime)
			require.Same(t, tt.want, got)
		})
	}
}
