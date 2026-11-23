package openapi2conv_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
)

func TestPathsExtensions(t *testing.T) {
	t.Parallel()
	for _, paths := range []string{`{"x-display":"catalog"}`, `{"/x-orders":{},"\u0078-display":"catalog"}`} {
		input := `{"swagger":"2.0","info":{"title":"Orders","version":"1"},"paths":` + paths + `}`
		var v2 openapi2.T
		require.NoError(t, json.Unmarshal([]byte(input), &v2))

		v3, err := openapi2conv.ToV3(&v2)
		require.NoError(t, err)
		require.NoError(t, v3.Validate(t.Context()))
		require.Equal(t, len(v2.Paths), v3.Paths.Len())
		require.Equal(t, "catalog", v3.Paths.Extensions["x-display"])

		back, err := openapi2conv.FromV3(v3)
		require.NoError(t, err)
		out, err := json.Marshal(back)
		require.NoError(t, err)
		require.JSONEq(t, input, string(out))
	}
}

func TestPathsExtensionNames(t *testing.T) {
	t.Parallel()
	v2 := openapi2.T{PathsExtensions: map[string]any{"x-ok": true, "X-other": true}}
	v3, err := openapi2conv.ToV3(&v2)
	require.NoError(t, err)
	require.Equal(t, map[string]any{"x-ok": true}, v3.Paths.Extensions)
	require.Contains(t, v2.PathsExtensions, "X-other")

	v3.Paths.Extensions["other"] = true
	back, err := openapi2conv.FromV3(v3)
	require.NoError(t, err)
	require.Equal(t, map[string]any{"x-ok": true}, back.PathsExtensions)
	require.Contains(t, v3.Paths.Extensions, "other")

	delete(v2.PathsExtensions, "x-ok")
	v3, err = openapi2conv.ToV3(&v2)
	require.NoError(t, err)
	require.Nil(t, v3.Paths)
}
