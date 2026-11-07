package openapi3

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

// A raw U+0085 is a character to json and a line break to yaml, which folds it
// into a space. A json document that holds one keeps it, in a name as much as
// in a value.
func TestNativeE2E_JSONKeepsARawNextLine(t *testing.T) {
	const src = "{\"openapi\":\"3.0.0\",\"info\":{\"title\":\"t\",\"version\":\"1\"},\"paths\":{}," +
		"\"x-a\":\"b\u0085c\",\"x-\u0085\":1}"

	loader := NewLoader()
	doc, err := loader.LoadFromData([]byte(src))
	require.NoError(t, err)
	require.Equal(t, "b\u0085c", doc.Extensions["x-a"])
	require.Contains(t, doc.Extensions, "x-\u0085")

	// The document reads back as itself, which is what the fuzzers assert over
	// the corpus.
	out, err := json.Marshal(doc)
	require.NoError(t, err)
	require.JSONEq(t, src, string(out))
}

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
			require.NoError(t, yaml.Unmarshal(data, &asAny))
			jsonBytes, err := json.Marshal(asAny)
			require.NoError(t, err)
			var viaJSON T
			require.NoError(t, json.Unmarshal(jsonBytes, &viaJSON))

			// Native: straight from the node tree.
			var viaYAML T
			require.NoError(t, yaml.Unmarshal(data, &viaYAML))

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
	data, err := os.ReadFile("testdata/callbacks.yml")
	require.NoError(t, err)
	var doc T
	// Through unmarshal rather than the yaml package: origins are stamped only
	// for the tree a decode indexed.
	_, err = unmarshal(data, &doc, true, nil)
	require.NoError(t, err)
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
