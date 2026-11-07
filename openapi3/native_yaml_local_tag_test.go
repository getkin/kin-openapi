package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// A scalar carrying an application's own tag reads as the text it wrote, as it
// does on the json path, whatever field it sits in.
func TestNativeYAMLLocalTag(t *testing.T) {
	doc := []byte(`openapi: 3.0.0
info:
  title: t
  version: !v '1'
  description: !md hello
  x-n: !n 42
  x-m: !m {a: !b true}
  x-s: !s [1, x]
paths: {}
components:
  schemas:
    S:
      type: !0 string
`)
	loader := NewLoader()
	got, err := loader.LoadFromData(doc)
	require.NoError(t, err)

	info, err := got.Info.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, `{"title":"t","version":"1","description":"hello","x-n":"42","x-m":{"a":"true"},"x-s":[1,"x"]}`, string(info))
	require.True(t, got.Components.Schemas["S"].Value.Type.Is("string"))
}
