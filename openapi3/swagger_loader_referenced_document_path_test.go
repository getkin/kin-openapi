package openapi3

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReferencedDocumentPath(t *testing.T) {
	httpURL, err := url.Parse("http://example.com/path/to/schemas/test1.yaml")
	require.NoError(t, err)

	fileURL, err := url.Parse("path/to/schemas/test1.yaml")
	require.NoError(t, err)

	refEmpty := ""
	refNoComponent := "moreschemas/test2.yaml"
	refWithComponent := "moreschemas/test2.yaml#/components/schemas/someobject"

	for _, test := range []struct {
		path          *url.URL
		ref, expected string
	}{
		{
			path:     httpURL,
			ref:      refEmpty,
			expected: "http://example.com/path/to/schemas/",
		},
		{
			path:     httpURL,
			ref:      refNoComponent,
			expected: "http://example.com/path/to/schemas/moreschemas/",
		},
		{
			path:     httpURL,
			ref:      refWithComponent,
			expected: "http://example.com/path/to/schemas/moreschemas/",
		},
		{
			path:     fileURL,
			ref:      refEmpty,
			expected: "path/to/schemas/",
		},
		{
			path:     fileURL,
			ref:      refNoComponent,
			expected: "path/to/schemas/moreschemas/",
		},
		{
			path:     fileURL,
			ref:      refWithComponent,
			expected: "path/to/schemas/moreschemas/",
		},
	} {
		result, err := referencedDocumentPath(test.path, test.ref)
		require.NoError(t, err)
		require.Equal(t, test.expected, result.String())
	}
}
