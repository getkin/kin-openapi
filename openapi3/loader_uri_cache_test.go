package openapi3_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// TestNewLoaderSeesEditedFile verifies that the default URI reader cache is
// scoped to a Loader: a brand-new Loader re-reads files that changed since an
// earlier load on a different Loader. A process-wide cache would return the
// stale first version forever (see issue #1416).
func TestNewLoaderSeesEditedFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "spec.yaml")
	write := func(title string) {
		content := "openapi: 3.0.0\ninfo: {title: " + title + ", version: '1'}\npaths: {}\n"
		require.NoError(t, os.WriteFile(p, []byte(content), 0o600))
	}

	write("first")
	_, err := openapi3.NewLoader().LoadFromFile(p)
	require.NoError(t, err)

	write("second")
	doc, err := openapi3.NewLoader().LoadFromFile(p)
	require.NoError(t, err)
	require.Equal(t, "second", doc.Info.Title)
}
