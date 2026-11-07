package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// An alias that leads back to a node it sits inside is refused, rather than
// expanded until the stack runs out, which no recover can catch.
func TestNativeYAMLAliasCycle(t *testing.T) {
	const head = "openapi: 3.1.0\ninfo: {title: t, version: '1'}\npaths: {}\n"

	for name, doc := range map[string]string{
		"property": head + "components:\n  schemas:\n    e: &a\n      properties:\n        p: *a\n",
		"items":    head + "components:\n  schemas:\n    e: &a\n      items: *a\n",
		"nested":   head + "components:\n  schemas:\n    e: &a\n      allOf:\n        - properties:\n            p:\n              items: *a\n",
		"mutual":   head + "components:\n  schemas:\n    a: &a\n      items: &b\n        not: *a\n    b: *b\n",
		"untyped":  "components:\n schemas:\n    e: &base\n      properties:\n       2: *base\n",
	} {
		t.Run(name, func(t *testing.T) {
			loader := NewLoader()
			_, err := loader.LoadFromData([]byte(doc))
			require.ErrorContains(t, err, "value contains itself")
		})
	}

	t.Run("shared", func(t *testing.T) {
		doc := head + "components:\n  schemas:\n    s: &s {type: string}\n    o:\n      properties:\n        a: *s\n        b: *s\n      items: *s\n"
		loader := NewLoader()
		got, err := loader.LoadFromData([]byte(doc))
		require.NoError(t, err)
		require.Equal(t, "string", got.Components.Schemas["o"].Value.Properties["b"].Value.Type.Slice()[0])
	})
}
