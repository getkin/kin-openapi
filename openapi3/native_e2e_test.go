package openapi3

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	goyaml "go.yaml.in/yaml/v3"
)

// A complete document decoded through UnmarshalYAML must equal the same
// document decoded as JSON.
func TestNativeE2E_WholeDocument(t *testing.T) {
	// Every full document in testdata, so this is breadth rather than a
	// hand-picked sample.
	paths, err := filepath.Glob("testdata/*.y*ml")
	require.NoError(t, err)
	var ran int
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil || !bytes.HasPrefix(bytes.TrimSpace(data), []byte("openapi:")) {
			continue
		}
		ran++
		t.Run(path, func(t *testing.T) {
			// Reference: the JSON path, via a plain YAML->JSON conversion.
			var asAny any
			require.NoError(t, goyaml.Unmarshal(data, &asAny))
			jsonBytes, err := json.Marshal(asAny)
			require.NoError(t, err)
			var viaJSON T
			require.NoError(t, json.Unmarshal(jsonBytes, &viaJSON))

			// Native: straight from the node tree.
			var viaYAML T
			require.NoError(t, goyaml.Unmarshal(data, &viaYAML))

			want, err := json.Marshal(&viaJSON)
			require.NoError(t, err)
			got, err := json.Marshal(&viaYAML)
			require.NoError(t, err)
			require.JSONEq(t, string(want), string(got), "native decode should match the JSON path")
		})
	}
	require.Positive(t, ran, "should have found documents to compare")
}

// Path items and operations must carry an origin naming their own key.
func TestNativeE2E_OriginsReachOperations(t *testing.T) {
	defer func(v bool) { originEnabledVar = v }(originEnabledVar)
	originEnabledVar = true

	data, err := os.ReadFile("testdata/callbacks.yml")
	require.NoError(t, err)
	var doc T
	require.NoError(t, goyaml.Unmarshal(data, &doc))
	require.NotNil(t, doc.Paths)

	var checked int
	for path, pi := range doc.Paths.Map() {
		require.NotNil(t, pi.Origin, "path item %q has no origin", path)
		require.NotNil(t, pi.Origin.Key, "path item %q has no key origin", path)
		require.Equal(t, path, pi.Origin.Key.Name)
		require.Positive(t, pi.Origin.Key.Line)
		for method, op := range pi.Operations() {
			require.NotNil(t, op.Origin, "%s %s has no origin", method, path)
			require.NotNil(t, op.Origin.Key, "%s %s has no key origin", method, path)
			// Operations() reports the method uppercased; the origin names
			// the key as it appears in the document.
			require.Equal(t, strings.ToLower(method), op.Origin.Key.Name)
			checked++
		}
	}
	require.Positive(t, checked, "should have checked some operations")
}
