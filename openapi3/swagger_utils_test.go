package openapi3_test

import (
	"net/http"

	"net/url"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestResettingExternalRefs(t *testing.T) {

	cs := startTestServer(http.Dir("testdata"))
	defer cs()

	loader := openapi3.NewSwaggerLoader(openapi3.WithAllowExternalRefs(true))
	remote, err := url.Parse("http://" + addr + "/test.refcache.openapi.yml")
	require.NoError(t, err)

	doc, err := loader.LoadSwaggerFromURI(remote)
	require.NoError(t, err)

	openapi3.ClearResolvedExternalRefs(doc)

	for _, s := range []string{"ref1", "ref2", "ref3", "ref4", "ref5", "ref6"} {
		sr := doc.Components.Schemas["AnotherTestSchema"].Value.Properties[s]
		require.True(t, sr.IsValue())
		require.False(t, sr.Resolved())
	}
}
