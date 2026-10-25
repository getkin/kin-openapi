//go:build goexperiment.jsonv2

package openapi3_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// TestKeysOnlyMatchFieldsOfTheSameCase covers keys that differ from a known
// field only in case: the json:",case:strict" tags make encoding/json keep
// them as extensions rather than fill the field.
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

func TestFieldsIgnoreKeysOfAnotherCase(t *testing.T) {
	t.Parallel()

	var info openapi3.Info
	err := json.Unmarshal([]byte(`{"title":"kept","Title":"other","Version":"1"}`), &info)
	require.NoError(t, err)
	require.Equal(t, "kept", info.Title)
	require.Empty(t, info.Version)
	require.Equal(t, map[string]any{"Title": "other", "Version": "1"}, info.Extensions)

	var schema openapi3.Schema
	err = json.Unmarshal([]byte(`{"Type":"string","MaxLength":1}`), &schema)
	require.NoError(t, err)
	require.Nil(t, schema.Type)
	require.Nil(t, schema.MaxLength)
	require.Equal(t, map[string]any{"Type": "string", "MaxLength": float64(1)}, schema.Extensions)
}
