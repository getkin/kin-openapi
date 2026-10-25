package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContent_Get(t *testing.T) {
	t.Parallel()

	fallback := NewMediaType()
	wildcard := NewMediaType()
	stripped := NewMediaType()
	fullMatch := NewMediaType()
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
			name:    "stripped match",
			content: content,
			mime:    "application/json;encoding=utf-16",
			want:    stripped,
		},
		{
			name:    "wildcard match",
			content: content,
			mime:    "application/yaml;encoding=utf-16",
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
			name:    "missing mime type",
			content: content,
			mime:    "",
			want:    fallback,
		},
		{
			name:    "case insensitive full match",
			content: content,
			mime:    "Application/JSON;Encoding=UTF-8",
			want:    fullMatch,
		},
		{
			name:    "case insensitive stripped match",
			content: content,
			mime:    "APPLICATION/JSON",
			want:    stripped,
		},
		{
			name:    "case insensitive wildcard match",
			content: content,
			mime:    "Application/YAML",
			want:    wildcard,
		},
		{
			name:    "case insensitive key",
			content: Content{"Application/JSON": stripped},
			mime:    "application/json",
			want:    stripped,
		},
		{
			name:    "key listing several media ranges",
			content: Content{"text/csv, Application/JSON": stripped, "*/*": fallback},
			mime:    "application/json",
			want:    stripped,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.content.Get(tt.mime)
			require.Same(t, tt.want, got)
		})
	}
}

func TestContent_ValidateKeysOverlap(t *testing.T) {
	for _, tc := range []struct {
		name    string
		keys    []string
		overlap string
	}{
		{name: "distinct", keys: []string{"application/json", "application/*", "*/*", "application/json;charset=utf-8"}},
		{name: "case folded duplicate", keys: []string{"application/json", "Application/JSON"}, overlap: "application/json"},
		{name: "case folded wildcard duplicate", keys: []string{"image/*", "IMAGE/*"}, overlap: "image/*"},
		{name: "listed in another key", keys: []string{"a/b", "c/d, A/B"}, overlap: "a/b"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			content := make(Content, len(tc.keys))
			for _, k := range tc.keys {
				content[k] = NewMediaType()
			}
			err := content.Validate(t.Context())
			if tc.overlap == "" {
				require.NoError(t, err)
				return
			}
			var e *ContentKeysOverlapError
			require.ErrorAs(t, err, &e)
			require.Equal(t, tc.overlap, e.MediaRange)
			require.ElementsMatch(t, tc.keys, []string{e.Key, e.OtherKey})
			require.Equal(t, "content-keys-overlap", e.Code())
		})
	}
}
