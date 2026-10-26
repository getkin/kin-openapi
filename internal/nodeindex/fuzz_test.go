package nodeindex_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	yaml "github.com/oasdiff/yaml3"

	"github.com/getkin/kin-openapi/internal/fuzzing"
	"github.com/getkin/kin-openapi/internal/nodeindex"
)

// A block ends no earlier than it starts, no later than the input does, and
// no earlier than any block it contains.
func FuzzNew(f *testing.F) {
	f.Add([]byte("a:\n  b: 1\n  c: |\n    x\n\n    y\n\nd: [1,\n  2]\n"))
	f.Add([]byte("- name: id\n  in: path\n- &a {x: 1}\n- *a\n- <<: *a\n  y: 2\n"))
	f.Add([]byte("a: >-\n  folded\n  # not a comment\nb:\n# comment\n  c: 1\n"))
	f.Add([]byte("a:\r\n  b: 1\r\n  c:\r\n    - 1\r\n"))
	f.Add([]byte("--- &r\nk: \"multi\n  line\"\n? complex\n: key\n"))
	f.Add([]byte("a:\n\tb\n"))
	for _, seed := range fuzzing.Testdata(f, "../../openapi3/testdata") {
		f.Add(seed)
	}
	for _, seed := range fuzzing.APIsGuru(f, "../../openapi3/testdata", "") {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 32<<10 {
			t.Skip()
		}
		var root yaml.Node
		if yaml.Unmarshal(data, &root) != nil {
			return
		}
		ix := nodeindex.New(&root, data)
		lines := bytes.Count(data, []byte("\n")) + 1

		var walk func(n *yaml.Node) int
		walk = func(n *yaml.Node) int {
			require.True(t, ix.Has(n))
			end, col := ix.EndOf(n)
			if end != 0 {
				require.LessOrEqual(t, end, lines)
				require.GreaterOrEqual(t, col, 1)
				if n.Line <= lines {
					require.GreaterOrEqual(t, end, n.Line, "node at %d:%d", n.Line, n.Column)
				}
			}
			for _, c := range n.Content {
				if childEnd := walk(c); end != 0 && childEnd != 0 {
					require.GreaterOrEqual(t, end, childEnd, "node at %d:%d ends before its child at %d:%d", n.Line, n.Column, c.Line, c.Column)
				}
			}
			return end
		}
		walk(&root)
	})
}
