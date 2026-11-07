package openapi3_test

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func writeTitledSpec(t *testing.T, p, title string) {
	t.Helper()
	content := "openapi: 3.0.0\ninfo: {title: " + title + ", version: '1'}\npaths: {}\n"
	require.NoError(t, os.WriteFile(p, []byte(content), 0o600))
}

func loadTitle(t *testing.T, readFromURI openapi3.ReadFromURIFunc, p string) string {
	t.Helper()
	loader := openapi3.NewLoader()
	loader.ReadFromURIFunc = readFromURI
	doc, err := loader.LoadFromFile(p)
	require.NoError(t, err)
	return doc.Info.Title
}

// TestReadFromURIPerLoaderSeesEditedFile verifies that ReadFromURIPerLoader
// caches per Loader: a new Loader re-reads a file that changed since an
// earlier Loader read it.
func TestReadFromURIPerLoaderSeesEditedFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "spec.yaml")

	writeTitledSpec(t, p, "first")
	require.Equal(t, "first", loadTitle(t, openapi3.ReadFromURIPerLoader, p))

	writeTitledSpec(t, p, "second")
	require.Equal(t, "second", loadTitle(t, openapi3.ReadFromURIPerLoader, p))
}

// TestDefaultReadFromURIKeepsProcessCache documents that DefaultReadFromURI
// keeps serving the first read of an absolute file path to every Loader.
func TestDefaultReadFromURIKeepsProcessCache(t *testing.T) {
	p := filepath.Join(t.TempDir(), "spec.yaml")

	writeTitledSpec(t, p, "first")
	require.Equal(t, "first", loadTitle(t, openapi3.DefaultReadFromURI, p))

	writeTitledSpec(t, p, "second")
	require.Equal(t, "first", loadTitle(t, openapi3.DefaultReadFromURI, p))
}

// TestURIMapCachePerLoaderCachesWithinLoader verifies that repeated reads of
// the same URI through one Loader hit the reader only once.
func TestURIMapCachePerLoaderCachesWithinLoader(t *testing.T) {
	p := filepath.Join(t.TempDir(), "spec.yaml")
	writeTitledSpec(t, p, "first")

	calls := 0
	readFromURI := openapi3.URIMapCachePerLoader(func(loader *openapi3.Loader, location *url.URL) ([]byte, error) {
		calls++
		return openapi3.ReadFromFile(loader, location)
	})

	loader := openapi3.NewLoader()
	loader.ReadFromURIFunc = readFromURI
	_, err := loader.LoadFromFile(p)
	require.NoError(t, err)
	_, err = loader.LoadFromFile(p)
	require.NoError(t, err)
	require.Equal(t, 1, calls)

	loader = openapi3.NewLoader()
	loader.ReadFromURIFunc = readFromURI
	_, err = loader.LoadFromFile(p)
	require.NoError(t, err)
	require.Equal(t, 2, calls)
}
