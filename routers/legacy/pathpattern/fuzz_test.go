package pathpattern_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/routers/legacy/pathpattern"
)

// FuzzPathPattern adds arbitrary patterns (one per line) to a trie and
// matches an arbitrary path against it. Nothing panics, and a match returns
// one value per variable of the matched pattern.
func FuzzPathPattern(f *testing.F) {
	f.Add("GET /abc\n/abc/{param}\n/abc/{param*}\n/abc/{f|(.*)\\.jpeg}", "/abc/x.jpeg", true)
	f.Add("/a/{b}/c\n/a/{d}", "/a/1/c/", false)
	f.Add("/{x*}", "", true)
	f.Add("{", "}", true)

	f.Fuzz(func(t *testing.T, patterns, path string, regexp bool) {
		if len(patterns) > 4<<10 || len(path) > 1<<10 {
			t.Skip()
		}
		opts := &pathpattern.Options{SupportWildcard: true, SupportRegExp: regexp}
		root := &pathpattern.Node{}
		for _, pattern := range strings.Split(patterns, "\n") {
			_ = root.Add(pattern, pattern, opts)
		}
		_ = root.String()

		node, values := root.Match(path)
		if node == nil {
			return
		}
		require.NotNil(t, node.Value)
		require.Len(t, values, len(node.VariableNames), "matched %v", node.Value)
	})
}
