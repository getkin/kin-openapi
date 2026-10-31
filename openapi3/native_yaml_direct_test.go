package openapi3_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	yaml "go.yaml.in/yaml/v3"

	"github.com/getkin/kin-openapi/openapi3"
)

// Decoding these types with the yaml package keeps what the document said.
// Without UnmarshalYAML the decoder fills declared fields by struct tag, so an
// extension is dropped, since Extensions carries `yaml:"-"`, and a key that
// differs from a field only in case is dropped with it.
func TestYAMLDecodeKeepsExtensionsAndKeyCase(t *testing.T) {
	t.Parallel()

	const src = "title: t\nversion: \"1\"\nx-custom: 7\nTitle: other\n"

	var info openapi3.Info
	require.NoError(t, yaml.Unmarshal([]byte(src), &info))
	require.Equal(t, "t", info.Title)
	require.Equal(t, "1", info.Version)
	require.Equal(t, map[string]any{"x-custom": float64(7), "Title": "other"}, info.Extensions)

	// And it marshals back to what a JSON decode of the same document gives.
	out, err := json.Marshal(&info)
	require.NoError(t, err)
	require.JSONEq(t, `{"title":"t","version":"1","x-custom":7,"Title":"other"}`, string(out))
}
