package nodeindex_test

import (
	"strings"
	"testing"

	yaml "github.com/oasdiff/yaml3"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/internal/nodeindex"
)

func parse(t *testing.T, src string) *yaml.Node {
	t.Helper()

	var root yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(src), &root))
	return &root
}

// The end of a block is the last line of its deepest descendant, grown over
// the lines of a block scalar, which carry no node of their own.
func TestEndOf(t *testing.T) {
	t.Parallel()

	const src = "a:\n  b: 1\n  c: |\n    text\n    more\nd: 2\n"
	root := parse(t, src)
	ix := nodeindex.New(root, []byte(src))

	mapping := root.Content[0]
	aValue := mapping.Content[1]
	line, column := ix.EndOf(aValue)
	require.Equal(t, 5, line, lastLineOfTheBlockScalar)
	require.Equal(t, 9, column, pastTheLastCharacter)

	line, _ = ix.EndOf(mapping.Content[3])
	require.Equal(t, 6, line)
}

const (
	lastLineOfTheBlockScalar = "the block ends at the last line of its scalar"
	pastTheLastCharacter     = "the column is one past the last character"
	aTreeThisIndexNeverSaw   = "a node from another tree is not in this index"
)

// A node the index never saw gets nothing, which is how a decode tells its own
// tree from another one.
func TestHas(t *testing.T) {
	t.Parallel()

	const src = "a: 1\n"
	ix := nodeindex.New(parse(t, src), []byte(src))

	foreign := parse(t, src)
	require.False(t, ix.Has(foreign.Content[0]), aTreeThisIndexNeverSaw)
	require.False(t, ix.Has(nil))

	root := parse(t, src)
	own := nodeindex.New(root, []byte(src))
	require.True(t, own.Has(root.Content[0]))
}

// An alias reports the key its anchor was defined under, not where it is used.
func TestAnchorKey(t *testing.T) {
	t.Parallel()

	const src = "base: &anchor\n  x: 1\nuse: *anchor\n"
	root := parse(t, src)
	ix := nodeindex.New(root, []byte(src))

	mapping := root.Content[0]
	anchored := mapping.Content[1]
	key, ok := ix.AnchorKey(anchored)
	require.True(t, ok)
	require.Equal(t, 1, key.Line)

	_, ok = ix.AnchorKey(mapping.Content[0])
	require.False(t, ok)
}

// Every method is safe on a nil index, which is what a decode without origins
// carries.
func TestNilIndex(t *testing.T) {
	t.Parallel()

	var ix *nodeindex.Index
	line, column := ix.EndOf(parse(t, "a: 1\n"))
	require.Zero(t, line)
	require.Zero(t, column)
	require.False(t, ix.Has(parse(t, "a: 1\n")))
	_, ok := ix.AnchorKey(parse(t, "a: 1\n"))
	require.False(t, ok)
	require.Nil(t, nodeindex.New(nil, nil))
}

// A column counts characters, as the parser's own columns do, so a multi-byte
// character does not push the end of a block past the width of its line.
func TestEndOfCountsCharactersNotBytes(t *testing.T) {
	t.Parallel()

	for name, src := range map[string]string{
		"ascii":      "a: \"hello\"\nb: 2\n",
		"multi-byte": "a: \"h\u00e9llo\"\nb: 2\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			root := parse(t, src)
			ix := nodeindex.New(root, []byte(src))

			line, column := ix.EndOf(root.Content[0].Content[1])
			require.Equal(t, 1, line)
			require.Equal(t, 11, column, pastTheLastCharacter)
		})
	}
}

// A line break is not a character on the line it ends, so the same document
// ends its nodes in the same place with either ending. Git checks files out
// with CRLF endings on Windows.
func TestEndOfWithEitherLineEnding(t *testing.T) {
	t.Parallel()

	const src = "a:\n  c: |\n    text\n\n    more\nb: 2\n"
	for ending, text := range map[string]string{
		"lf":   src,
		"crlf": strings.ReplaceAll(src, "\n", "\r\n"),
	} {
		t.Run(ending, func(t *testing.T) {
			t.Parallel()

			root := parse(t, text)
			ix := nodeindex.New(root, []byte(text))

			line, column := ix.EndOf(root.Content[0].Content[1])
			require.Equal(t, 5, line, lastLineOfTheBlockScalar)
			require.Equal(t, 9, column, pastTheLastCharacter)
		})
	}
}
