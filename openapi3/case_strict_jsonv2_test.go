//go:build goexperiment.jsonv2

package openapi3_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

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
