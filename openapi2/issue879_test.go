package openapi2_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/internal/yamlconv"
	"github.com/getkin/kin-openapi/openapi2"
)

func TestPathsExtensions(t *testing.T) {
	t.Parallel()
	input := []byte(`swagger: "2.0"
info: {title: Orders, version: "1"}
paths:
  /x-orders: {}
  x-display: catalog
  x-config: {labels: [orders, shipping]}
  x-null: null
`)
	var doc openapi2.T
	require.NoError(t, openapi2.UnmarshalFromData(input, &doc))
	require.Len(t, doc.Paths, 1)
	require.NotNil(t, doc.Paths["/x-orders"])

	want := map[string]any{
		"x-display": "catalog",
		"x-config":  map[string]any{"labels": []any{"orders", "shipping"}},
		"x-null":    nil,
	}
	require.Equal(t, want, doc.PathsExtensions)
	require.Empty(t, doc.Extensions)

	for _, marshal := range []func(any) ([]byte, error){json.Marshal, yamlconv.Marshal} {
		data, err := marshal(doc)
		require.NoError(t, err)
		var back openapi2.T
		require.NoError(t, openapi2.UnmarshalFromData(data, &back))
		require.Equal(t, doc, back)
	}
}

func TestEmptyPaths(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		input    string
		nilPaths bool
	}{
		{`{}`, true},
		{`{"paths":null}`, true},
		{`{"paths":{}}`, false},
	} {
		doc := openapi2.T{PathsExtensions: map[string]any{"x-previous": true}}
		require.NoError(t, json.Unmarshal([]byte(tc.input), &doc))
		require.Equal(t, tc.nilPaths, doc.Paths == nil)
		require.Nil(t, doc.PathsExtensions)
		out, err := json.Marshal(doc)
		require.NoError(t, err)
		require.JSONEq(t, `{"swagger":"","info":{"title":"","version":""}}`, string(out))
	}
}

func TestPathItemErrorsWithExtensions(t *testing.T) {
	t.Parallel()
	for _, input := range []string{
		`{"paths":{"x-display":"catalog","/orders":42}}`,
		`{"paths":{"x-display":"catalog","X-display":42}}`,
	} {
		var doc openapi2.T
		err := openapi2.UnmarshalFromData([]byte(input), &doc)
		require.ErrorContains(t, err, "PathItem")
	}
}
