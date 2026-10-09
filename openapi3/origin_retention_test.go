package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// An ordinary spec never consults its origin tree, so the loader does not keep
// one. Origins are still attached to the document itself; what goes away is the
// second copy the loader used to hold for the lifetime of the loader, which on
// a large spec was a third of everything it retained.
func TestOriginTree_NotRetainedForAnOrdinarySpec(t *testing.T) {
	t.Parallel()

	loader := NewLoader()
	loader.IncludeOrigin = true
	loader.Context = t.Context()

	doc, err := loader.LoadFromFile("testdata/origin/simple.yaml")
	require.NoError(t, err)

	require.NotNil(t, doc.Origin, "origins are still attached to the document")
	require.Empty(t, loader.originTrees, "no tree is retained: nothing could read it")
}

// The tree is kept for exactly the documents that can consult it, and this OAD
// contains one of each: the entry document has only fields OpenAPI defines,
// while the file it $refs is a shared fragment whose top level is the schema
// name "User", which is what arrives untyped and what attachOriginToResolved
// walks the tree for.
//
// The retained tree still does its job here, so this pins the saving and the
// capability together: dropping the tree for the referenced file would leave
// the resolved schema without an origin.
func TestOriginTree_RetainedOnlyWhereItCanBeRead(t *testing.T) {
	t.Parallel()

	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.IncludeOrigin = true
	loader.Context = t.Context()

	doc, err := loader.LoadFromFile("testdata/origin/arbitrary_key.yaml")
	require.NoError(t, err)

	require.NotContains(t, loader.originTrees, doc,
		"the entry document has only fields OpenAPI defines, so no $ref reaches into it untyped")
	require.Len(t, loader.originTrees, 1, "only the shared-fragment file keeps its tree")
	for retained := range loader.originTrees {
		require.NotEmpty(t, retained.Extensions,
			"a tree is kept only for a document with top-level fields OpenAPI does not define")
	}

	// The capability the retained tree exists for: the resolved schema carries
	// the origin of its own file, not of the $ref site.
	schema := doc.Paths.Find("/users").Get.Responses.Value("200").Value.
		Content["application/json"].Schema.Value
	require.NotNil(t, schema.Origin, "the resolved schema keeps its origin")
	require.Contains(t, schema.Origin.Key.File, "arbitrary_key_schemas.yaml")
}

// Without IncludeOrigin there is no tree to begin with, so the new condition
// cannot change anything here.
func TestOriginTree_NotRetainedWhenOriginsAreOff(t *testing.T) {
	t.Parallel()

	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.Context = t.Context()

	doc, err := loader.LoadFromFile("testdata/origin/arbitrary_key.yaml")
	require.NoError(t, err)

	require.Nil(t, doc.Origin)
	require.Empty(t, loader.originTrees)
}
