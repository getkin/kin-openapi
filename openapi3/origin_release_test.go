package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// A decode releases its per-decode state when it ends. originEndsVar indexes
// the whole node tree, so leaving it set pins that tree until the next decode
// replaces it, and does so for every caller whether or not rememberOriginTree
// kept the tree. On a 25 MB spec that tree is most of what a load holds.
func TestUnmarshalReleasesPerDecodeState(t *testing.T) {
	IncludeOrigin = true
	defer func() { IncludeOrigin = false }()

	loader := NewLoader()
	_, err := loader.LoadFromData([]byte("openapi: 3.0.0\ninfo: {title: t, version: \"1\"}\npaths: {}\n"))
	require.NoError(t, err)

	require.Nil(t, originEndsVar, "the end index outlived its decode, pinning the node tree it indexes")
	require.Empty(t, originFileVar)
	require.False(t, originEnabledVar)
}
