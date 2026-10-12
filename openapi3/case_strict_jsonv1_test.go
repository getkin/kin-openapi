//go:build !goexperiment.jsonv2

package openapi3_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// TestKeysOnlyMatchFieldsOfTheSameCase records how encoding/json v1, which
// ignores case:strict, matches keys to fields regardless of case: such keys
// fill the field and are also kept as extensions, so the document marshals
// both keys and does not load again.
func TestKeysOnlyMatchFieldsOfTheSameCase(t *testing.T) {
	const spec = "Components:\n schemas:\n SChemAs:\n  0:"

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	require.NotNil(t, doc.Components)
	require.Empty(t, doc.Components.Schemas)
	require.Equal(t, map[string]any{"SChemAs": map[string]any{"0": nil}}, doc.Components.Extensions)
	require.Equal(t, map[string]any{
		"Components": map[string]any{
			"schemas": nil,
			"SChemAs": map[string]any{"0": nil},
		},
	}, doc.Extensions)

	once, err := doc.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, `{"openapi": "", "info": null, "paths": null, "Components": {"schemas": null, "SChemAs": {"0": null}}, "components": {"SChemAs": {"0": null}}}`, string(once))

	loader = openapi3.NewLoader()
	_, err = loader.LoadFromData(once)
	require.EqualError(t, err, "invalid schema: value MUST be an object")
}

func TestFieldsIgnoreKeysOfAnotherCase(t *testing.T) {
	var info openapi3.Info
	err := json.Unmarshal([]byte(`{"title":"kept","Title":"other","Version":"1"}`), &info)
	require.NoError(t, err)
	require.Equal(t, "other", info.Title)
	require.Equal(t, "1", info.Version)
	require.Equal(t, map[string]any{"Title": "other", "Version": "1"}, info.Extensions)

	var schema openapi3.Schema
	err = json.Unmarshal([]byte(`{"Type":"string","MaxLength":1}`), &schema)
	require.NoError(t, err)
	require.Equal(t, &openapi3.Types{"string"}, schema.Type)
	require.Equal(t, uint64(1), *schema.MaxLength)
	require.Equal(t, map[string]any{"Type": "string", "MaxLength": float64(1)}, schema.Extensions)
}
