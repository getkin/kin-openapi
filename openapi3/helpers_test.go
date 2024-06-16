package openapi3

import (
	"net/url"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReferencesComponentInRootDocument(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true

	runAssertions := func(doc *T) {
		// The element type of ./records.yml references a document which is also in the root document.
		v, ok := ReferencesComponentInRootDocument(doc, doc.Components.Schemas["BookRecords"].Value.Items)
		assert.True(t, ok)
		assert.Equal(t, "#/components/schemas/BookRecord", v)

		// The array element type directly references the component in the root document.
		v, ok = ReferencesComponentInRootDocument(doc, doc.Components.Schemas["CdRecords"].Value.Items)
		assert.True(t, ok)
		assert.Equal(t, "#/components/schemas/CdRecord", v)

		// A component from the root document should
		v, ok = ReferencesComponentInRootDocument(doc, doc.Components.Schemas["CdRecord"])
		assert.True(t, ok)
		assert.Equal(t, "#/components/schemas/CdRecord", v)

		// The error response component is in the root doc.
		v, ok = ReferencesComponentInRootDocument(doc, doc.Paths.Find("/records").Get.Responses.Value("500"))
		assert.True(t, ok)
		assert.Equal(t, "#/components/responses/ErrorResponse", v)

		v, ok = ReferencesComponentInRootDocument(doc, doc.Paths.Find("/records").Get.Responses.Value("500").Value.Content.Get("application/json").Schema)
		assert.False(t, ok)
		assert.Empty(t, v)

		// Ref path doesn't include a './'
		v, ok = ReferencesComponentInRootDocument(doc, doc.Paths.Find("/record").Get.Parameters[0])
		assert.True(t, ok)
		assert.Equal(t, "#/components/parameters/BookIDParameter", v)

		v, ok = ReferencesComponentInRootDocument(doc, doc.Paths.Find("/record").Get.Responses.Value("200").Value.Content.Get("application/json").Examples["first-example"])
		assert.True(t, ok)
		assert.Equal(t, "#/components/examples/RecordResponseExample", v)

		// Matches equivalent paths where string is no equal.
		v, ok = ReferencesComponentInRootDocument(doc, doc.Paths.Find("/record").Get.Responses.Value("200").Value.Headers["X-Custom-Header"])
		assert.True(t, ok)
		assert.Equal(t, "#/components/headers/CustomHeader", v)

		// Same structure distinct definition of the same header
		v, ok = ReferencesComponentInRootDocument(doc, doc.Paths.Find("/record").Get.Responses.Value("200").Value.Headers["X-Custom-Header2"])
		assert.False(t, ok)
		assert.Empty(t, v)
	}

	// Load from the file system
	doc, err := loader.LoadFromFile("testdata/refsToRoot/openapi.yml")
	require.NoError(t, err)

	runAssertions(doc)

	// Loading from a URL by mocking HTTP calls.
	// Loads the data using the URI path from the testdata/ folder.
	loader.ReadFromURIFunc = func(loader *Loader, url *url.URL) ([]byte, error) {
		localURL := *url
		localURL.Scheme = ""
		localURL.Host = ""
		localURL.Path = filepath.Join("testdata", localURL.Path)

		return ReadFromFile(loader, &localURL)
	}

	u, _ := url.Parse("https://example.com/refsToRoot/openapi.yml")
	doc, err = loader.LoadFromURI(u)
	require.NoError(t, err)

	runAssertions(doc)
}
