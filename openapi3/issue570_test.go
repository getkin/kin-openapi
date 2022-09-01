package openapi3

import (
	// "net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

// func TestIssue570FromURIFIXMEReplaceMe(t *testing.T) {
// 	uri, err := url.Parse("https://rubrikinc.github.io/api-doc-internal-6.0/openapi.json")
// 	require.NoError(t, err)
// https://github.com/getkin/kin-openapi/pull/571
// 	loader := NewLoader()
// 	doc, err := loader.LoadFromURI(uri)
// 	require.NoError(t, err)
// 	err = doc.Validate(loader.Context)
// 	require.NoError(t, err)
// }

func TestIssue570TODOMinimizeMe(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromFile("testdata/https:,,rubrikinc.github.io,api-doc-internal-6.0,openapi.json")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
}
