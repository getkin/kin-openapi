package openapi3_test

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// FuzzPathsFind builds a Paths object from arbitrary keys (one per line) and
// looks them up. InMatchingOrder is a permutation of the keys with
// fewer-variable paths first, and Find returns the item of every key it holds
// and never panics on any other.
func FuzzPathsFind(f *testing.F) {
	f.Add("/items\n/items/{id}\n/items/{id}/tags/{tag}\n/files/{name}.{ext}", "/items/{x}")
	f.Add("/a/{b}\n/a/{c}", "/a/{d}")
	f.Add("/{a*}\n/x", "/{b}")
	f.Add("{", "}")
	f.Add("", "")

	f.Fuzz(func(t *testing.T, keys, query string) {
		if len(keys) > 8<<10 || len(query) > 1<<10 {
			t.Skip()
		}
		paths := openapi3.NewPaths()
		items := map[string]*openapi3.PathItem{}
		for key := range strings.SplitSeq(keys, "\n") {
			item := &openapi3.PathItem{Summary: key}
			paths.Set(key, item)
			items[key] = item
		}

		ordered := paths.InMatchingOrder()
		require.Len(t, ordered, len(items))
		require.ElementsMatch(t, ordered, slices.Collect(maps.Keys(items)))
		require.True(t, slices.IsSortedFunc(ordered, func(a, b string) int {
			return strings.Count(a, "}") - strings.Count(b, "}")
		}), "%q", ordered)

		for key, item := range items {
			require.Same(t, item, paths.Find(key), "key %q", key)
		}
		_ = paths.Find(query)
	})
}
