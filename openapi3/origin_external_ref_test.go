package openapi3

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// A $ref to a schema stored under an arbitrary top-level key in another file
// (e.g. "./schemas.yaml#/User", the Swagger-2-era "definitions bag") must carry
// the origin of the file it lives in, like a $ref into /components/schemas does.
// The key is not a typed field of T, so the loader reaches it through T.Extensions
// (a generic map); the origin must survive that path.
func TestOrigin_ExternalRefToArbitraryTopLevelKey(t *testing.T) {
	loader := NewLoader()
	loader.IncludeOrigin = true
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromFile("testdata/origin/arbitrary_key.yaml")
	require.NoError(t, err)

	user := doc.Paths.Value("/users").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value
	require.NotNil(t, user, "the User schema resolves")

	require.NotNil(t, user.Origin, "a schema $ref'd under an arbitrary top-level key must carry an origin")
	require.NotNil(t, user.Origin.Key, "the origin has a key location")
	require.Equal(t, "arbitrary_key_schemas.yaml", filepath.Base(user.Origin.Key.File),
		"origin points at the file the schema lives in, not the referencing document")
	require.NotZero(t, user.Origin.Key.Line, "the origin has a source line")
}
