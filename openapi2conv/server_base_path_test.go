package openapi2conv_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

// TestToV3BasePathEscaping covers the base path of a document with no host,
// which becomes a server URL of its own. The URL is read back as a URL, so a
// base path that is not already one has to be escaped on the way out, or the
// path that comes back differs from the path that went in.
func TestToV3BasePathEscaping(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		basePath string
		url      string
	}{
		{name: "plain", basePath: "/v1", url: "/v1"},
		{name: "percent escape", basePath: "0%00", url: "0%2500"},
		{name: "control character", basePath: "0\x00", url: "0%00"},
		{name: "space", basePath: "/a b", url: "/a%20b"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc2 := &openapi2.T{
				Swagger:  "2.0",
				Info:     openapi3.Info{Title: "t", Version: "1"},
				BasePath: tc.basePath,
			}

			doc3, err := openapi2conv.ToV3(doc2)
			require.NoError(t, err)
			require.Len(t, doc3.Servers, 1)
			require.Equal(t, tc.url, doc3.Servers[0].URL)

			// The base path has to come back as it was written, and the
			// conversion has to settle: the server URL of the second pass is
			// the server URL of the first.
			back, err := openapi2conv.FromV3(doc3)
			require.NoError(t, err)
			require.Equal(t, tc.basePath, back.BasePath)
			require.Empty(t, back.Host)

			again, err := openapi2conv.ToV3(back)
			require.NoError(t, err)
			require.Len(t, again.Servers, 1)
			require.Equal(t, tc.url, again.Servers[0].URL)
		})
	}
}

// TestFromV3UnparseableServer pins that a server URL which is not a URL is
// kept rather than dropped. basePath is the one field that can hold it as it
// was written, and dropping it left the document with no server at all.
func TestFromV3UnparseableServer(t *testing.T) {
	t.Parallel()

	doc3 := &openapi3.T{
		OpenAPI: "3.0.3",
		Info:    &openapi3.Info{Title: "t", Version: "1"},
		Servers: openapi3.Servers{{URL: "ht tp://example.com/v1"}},
		Paths:   openapi3.NewPaths(),
	}

	back, err := openapi2conv.FromV3(doc3)
	require.NoError(t, err)
	require.Equal(t, "ht tp://example.com/v1", back.BasePath)
	require.Empty(t, back.Host)
	require.Empty(t, back.Schemes)
}
