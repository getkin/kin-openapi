package openapi3filter_test

import (
	"archive/zip"
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3filter"
)

// ZipFileBodyDecoder must return exactly the archived files' bytes: the read
// loop used to append the whole reused 256-byte buffer per read instead of
// the n bytes actually read, padding the result with NUL bytes and leaking
// bytes of previously read files into later ones. See #1246.
func TestIssue1246(t *testing.T) {
	buildZip := func(t *testing.T, files [][2]string) []byte {
		t.Helper()
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		for _, f := range files {
			w, err := zw.Create(f[0])
			require.NoError(t, err)
			_, err = w.Write([]byte(f[1]))
			require.NoError(t, err)
		}
		require.NoError(t, zw.Close())
		return buf.Bytes()
	}

	t.Run("file shorter than the read buffer is not padded", func(t *testing.T) {
		data := buildZip(t, [][2]string{
			{"a.txt", "hello world"},
		})

		got, err := openapi3filter.ZipFileBodyDecoder(bytes.NewReader(data), http.Header{}, nil, nil)
		require.NoError(t, err)
		require.Equal(t, "hello world", got)
	})

	t.Run("later file does not see bytes of an earlier one", func(t *testing.T) {
		first := strings.Repeat("A", 256)
		data := buildZip(t, [][2]string{
			{"first.txt", first},
			{"second.txt", "B"},
		})

		got, err := openapi3filter.ZipFileBodyDecoder(bytes.NewReader(data), http.Header{}, nil, nil)
		require.NoError(t, err)
		require.Equal(t, first+"B", got)
	})
}
