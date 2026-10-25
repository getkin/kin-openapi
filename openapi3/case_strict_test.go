package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// TestKeysOnlyMatchFieldsOfTheSameCase covers keys that differ from a known
// field only in case. A document decodes from its nodes, where a key matches a
// field by name and nothing else, so such a key is an extension rather than a
// value for the field it resembles. That holds whichever encoding/json the
// build has, since the json package decodes no document here.
func TestKeysOnlyMatchFieldsOfTheSameCase(t *testing.T) {
	t.Parallel()

	const spec = "Components:\n schemas:\n SChemAs:\n  0:"

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	require.Nil(t, doc.Components)
	require.Equal(t, map[string]any{
		"Components": map[string]any{
			"schemas": nil,
			"SChemAs": map[string]any{"0": nil},
		},
	}, doc.Extensions)

	once, err := doc.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, `{"openapi": "", "info": null, "paths": null, "Components": {"schemas": null, "SChemAs": {"0": null}}}`, string(once))

	loader = openapi3.NewLoader()
	reloaded, err := loader.LoadFromData(once)
	require.NoError(t, err)
	twice, err := reloaded.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, string(once), string(twice))
}
