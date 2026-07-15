package openapi3

import (
	"net/url"
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

// Re-attaching origins reuses the origin tree retained at load time: resolving
// the arbitrary-key $ref must not read any file a second time.
func TestOrigin_ExternalRefToArbitraryTopLevelKey_NoRereads(t *testing.T) {
	reads := map[string]int{}
	loader := NewLoader()
	loader.IncludeOrigin = true
	loader.ReadFromURIFunc = func(l *Loader, location *url.URL) ([]byte, error) {
		reads[location.String()]++
		return DefaultReadFromURI(l, location)
	}

	doc, err := loader.LoadFromFile("testdata/origin/arbitrary_key.yaml")
	require.NoError(t, err)

	user := doc.Paths.Value("/users").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value
	require.NotNil(t, user.Origin, "origins are attached")

	require.Len(t, reads, 2, "the root and the $ref'd file")
	for location, n := range reads {
		require.Equalf(t, 1, n, "%s must be read exactly once", location)
	}
}

// Keying the retained trees by document also serves documents loaded from
// memory (no location): a $ref to an arbitrary top-level key in the same
// document gets its origin attached too.
func TestOrigin_InternalRefToArbitraryTopLevelKey_FromData(t *testing.T) {
	const spec = `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /users:
    get:
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/User'
User:
  type: object
  properties:
    id: { type: string }
`
	loader := NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	user := doc.Paths.Value("/users").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value
	require.NotNil(t, user.Origin, "a same-document arbitrary-key $ref carries an origin")
	require.NotZero(t, user.Origin.Key.Line)
}
